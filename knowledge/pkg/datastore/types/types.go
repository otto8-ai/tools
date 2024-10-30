package types

import (
	"context"

	vs "github.com/gptscript-ai/knowledge/pkg/vectorstore/types"
)

const (
	ArchivePrefix = "archive://"
)

type DocumentTransformerFunc func(context.Context, []vs.Document) ([]vs.Document, error)

type DocumentTransformer interface {
	Transform(context.Context, []vs.Document) ([]vs.Document, error)
	Name() string
}

type DocumentLoader interface {
	Load(ctx context.Context) ([]vs.Document, error)
	LoadAndSplit(ctx context.Context, splitter TextSplitter) ([]vs.Document, error)
}

type TextSplitter interface {
	SplitDocuments(docs []vs.Document) ([]vs.Document, error)
}

type Response struct {
	Query           string        `json:"subquery"`
	NumDocs         int           `json:"numResultDocuments"`
	ResultDocuments []vs.Document `json:"resultDocuments"`
}

type Stats struct {
	RetrievalTimeSeconds float64 `json:"retrievalTimeSeconds,omitempty"`
}

type RetrievalResponse struct {
	Query     string     `json:"originalQuery"`
	Datasets  []string   `json:"queriedDatasets"`
	Responses []Response `json:"subqueryResults"`
	Stats     Stats      `json:"stats,omitempty"`
}
