package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/gptscript-ai/go-gptscript"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	drives2 "github.com/microsoftgraph/msgraph-sdk-go/drives"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/shares"
	"github.com/sirupsen/logrus"
)

type MetadataInput struct {
	OneDriveConfig *OneDriveConfig `json:"onedriveConfig,omitempty"`
}

type OneDriveConfig struct {
	SharedLinks []string `json:"sharedLinks"`
}

type MetadataOutput struct {
	Status string                 `json:"status"`
	Files  map[string]FileDetails `json:"files"`
	State  State                  `json:"state"`
}

type State struct {
	OneDriveState *OneDriveLinksConnectorState `json:"onedriveState,omitempty"`
}

type OneDriveLinksConnectorState struct {
	Files map[string]FileState `json:"files,omitempty"`
	Links map[string]LinkState `json:"links,omitempty"`
}

type LinkState struct {
	IsFolder bool   `json:"isFolder"`
	Name     string `json:"name"`
}

type FileState struct {
	FolderPath string `json:"folderPath"`
	FileName   string `json:"fileName"`
	URL        string `json:"url"`
}

type StaticTokenCredential struct {
	token string
}

func NewStaticTokenCredential(token string) StaticTokenCredential {
	return StaticTokenCredential{
		token: token,
	}
}

func (s StaticTokenCredential) GetToken(ctx context.Context, options policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return azcore.AccessToken{
		Token: s.token,
	}, nil
}

type FileDetails struct {
	FilePath    string `json:"filePath"`
	URL         string `json:"url"`
	SizeInBytes int64  `json:"sizeInBytes"`
	UpdatedAt   string `json:"updatedAt"`
}

func main() {
	logOut := logrus.New()
	logOut.SetOutput(os.Stdout)
	logOut.SetFormatter(&logrus.JSONFormatter{})
	logErr := logrus.New()
	logErr.SetOutput(os.Stderr)

	cred := NewStaticTokenCredential(os.Getenv("GPTSCRIPT_GRAPH_MICROSOFT_COM_BEARER_TOKEN"))
	client, err := msgraphsdk.NewGraphServiceClientWithCredentials(cred, []string{})
	if err != nil {
		logOut.WithError(fmt.Errorf("failed to create ms graph client, error: %w", err)).Error()
		os.Exit(0)
	}

	ctx := context.Background()
	gptscriptClient, err := gptscript.NewGPTScript()
	if err != nil {
		logOut.WithError(fmt.Errorf("failed to create gptscript client, error: %w", err)).Error()
		os.Exit(0)
	}

	inputData := os.Getenv("GPTSCRIPT_INPUT")
	input := MetadataInput{}

	if err := json.Unmarshal([]byte(inputData), &input); err != nil {
		logOut.WithError(fmt.Errorf("failed to unmarshal input data, error: %w", err)).Error()
		os.Exit(0)
	}
	if input.OneDriveConfig == nil {
		input.OneDriveConfig = &OneDriveConfig{}
	}

	output := MetadataOutput{}

	var notfoundErr *gptscript.NotFoundInWorkspaceError
	outputData, err := gptscriptClient.ReadFileInWorkspace(ctx, ".metadata.json")
	if err != nil && !errors.As(err, &notfoundErr) {
		logOut.WithError(fmt.Errorf("failed to read .metadata.json in workspace, error: %w", err)).Error()
		os.Exit(0)
	} else if err == nil {
		if err := json.Unmarshal(outputData, &output); err != nil {
			logrus.WithError(err).Fatal("Failed to unmarshal output data")
		}
	}

	if output.Files == nil {
		output.Files = make(map[string]FileDetails)
	}

	if output.State.OneDriveState == nil {
		output.State.OneDriveState = &OneDriveLinksConnectorState{
			Files: make(map[string]FileState),
		}
	}
	if output.State.OneDriveState.Files == nil {
		output.State.OneDriveState.Files = make(map[string]FileState)
	}
	if output.State.OneDriveState.Links == nil {
		output.State.OneDriveState.Links = make(map[string]LinkState)
	}

	for i := range input.OneDriveConfig.SharedLinks {
		input.OneDriveConfig.SharedLinks[i] = strings.TrimSpace(input.OneDriveConfig.SharedLinks[i])
	}

	if err := sync(ctx, logErr, input, &output, client, gptscriptClient); err != nil {
		logOut.WithError(fmt.Errorf("failed to sync onedrive links, error: %w", err)).Error()
		os.Exit(0)
	}

	output.Status = ""
	if err := writeMetadata(ctx, &output, gptscriptClient); err != nil {
		logOut.WithError(fmt.Errorf("failed to write metadata, error: %w", err)).Error()
		os.Exit(0)
	}
}

