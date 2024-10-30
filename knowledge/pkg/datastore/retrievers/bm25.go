package retrievers

import (
	"context"
	"log/slog"
	"slices"
	"strings"

	"github.com/gptscript-ai/knowledge/pkg/datastore/lib/bm25"
	"github.com/gptscript-ai/knowledge/pkg/datastore/lib/scores"
	"github.com/gptscript-ai/knowledge/pkg/datastore/postprocessors"
	"github.com/gptscript-ai/knowledge/pkg/datastore/store"
	vs "github.com/gptscript-ai/knowledge/pkg/vectorstore/types"
	"github.com/philippgille/chromem-go"
)

const BM25RetrieverName = postprocessors.BM25PostprocessorName

type BM25Retriever struct {
	TopN int

	K1 float64 // K1 should be between 1.2 and 2 - controls term frequency saturation
	B  float64 // B should be around 0.75 - controls the influence of document length normalization

	CleanStopWords []string // list of stopwords to remove from the documents - if empty, no stopwords are removed, if only "auto" is present, the language is detected automatically
}

func (r *BM25Retriever) Name() string {
	return BM25RetrieverName
}

func (r *BM25Retriever) NormalizedScores() bool {
	return false
}

func (r *BM25Retriever) DecodeConfig(cfg map[string]any) error {
	return DefaultConfigDecoder(r, cfg)
}

func (r *BM25Retriever) Retrieve(ctx context.Context, store store.Store, query string, datasetIDs []string, where map[string]string, whereDocument []chromem.WhereDocument) ([]vs.Document, error) {
	log := slog.With("component", "BM25Retriever")

	var docs []vs.Document
	for _, datasetID := range datasetIDs {
		// TODO: make configurable via RetrieveOpts
		// silently ignore non-existent datasets
		ds, err := store.GetDataset(ctx, datasetID)
		if err != nil {
			if strings.HasPrefix(err.Error(), "dataset not found") {
				slog.Info("Dataset not found", "dataset", datasetID)
				continue
			}
			return nil, err
		}
		if ds == nil {
			continue
		}

		log.Debug("Retrieving documents from dataset", "dataset", datasetID)
		docsDataset, err := store.GetDocuments(ctx, datasetID, where, whereDocument)
		if err != nil {
			log.Error("Failed to retrieve documents from dataset", "dataset", datasetID, "error", err)
			return nil, err
		}
		docs = append(docs, docsDataset...)
	}

	if len(docs) == 0 {
		slog.Info("No documents found for BM25 retrieval", "datasets", datasetIDs)
		return nil, nil
	}

	bm25scores, err := bm25.BM25Run(docs, query, r.K1, r.B, r.CleanStopWords)
	if err != nil {
		log.Error("Failed to run BM25", "error", err)
		return nil, err
	}

	for i, _ := range docs {
		docs[i].SimilarityScore = float32(bm25scores[i])
		// slog.Debug("BM25 score", "docID", docs[i].ID, "score", docs[i].SimilarityScore)
	}

	slices.SortFunc(docs, scores.SortBySimilarityScore)

	topN := r.TopN
	if topN > len(docs) {
		topN = len(docs)
	}

	return docs[:topN], nil
}
