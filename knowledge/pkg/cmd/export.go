package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

type ClientExportDatasets struct {
	Client
	Output string `usage:"Output path" default:"."`
	All    bool   `usage:"Export all datasets" short:"a"`
}

func (s *ClientExportDatasets) Customize(cmd *cobra.Command) {
	cmd.Use = "export <dataset-id> [<dataset-id>...]"
	cmd.Short = "Export one or more datasets as an archive (zip)"
}

func (s *ClientExportDatasets) Run(cmd *cobra.Command, args []string) error {
	c, err := s.getClient(cmd.Context())
	if err != nil {
		return err
	}

	if s.All && len(args) > 0 {
		return fmt.Errorf("cannot use --all with dataset IDs")
	}

	dsnames := args

	if s.All {
		lds, err := c.ListDatasets(cmd.Context())
		if err != nil {
			return err
		}

		dsnames = make([]string, len(lds))
		for i, ds := range lds {
			dsnames[i] = ds.ID
		}
	} else {
		for _, datasetID := range dsnames {
			ds, err := c.GetDataset(cmd.Context(), datasetID)
			if err != nil {
				return err
			}

			if ds == nil || ds.ID == "" {
				return fmt.Errorf("dataset %q not found", datasetID)
			}
		}
	}

	return c.ExportDatasets(cmd.Context(), s.Output, dsnames...)
}
