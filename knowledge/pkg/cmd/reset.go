package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/gptscript-ai/knowledge/pkg/datastore"
	"github.com/spf13/cobra"
)

type ClientResetDatastore struct {
	Client
}

func (s *ClientResetDatastore) Customize(cmd *cobra.Command) {
	cmd.Use = "reset-datastore"
	cmd.Short = "Reset the knowledge datastore (WARNING: this deletes all datasets and ingested data)"
	cmd.Args = cobra.ExactArgs(0)
	cmd.Hidden = true
}

func (s *ClientResetDatastore) Run(cmd *cobra.Command, args []string) error {
	// TODO: adjust to new changeable datastore parts
	indexDSN, vectorDSN, _, err := datastore.GetDefaultDSNs(s.DatabaseConfig.DSN, s.VectorDBConfig.DSN)
	if err != nil {
		return err
	}

	if err := os.RemoveAll(strings.TrimPrefix(indexDSN, "sqlite://")); err != nil {
		return fmt.Errorf("failed to remove database file: %w", err)
	}

	if err := os.RemoveAll(vectorDSN); err != nil {
		return fmt.Errorf("failed to remove vector database directory: %w", err)
	}

	fmt.Printf("Successfully reset datastore (DSN: %q, DSN: %q)\n", indexDSN, vectorDSN)
	return nil
}
