package cmd

import (
	"github.com/spf13/cobra"
)

type ClientImportDatasets struct {
	Client
}

func (s *ClientImportDatasets) Customize(cmd *cobra.Command) {
	cmd.Use = "import <path> [<dataset-id>...]"
	cmd.Short = "Import one or more datasets from an archive (zip) (default: all datasets)"
	cmd.Long = `Import one or more datasets from an archive (zip) (default: all datasets).
## IMPORTANT: Embedding functions
   When someone first ingests some data into a dataset, the embedding provider configured at that time will be attached to the dataset.
   Upon subsequent ingestion actions, the same embedding provider must be used to ensure that the embeddings are consistent.
   Most of the times, the only field that has to be the same is the model, as that defines the dimensionality usually.
   Note: This is only relevant if you plan to add more documents to the dataset after importing it.
`
	cmd.Args = cobra.MinimumNArgs(1)
}

func (s *ClientImportDatasets) Run(cmd *cobra.Command, args []string) error {
	c, err := s.getClient(cmd.Context())
	if err != nil {
		return err
	}

	return c.ImportDatasets(cmd.Context(), args[0], args[1:]...)
}
