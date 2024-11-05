package store

import (
	"context"

	"github.com/gptscript-ai/knowledge/pkg/index/types"
	vs "github.com/gptscript-ai/knowledge/pkg/vectorstore/types"
	"github.com/philippgille/chromem-go"
)

type Store interface {
	ListDatasets(ctx context.Context) ([]types.Dataset, error)
	GetDataset(ctx context.Context, datasetID string) (*types.Dataset, error)
	SimilaritySearch(ctx context.Context, query string, numDocuments int, collection string, where map[string]string, whereDocument []chromem.WhereDocument) ([]vs.Document, error)
	GetDocuments(ctx context.Context, datasetID string, where map[string]string, whereDocument []chromem.WhereDocument) ([]vs.Document, error)
}
