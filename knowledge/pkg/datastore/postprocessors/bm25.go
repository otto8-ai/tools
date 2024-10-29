package postprocessors

import (
	"context"
	"log/slog"
	"sort"

	"github.com/gptscript-ai/knowledge/pkg/datastore/lib/bm25"
	"github.com/gptscript-ai/knowledge/pkg/datastore/types"
	vs "github.com/gptscript-ai/knowledge/pkg/vectorstore/types"
)

const BM25PostprocessorName = "bm25"

type BM25Postprocessor struct {
	TopN         int
	SparseWeight float64 // How to weight the BM25 scores against the similarity scores from dense vector search

	K1 float64 // K1 should be between 1.2 and 2 - controls term frequency saturation
	B  float64 // B should be around 0.75 - controls the influence of document length normalization

	CleanStopWords []string // list of stopwords to remove from the documents - if empty, no stopwords are removed, if only "auto" is present, the language is detected automatically
}

func (c *BM25Postprocessor) Transform(ctx context.Context, response *types.RetrievalResponse) error {
	if c.K1 == 0 {
		c.K1 = bm25.DefaultK1
	}
	if c.B == 0 {
		c.B = bm25.DefaultB
	}

	for i, resp := range response.Responses {
		docs, err := c.transform(ctx, resp.Query, resp.ResultDocuments)
		if err != nil {
			return err
		}
		response.Responses[i].ResultDocuments = docs
	}
	return nil
}

func (c *BM25Postprocessor) transform(ctx context.Context, query string, docs []vs.Document) ([]vs.Document, error) {
	slog.Debug("BM25", "topN", c.TopN, "numDocs", len(docs), "sparseWeight", c.SparseWeight)

	corpus := make([]string, len(docs))
	for i, doc := range docs {
		content := doc.Content
		corpus[i] = bm25.CleanStopwords(content, c.CleanStopWords)
	}

	scores, err := bm25.Score(corpus, query, c.K1, c.B)
	if err != nil {
		return nil, err
	}

	for i, doc := range docs {
		docs[i].Metadata["bm25Score"] = scores[i]

		// Combine BM25 score with similarity score
		docs[i].Metadata["combinedScore"] = c.SparseWeight*scores[i] + (1-c.SparseWeight)*float64(doc.SimilarityScore)
	}

	// Sort by combined score
	sort.Slice(docs, func(i, j int) bool {
		return docs[i].Metadata["combinedScore"].(float64) > docs[i].Metadata["combinedScore"].(float64)
	})

	topN := c.TopN
	if topN > len(docs) {
		topN = len(docs)
	}

	return docs[:topN], nil
}

func (c *BM25Postprocessor) Name() string {
	return BM25PostprocessorName
}
