package client

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const MetadataFilename = ".knowledge.json"

type Metadata struct {
	MetadataFileAbsPath string
	Metadata            map[string]FileMetadata `json:"metadata"` // Map of file paths to metadata
	// TODO (idea): add other fields like description here, so we can hierarchically build a dataset description? Challenge is pruning and merging.
}

type FileMetadata map[string]any

// loadAndMergeMetadata checks if the given directory contains a metadata file.
// If so, it reads it in and merges it with the previous level of metadata.
// Doing so, the parentMetadata is trimmed down to only the entries relevant to this directory.
func loadDirMetadata(dirPath string) (*Metadata, error) {
	metadataPath := filepath.Join(dirPath, MetadataFilename)
	metaAbsPath, err := filepath.Abs(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for %s: %w", metadataPath, err)
	}
	dirPath = filepath.Dir(metadataPath)
	if _, err := os.Stat(metadataPath); err != nil {
		return nil, nil
	}
	// Metadata file exists
	fileContent, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file %s: %w", metadataPath, err)
	}

	metadata := &Metadata{
		MetadataFileAbsPath: metaAbsPath,
	}
	if err := json.Unmarshal(fileContent, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata file %s: %w", metadataPath, err)
	}

	return metadata, nil
}

func findMetadata(path string, metadataStack []Metadata, globalMetadata map[string]string) (FileMetadata, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	metadata := make(map[string]any)

	for _, metadataEntry := range metadataStack {
		target := strings.TrimPrefix(strings.TrimPrefix(absPath, filepath.Dir(metadataEntry.MetadataFileAbsPath)), string(filepath.Separator))

		if m, ok := metadataEntry.Metadata[target]; ok {
			for k, v := range m {
				metadata[k] = v
			}
		}
	}

	for k, v := range globalMetadata {
		metadata[k] = v
	}

	return metadata, nil
}
