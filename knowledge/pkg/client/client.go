package client

import (
	"context"

	"github.com/gptscript-ai/knowledge/pkg/datastore"
	dstypes "github.com/gptscript-ai/knowledge/pkg/datastore/types"
	"github.com/gptscript-ai/knowledge/pkg/flows"
	types2 "github.com/gptscript-ai/knowledge/pkg/index/types"
)

type IngestWorkspaceOpts struct {
	SharedIngestionOpts
}

type SharedIngestionOpts struct {
	IngestionFlows      []flows.IngestionFlow
	IsDuplicateFuncName string
	Metadata            map[string]string
}

type IngestPathsOpts struct {
	SharedIngestionOpts
	IgnoreExtensions     []string
	Concurrency          int
	Recursive            bool
	IgnoreFile           string
	IncludeHidden        bool
	NoCreateDataset      bool
	Prune                bool // Prune deleted files
	ErrOnUnsupportedFile bool
	ExitOnFailedFile     bool
}

type Client interface {
	CreateDataset(ctx context.Context, datasetID string, opts *types2.DatasetCreateOpts) (*types2.Dataset, error)
	DeleteDataset(ctx context.Context, datasetID string) error
	GetDataset(ctx context.Context, datasetID string) (*types2.Dataset, error)
	FindFile(ctx context.Context, searchFile types2.File) (*types2.File, error)
	DeleteFile(ctx context.Context, datasetID, fileID string) error
	ListDatasets(ctx context.Context) ([]types2.Dataset, error)
	Ingest(ctx context.Context, datasetID string, name string, data []byte, opts datastore.IngestOpts) ([]string, error)
	IngestPaths(ctx context.Context, datasetID string, opts *IngestPathsOpts, paths ...string) (int, int, error) // returns number of files ingested, number of files skipped and first encountered error
	AskDirectory(ctx context.Context, path string, query string, opts *IngestPathsOpts, ropts *datastore.RetrieveOpts) (*dstypes.RetrievalResponse, error)
	PrunePath(ctx context.Context, datasetID string, path string, keep []string) ([]types2.File, error)
	DeleteDocuments(ctx context.Context, datasetID string, documentIDs ...string) error
	Retrieve(ctx context.Context, datasetIDs []string, query string, opts datastore.RetrieveOpts) (*dstypes.RetrievalResponse, error)
	ExportDatasets(ctx context.Context, path string, datasets ...string) error
	ImportDatasets(ctx context.Context, path string, datasets ...string) error
	UpdateDataset(ctx context.Context, dataset types2.Dataset, opts *datastore.UpdateDatasetOpts) (*types2.Dataset, error)
}
