package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

type ClientListDatasets struct {
	Client
	Archive string `usage:"Path to the archive file"`
}

func (s *ClientListDatasets) Customize(cmd *cobra.Command) {
	cmd.Use = "list-datasets"
	cmd.Short = "List existing datasets"
	cmd.Args = cobra.NoArgs
}

func (s *ClientListDatasets) Run(cmd *cobra.Command, args []string) error {
	c, err := s.getClient(cmd.Context())
	if err != nil {
		return err
	}

	ds, err := c.ListDatasets(cmd.Context())
	if err != nil {
		return fmt.Errorf("failed to list datasets: %w", err)
	}

	if len(ds) == 0 {
		fmt.Println("no datasets found")
		return nil
	}

	jsonOutput, err := json.Marshal(ds)
	if err != nil {
		return fmt.Errorf("failed to marshal datasets: %w", err)
	}

	fmt.Println(string(jsonOutput))
	return nil
}
