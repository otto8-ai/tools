package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/gptscript-ai/knowledge/pkg/datastore"
	"github.com/gptscript-ai/knowledge/pkg/index"
	"github.com/spf13/cobra"
)

type ClientEditDataset struct {
	Client
	ResetMetadata   bool              `usage:"reset metadata to default (empty)"`
	UpdateMetadata  map[string]string `usage:"update metadata key-value pairs (existing metadata will be updated/preserved)"`
	ReplaceMetadata map[string]string `usage:"replace metadata with key-value pairs (existing metadata will be removed)"`
}

func (s *ClientEditDataset) Customize(cmd *cobra.Command) {
	cmd.Use = "edit-dataset <dataset-id>"
	cmd.Short = "Edit an existing dataset"
	cmd.Args = cobra.ExactArgs(1)
	cmd.MarkFlagsMutuallyExclusive("reset-metadata", "update-metadata", "replace-metadata")
}

func (s *ClientEditDataset) Run(cmd *cobra.Command, args []string) error {
	c, err := s.getClient(cmd.Context())
	if err != nil {
		return err
	}

	datasetID := args[0]

	// Get current dataset
	dataset, err := c.GetDataset(cmd.Context(), datasetID)
	if err != nil {
		return fmt.Errorf("failed to get dataset: %w", err)
	}

	if dataset == nil {
		fmt.Printf("dataset not found: %q\n", datasetID)
		return fmt.Errorf("dataset not found: %s", datasetID)
	}

	updatedDataset := index.Dataset{
		ID: dataset.ID,
	}

	// Update Metadata - since flags are mutually exclusive, this should be either an empty map, or one of the update/replace maps
	metadata := map[string]any{}

	for k, v := range s.UpdateMetadata {
		metadata[k] = v
	}

	for k, v := range s.ReplaceMetadata {
		metadata[k] = v
	}

	updatedDataset.Metadata = metadata

	dataset, err = c.UpdateDataset(cmd.Context(), updatedDataset, &datastore.UpdateDatasetOpts{ReplaceMedata: s.ResetMetadata || len(s.ReplaceMetadata) > 0})
	if err != nil {
		return fmt.Errorf("failed to update dataset: %w", err)
	}

	dataset.Files = nil // Don't print files

	jsonOutput, err := json.Marshal(dataset)
	if err != nil {
		return fmt.Errorf("failed to marshal dataset: %w", err)
	}

	fmt.Println("Updated dataset:\n", string(jsonOutput))
	return nil
}
