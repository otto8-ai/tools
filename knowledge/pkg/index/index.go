package index

import (
	"context"

	"github.com/gptscript-ai/knowledge/pkg/index/types"
)

type Index interface {
	// Database Ops
	AutoMigrate() error

	// Fundamental Dataset Operations
	CreateDataset(ctx context.Context, dataset types.Dataset) error
	GetDataset(ctx context.Context, datasetID string) (*types.Dataset, error)
	ListDatasets(ctx context.Context) ([]types.Dataset, error)
	DeleteDataset(ctx context.Context, datasetID string) error

	// Advanced Dataset Operations
	ExportDatasetsToFile(ctx context.Context, path string, ids ...string) error
	ImportDatasetsFromFile(ctx context.Context, path string) error
	UpdateDataset(ctx context.Context, dataset types.Dataset) error

	// Fundamental File Operations
	CreateFile(ctx context.Context, file types.File) error
	DeleteFile(ctx context.Context, datasetID, fileID string) error
	FindFile(ctx context.Context, searchFile types.File) (*types.File, error)
	FindFileByMetadata(ctx context.Context, dataset string, metadata types.FileMetadata, includeDocuments bool) (*types.File, error)

	// Advanced File Operations
	PruneFiles(ctx context.Context, datasetID string, pathPrefix string, keep []string) ([]types.File, error)

	// Fundamental Document Operations
	DeleteDocument(ctx context.Context, documentID, datasetID string) error
}