func sync(ctx context.Context, logErr *logrus.Logger, input MetadataInput, output *MetadataOutput, client *msgraphsdk.GraphServiceClient, gptscript *gptscript.GPTScript) error {
	items := map[string]models.DriveItemable{}
	for _, link := range input.OneDriveConfig.SharedLinks {
		requestParameters := &shares.ItemDriveItemRequestBuilderGetQueryParameters{
			Expand: []string{"children"},
		}
		configuration := &shares.ItemDriveItemRequestBuilderGetRequestConfiguration{
			QueryParameters: requestParameters,
		}
		shareDriveItem, err := client.Shares().BySharedDriveItemId(encodeURL(link)).DriveItem().Get(ctx, configuration)
		if err != nil {
			return err
		}
		root := path.Dir(getFullName(shareDriveItem))
		output.State.OneDriveState.Links[link] = LinkState{
			IsFolder: shareDriveItem.GetFile() == nil,
			Name:     *shareDriveItem.GetName(),
		}

		children, err := syncChildrenFileForItem(ctx, client, gptscript, shareDriveItem, output, root, logErr)
		if err != nil {
			return err
		}
		for _, child := range children {
			items[*child.GetId()] = child
		}
	}

	for id := range output.Files {
		if _, ok := items[id]; !ok {
			if output.Files[id].FilePath != "" {
				logErr.Infof("Deleting %s", output.Files[id].FilePath)
				if err := gptscript.DeleteFileInWorkspace(ctx, output.Files[id].FilePath); err != nil {
					return err
				}
			}
			delete(output.Files, id)
		}
	}

	return nil
}

func writeMetadata(ctx context.Context, output *MetadataOutput, gptscript *gptscript.GPTScript) error {
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}
	return gptscript.WriteFileInWorkspace(ctx, ".metadata.json", data)
}

func syncChildrenFileForItem(ctx context.Context, client *msgraphsdk.GraphServiceClient, gptscriptClient *gptscript.GPTScript, item models.DriveItemable, output *MetadataOutput, root string, logErr *logrus.Logger) ([]models.DriveItemable, error) {
	if item.GetFile() != nil {
		// We only sync item that is less than 50 MB, as most of the bigger files won't be supported from knowledge
		if item.GetSize() != nil && *item.GetSize() >= 1024*1024*50 {
			return nil, nil
		}
		if err := saveToMetadata(ctx, logErr, output, client, gptscriptClient, item, root); err != nil {
			return nil, err
		}
		return []models.DriveItemable{item}, nil
	}

	var result []models.DriveItemable
	for _, child := range item.GetChildren() {
		item, err := client.Drives().ByDriveId(*child.GetParentReference().GetDriveId()).Items().ByDriveItemId(*child.GetId()).Get(ctx, &drives2.ItemItemsDriveItemItemRequestBuilderGetRequestConfiguration{
			QueryParameters: &drives2.ItemItemsDriveItemItemRequestBuilderGetQueryParameters{
				Expand: []string{"children"},
			},
		})
		if err != nil {
			return nil, err
		}
		children, err := syncChildrenFileForItem(ctx, client, gptscriptClient, item, output, root, logErr)
		if err != nil {
			return nil, err
		}
		result = append(result, children...)
	}
	return result, nil
}

func saveToMetadata(ctx context.Context, logErr *logrus.Logger, output *MetadataOutput, client *msgraphsdk.GraphServiceClient, gptscriptClient *gptscript.GPTScript, item models.DriveItemable, root string) error {
	folders := make(map[string]struct{})
	files := make(map[string]FileState)
	fullPath := getFullName(item)
	relativePath := strings.TrimPrefix(fullPath, root)
	topRootFolder := strings.Split(strings.TrimPrefix(relativePath, string(os.PathSeparator)), string(os.PathSeparator))[0]
	created := false
	detail, ok := output.Files[*item.GetId()]
	if !ok {
		created = true
		detail.FilePath = relativePath
		detail.URL = *item.GetWebUrl()
		detail.UpdatedAt = (*item.GetLastModifiedDateTime()).String()
		detail.SizeInBytes = *item.GetSize()
		output.Files[*item.GetId()] = detail
	}
	files[*item.GetId()] = FileState{
		FolderPath: strings.TrimPrefix(filepath.Dir(relativePath), string(os.PathSeparator)),
		FileName:   path.Base(relativePath),
		URL:        *item.GetWebUrl(),
	}
	if created || detail.UpdatedAt != item.GetLastModifiedDateTime().String() {
		driveID := *item.GetParentReference().GetDriveId()
		data, err := client.Drives().ByDriveId(driveID).Items().ByDriveItemId(*item.GetId()).Content().Get(ctx, nil)
		if err != nil {
			return err
		}

		if err := gptscriptClient.WriteFileInWorkspace(ctx, relativePath, data); err != nil {
			return err
		}
		logErr.Infof("Downloaded %s", relativePath)
		detail.UpdatedAt = item.GetLastModifiedDateTime().String()
		detail.URL = *item.GetWebUrl()
		detail.FilePath = relativePath
		output.Files[*item.GetId()] = detail
	} else {
		logErr.Infof("Skipping %s because it is not changed", relativePath)
	}
	folders[topRootFolder] = struct{}{}
	output.State.OneDriveState.Files = files
	output.Status = fmt.Sprintf("Syncing file %v", relativePath)
	if err := writeMetadata(ctx, output, gptscriptClient); err != nil {
		return err
	}

	output.State.OneDriveState.Files = files
	return nil
}

func getFullName(item models.DriveItemable) string {
	p := item.GetParentReference().GetPath()
	if p != nil {
		_, after, found := strings.Cut(*p, ":")
		if found {
			return path.Join(after, *item.GetName())
		}
	}
	return ""
}

func encodeURL(u string) string {
	base64Value := base64.StdEncoding.EncodeToString([]byte(u))

	encodedUrl := "u!" + strings.TrimRight(base64Value, "=")
	encodedUrl = strings.ReplaceAll(encodedUrl, "/", "_")
	encodedUrl = strings.ReplaceAll(encodedUrl, "+", "-")
	return encodedUrl
}
