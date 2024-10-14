package main

import (
	"encoding/json"
	"os"
	"path"

	"github.com/sirupsen/logrus"
)

type Metadata struct {
	Input     MetadataInput  `json:"input"`
	Output    MetadataOutput `json:"output"`
	OutputDir string         `json:"outputDir"`
}

type MetadataInput struct {
	WebsiteCrawlingConfig WebsiteCrawlingConfig `json:"websiteCrawlingConfig"`
	Exclude               []string              `json:"exclude"`
}

type WebsiteCrawlingConfig struct {
	URLs []string `json:"urls"`
}

type MetadataOutput struct {
	Status string                 `json:"status"`
	Error  string                 `json:"error"`
	State  State                  `json:"state"`
	Files  map[string]FileDetails `json:"files"`
}

type State struct {
	WebsiteCrawlingState WebsiteCrawlingState `json:"websiteCrawlingState"`
}

type WebsiteCrawlingState struct {
	ScrapeJobIds map[string]string   `json:"scrapeJobIds"`
	Folders      map[string]struct{} `json:"folders"`
	Pages        map[string]struct{} `json:"pages"`
}

type FileDetails struct {
	FilePath  string `json:"filePath,omitempty"`
	URL       string `json:"url,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
	Checksum  string `json:"checksum,omitempty"`
}

func main() {
	var err error
	workingDir := os.Getenv("GPTSCRIPT_WORKSPACE_DIR")
	if workingDir == "" {
		workingDir, err = os.Getwd()
		if err != nil {
			logrus.Error(err)
			os.Exit(1)
		}
	}

	mode := os.Getenv("MODE")
	if mode == "" {
		mode = "colly"
	}

	metadata := Metadata{}
	metadataPath := path.Join(workingDir, ".metadata.json")
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		logrus.Fatal("metadata.json not found")
	}
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		logrus.Fatal(err)
	}

	err = json.Unmarshal(data, &metadata)
	if err != nil {
		logrus.Fatal(err)
	}

	if metadata.Output.Files == nil {
		metadata.Output.Files = make(map[string]FileDetails)
	}

	if metadata.OutputDir != "" {
		workingDir = metadata.OutputDir
	}

	if err := os.MkdirAll(workingDir, 0755); err != nil {
		logrus.Fatal(err)
	}

	if mode == "colly" {
		NewColly().Crawl(&metadata, metadataPath, workingDir)
	} else if mode == "firecrawl" {
		NewFirecrawl().Crawl(&metadata, metadataPath, workingDir)
	}
}

func writeMetadata(metadata *Metadata, path string) error {
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
