package cmd

import (
	"fmt"

	"github.com/gptscript-ai/knowledge/pkg/index/types"
	"github.com/spf13/cobra"
)

type ClientCreateDataset struct {
	Client
	ErrOnExists bool `usage:"Return an error if the dataset already exists"`
}

func (s *ClientCreateDataset) Customize(cmd *cobra.Command) {
	cmd.Use = "create-dataset <dataset-id>"
	cmd.Short = "Create a new dataset"
	cmd.Args = cobra.ExactArgs(1)
}

func (s *ClientCreateDataset) Run(cmd *cobra.Command, args []string) error {
	c, err := s.getClient(cmd.Context())
	if err != nil {
		return err
	}
	defer c.Close()

	datasetID := args[0]

	ds, err := c.CreateDataset(cmd.Context(), datasetID, &types.DatasetCreateOpts{ErrOnExists: s.ErrOnExists})
	if err != nil {
		return err
	}

	fmt.Printf("Created dataset %q\n", ds.ID)
	return nil
}
