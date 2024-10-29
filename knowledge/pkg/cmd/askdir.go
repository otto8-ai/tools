package cmd

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/acorn-io/z"
	"github.com/gptscript-ai/knowledge/pkg/client"
	"github.com/gptscript-ai/knowledge/pkg/datastore"
	flowconfig "github.com/gptscript-ai/knowledge/pkg/flows/config"
	"github.com/spf13/cobra"
)

type ClientAskDir struct {
	Client
	Path    string `usage:"Path to the directory to query" short:"p" default:"."`
	NoPrune bool   `usage:"Do not prune deleted files" env:"KNOW_ASKDIR_NO_PRUNE"`
	ClientIngestOpts
	ClientRetrieveOpts
	ClientFlowsConfig
}

func (s *ClientAskDir) Customize(cmd *cobra.Command) {
	cmd.Use = "askdir [--path <path>] <query>"
	cmd.Short = "Retrieve sources for a query from a dataset generated from a directory"
	cmd.Args = cobra.ExactArgs(1)
}

func (s *ClientAskDir) Run(cmd *cobra.Command, args []string) error {
	c, err := s.getClient(cmd.Context())
	if err != nil {
		return err
	}

	path := s.Path

	slog.Debug("Asking directory", "path", path, "GPTSCRIPT_WORKSPACE_DIR", os.Getenv("GPTSCRIPT_WORKSPACE_DIR"))

	query := args[0]

	ingestOpts := &client.IngestPathsOpts{
		SharedIngestionOpts: client.SharedIngestionOpts{
			IsDuplicateFuncName: s.DeduplicationFuncName,
		},
		IgnoreExtensions:     strings.Split(s.IgnoreExtensions, ","),
		Concurrency:          s.Concurrency,
		Recursive:            !s.NoRecursive,
		IgnoreFile:           s.IgnoreFile,
		IncludeHidden:        s.IncludeHidden,
		Prune:                !s.NoPrune,
		ErrOnUnsupportedFile: s.ErrOnUnsupportedFile,
	}

	retrieveOpts := &datastore.RetrieveOpts{
		TopK:     s.TopK,
		Keywords: s.Keywords,
	}

	if s.FlowsFile != "" {
		abspath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("failed to get absolute path from %q: %w", path, err)
		}

		datasetID := client.HashPath(abspath)

		slog.Debug("Loading ingestion flows from config", "flows_file", s.FlowsFile, "dataset", datasetID)
		flowCfg, err := flowconfig.Load(s.FlowsFile)
		if err != nil {
			return err
		}
		var flow *flowconfig.FlowConfigEntry
		if s.Flow != "" {
			flow, err = flowCfg.GetFlow(s.Flow)
			if err != nil {
				return err
			}
		} else {
			flow, err = flowCfg.ForDataset(datasetID) // get flow for the dataset
			if err != nil {
				return err
			}
		}

		for _, ingestionFlowConfig := range flow.Ingestion {
			ingestionFlow, err := ingestionFlowConfig.AsIngestionFlow(&flow.Globals.Ingestion)
			if err != nil {
				return err
			}
			ingestOpts.IngestionFlows = append(ingestOpts.IngestionFlows, z.Dereference(ingestionFlow))
		}
		slog.Debug("Loaded ingestion flows from config", "flows_file", s.FlowsFile, "dataset", datasetID, "flows", len(ingestOpts.IngestionFlows))

		if flow.Retrieval == nil {
			slog.Info("No retrieval config in assigned flow", "flows_file", s.FlowsFile, "dataset", datasetID)
		} else {
			rf, err := flow.Retrieval.AsRetrievalFlow()
			if err != nil {
				return err
			}
			retrieveOpts.RetrievalFlow = rf
			slog.Debug("Loaded retrieval flow from config", "flows_file", s.FlowsFile, "dataset", datasetID)
		}
	}

	retrievalResp, err := c.AskDirectory(cmd.Context(), path, query, ingestOpts, retrieveOpts)
	if err != nil {
		return fmt.Errorf("failed to retrieve sources: %w", err)
	}

	if len(retrievalResp.Responses) == 0 {
		fmt.Printf("No sources found for the query %q from path %q\n", query, path)
		return nil
	}

	jsonSources, err := json.Marshal(retrievalResp.Responses)
	if err != nil {
		return err
	}

	fmt.Printf("Retrieved the following %d source collections for the query %q (keywords: %q) from path %q: %s\n", len(retrievalResp.Responses), query, strings.Join(retrieveOpts.Keywords, ","), path, jsonSources)

	return nil
}
