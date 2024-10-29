package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"

	"github.com/gptscript-ai/go-gptscript"
	"github.com/sirupsen/logrus"
)

type MetadataInput struct {
	WebsiteCrawlingConfig WebsiteCrawlingConfig `json:"websiteCrawlingConfig"`
}

type WebsiteCrawlingConfig struct {
	URLs []string `json:"urls"`
}

type MetadataOutput struct {
	Status string                 `json:"status"`
	State  State                  `json:"state"`
	Files  map[string]FileDetails `json:"files"`
}

type State struct {
	WebsiteCrawlingState WebsiteCrawlingState `json:"websiteCrawlingState"`
}

type WebsiteCrawlingState struct {
	Folders map[string]struct{}    `json:"folders"`
	Pages   map[string]PageDetails `json:"pages"`
}

type PageDetails struct {
	ParentURL string `json:"parentURL"`
}

type FileDetails struct {
	FilePath  string `json:"filePath,omitempty"`
	URL       string `json:"url,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
	Checksum  string `json:"checksum,omitempty"`
}

func main() {
	logOut := logrus.New()
	logOut.SetOutput(os.Stdout)
	logOut.SetFormatter(&logrus.JSONFormatter{})
	logErr := logrus.New()
	logErr.SetOutput(os.Stderr)
	logErr.SetFormatter(&logrus.JSONFormatter{})

	ctx := context.Background()
	gptscriptClient, err := gptscript.NewGPTScript()
	if err != nil {
		logErr.WithError(err).Fatal("Failed to create gptscript client")
	}

	inputData := os.Getenv("GPTSCRIPT_INPUT")
	input := MetadataInput{}
	for i := range input.WebsiteCrawlingConfig.URLs {
		input.WebsiteCrawlingConfig.URLs[i] = strings.TrimSpace(input.WebsiteCrawlingConfig.URLs[i])
	}

	if err := json.Unmarshal([]byte(inputData), &input); err != nil {
		logErr.WithError(err).Fatal("Failed to unmarshal input data")
	}

	output := MetadataOutput{}

	var notfoundErr *gptscript.NotFoundInWorkspaceError
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

	if output.State.WebsiteCrawlingState.Pages == nil {
		output.State.WebsiteCrawlingState.Pages = make(map[string]PageDetails)
	}

	mode := os.Getenv("MODE")
	if mode == "" {
		mode = "colly"
	}

	CrawlColly(ctx, &input, &output, logErr, gptscriptClient)
}

func writeMetadata(ctx context.Context, output *MetadataOutput, gptscript *gptscript.GPTScript) error {
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}
	return gptscript.WriteFileInWorkspace(ctx, ".metadata.json", data)
}
