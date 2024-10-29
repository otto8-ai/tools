package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

type ClientDeleteDataset struct {
	Client
}

func (s *ClientDeleteDataset) Customize(cmd *cobra.Command) {
	cmd.Use = "delete-dataset <dataset-id>"
	cmd.Short = "Delete a dataset"
	cmd.Args = cobra.ExactArgs(1)
}

func (s *ClientDeleteDataset) Run(cmd *cobra.Command, args []string) error {
	c, err := s.getClient(cmd.Context())
	if err != nil {
		return err
	}

	datasetID := args[0]

	err = c.DeleteDataset(cmd.Context(), datasetID)
	if err != nil {
		return err
	}

	fmt.Printf("Deleted dataset %q\n", datasetID)
	return nil
}
