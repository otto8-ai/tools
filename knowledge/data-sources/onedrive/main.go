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
	FilePath  string `json:"filePath"`
	URL       string `json:"url"`
	UpdatedAt string `json:"updatedAt"`
}

func main() {
	logOut := logrus.New()
	logOut.SetOutput(os.Stdout)
	logOut.SetFormatter(&logrus.JSONFormatter{})
	logErr := logrus.New()
	logErr.SetOutput(os.Stderr)
	logErr.SetFormatter(&logrus.JSONFormatter{})

	cred := NewStaticTokenCredential(os.Getenv("GPTSCRIPT_GRAPH_MICROSOFT_COM_BEARER_TOKEN"))
	client, err := msgraphsdk.NewGraphServiceClientWithCredentials(cred, []string{})
	if err != nil {
		logErr.WithError(err).Fatal("Failed to create ms graph client")
	}

	ctx := context.Background()
	gptscriptClient, err := gptscript.NewGPTScript()
	if err != nil {
		logErr.WithError(err).Fatal("Failed to create gptscript client")
	}

	inputData := os.Getenv("GPTSCRIPT_INPUT")
	input := MetadataInput{}

	if err := json.Unmarshal([]byte(inputData), &input); err != nil {
		logErr.WithError(err).Fatal("Failed to unmarshal input data")
	}
	if input.OneDriveConfig == nil {
		input.OneDriveConfig = &OneDriveConfig{}
	}

	output := MetadataOutput{}

	var notfoundErr gptscript.NotFoundInWorkspaceError
	outputData, err := gptscriptClient.ReadFileInWorkspace(ctx, ".metadata.json")
	if err != nil && !errors.As(err, &notfoundErr) {
		logrus.WithError(err).Fatal("Failed to read .metadata.json in workspace")
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
		logrus.WithError(err).Fatal("Failed to sync onedrive links")
	}

	output.Status = ""
	if err := writeMetadata(ctx, &output, gptscriptClient); err != nil {
		logrus.Error(err)
		os.Exit(0)
	}
}

func sync(ctx context.Context, logErr *logrus.Logger, input MetadataInput, output *MetadataOutput, client *msgraphsdk.GraphServiceClient, gptscript *gptscript.GPTScript) error {
	items := make(map[string]struct {
		Item models.DriveItemable
		Root string
	})
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

		children, err := getChildrenFileForItem(ctx, client, shareDriveItem)
		if err != nil {
			return err
		}
		for _, child := range children {
			items[*child.GetId()] = struct {
				Item models.DriveItemable
				Root string
			}{
				Item: child,
				Root: root,
			}
		}
	}
	if err := saveToMetadata(ctx, logErr, output, client, gptscript, items); err != nil {
		return err
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

func getChildrenFileForItem(ctx context.Context, client *msgraphsdk.GraphServiceClient, item models.DriveItemable) ([]models.DriveItemable, error) {
	if item.GetFile() != nil {
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
		children, err := getChildrenFileForItem(ctx, client, item)
		if err != nil {
			return nil, err
		}
		result = append(result, children...)
	}
	return result, nil
}

func saveToMetadata(ctx context.Context, logErr *logrus.Logger, output *MetadataOutput, client *msgraphsdk.GraphServiceClient, gptscriptClient *gptscript.GPTScript, items map[string]struct {
	Item models.DriveItemable
	Root string
}) error {
	folders := make(map[string]struct{})
	files := make(map[string]FileState)
	for _, item := range items {
		fullPath := getFullName(item.Item)
		relativePath := strings.TrimPrefix(fullPath, item.Root)
		topRootFolder := strings.Split(strings.TrimPrefix(relativePath, string(os.PathSeparator)), string(os.PathSeparator))[0]
		created := false
		detail, ok := output.Files[*item.Item.GetId()]
		if !ok {
			created = true
			detail.FilePath = relativePath
			detail.URL = *item.Item.GetWebUrl()
			detail.UpdatedAt = (*item.Item.GetLastModifiedDateTime()).String()
			output.Files[*item.Item.GetId()] = detail
		}
		files[*item.Item.GetId()] = FileState{
			FolderPath: strings.TrimPrefix(filepath.Dir(relativePath), string(os.PathSeparator)),
			FileName:   path.Base(relativePath),
			URL:        *item.Item.GetWebUrl(),
		}
		if created || detail.UpdatedAt != item.Item.GetLastModifiedDateTime().String() {
			driveID := *item.Item.GetParentReference().GetDriveId()
			data, err := client.Drives().ByDriveId(driveID).Items().ByDriveItemId(*item.Item.GetId()).Content().Get(ctx, nil)
			if err != nil {
				return err
			}

			if err := gptscriptClient.WriteFileInWorkspace(ctx, relativePath, data); err != nil {
				return err
			}
			logErr.Infof("Downloaded %s", relativePath)
			detail.UpdatedAt = item.Item.GetLastModifiedDateTime().String()
			output.Files[*item.Item.GetId()] = detail
		} else {
			logErr.Infof("Skipping %s because it is not changed", relativePath)
		}
		folders[topRootFolder] = struct{}{}
		output.State.OneDriveState.Files = files
		output.Status = fmt.Sprintf("Synced %d files out of %d", len(output.Files), len(items))
		if err := writeMetadata(ctx, output, gptscriptClient); err != nil {
			return err
		}
	}
	for id := range output.Files {
		found := false
		if _, ok := items[id]; ok {
			found = true
		}
		if !found {
			if output.Files[id].FilePath != "" {
				logErr.Infof("Deleting %s", output.Files[id].FilePath)
				if err := gptscriptClient.DeleteFileInWorkspace(ctx, output.Files[id].FilePath); err != nil {
					return err
				}
			}
			delete(output.Files, id)
		}
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
