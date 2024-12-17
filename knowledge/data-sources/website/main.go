package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/gptscript-ai/go-gptscript"
	"github.com/sirupsen/logrus"
)

type MetadataInput struct {
	WebsiteCrawlingConfig WebsiteCrawlingConfig `json:"websiteCrawlingConfig"`
	Limit                 int                   `json:"limit"`
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
	Folders map[string]struct{} `json:"folders"`
}

type FileDetails struct {
	FilePath    string `json:"filePath,omitempty"`
	URL         string `json:"url,omitempty"`
	UpdatedAt   string `json:"updatedAt,omitempty"`
	Checksum    string `json:"checksum,omitempty"`
	SizeInBytes int64  `json:"sizeInBytes,omitempty"`
}

func main() {
	logOut := logrus.New()
	logOut.SetOutput(os.Stdout)
	logOut.SetFormatter(&logrus.JSONFormatter{})
	logErr := logrus.New()
	logErr.SetOutput(os.Stderr)

	ctx := context.Background()
	gptscriptClient, err := gptscript.NewGPTScript()
	if err != nil {
		logOut.WithError(fmt.Errorf("failed to create gptscript client, error: %v", err)).Error()
		os.Exit(0)
	}

	inputData := os.Getenv("GPTSCRIPT_INPUT")
	input := MetadataInput{}
	if err := json.Unmarshal([]byte(inputData), &input); err != nil {
		logOut.WithError(fmt.Errorf("failed to unmarshal input data, error: %v", err)).Error()
		os.Exit(0)
	}

	if input.Limit == 0 {
		input.Limit = getFromEnvOrDefault("OBOT_WEBSCRAPER_LIMIT", 250)
	}

	output := MetadataOutput{}

	var notfoundErr *gptscript.NotFoundInWorkspaceError
	outputData, err := gptscriptClient.ReadFileInWorkspace(ctx, ".metadata.json")
	if err != nil && !errors.As(err, &notfoundErr) {
		logOut.WithError(fmt.Errorf("failed to read .metadata.json in workspace, error: %w", err)).Error()
		os.Exit(0)
	} else if err == nil {
		if err := json.Unmarshal(outputData, &output); err != nil {
			logOut.WithError(fmt.Errorf("failed to unmarshal output data, error: %w", err)).Error()
			os.Exit(0)
		}
	}

	if output.Files == nil {
		output.Files = make(map[string]FileDetails)
	}

	if err := crawlColly(ctx, &input, &output, logErr, gptscriptClient); err != nil {
		logOut.WithError(fmt.Errorf("failed to crawl website: error: %w", err)).Error()
		os.Exit(0)
	}
}

func getFromEnvOrDefault(env string, def int) int {
	v, _ := strconv.Atoi(os.Getenv(env))
	if v != 0 {
		return v
	}
	return def
}

func writeMetadata(ctx context.Context, output *MetadataOutput, gptscript *gptscript.GPTScript) error {
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}
	return gptscript.WriteFileInWorkspace(ctx, ".metadata.json", data)
}
