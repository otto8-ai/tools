package postprocessors

import (
	"context"
	"log/slog"

	"github.com/gptscript-ai/knowledge/pkg/datastore/types"
	vs "github.com/gptscript-ai/knowledge/pkg/vectorstore/types"
)

const SimilarityPostprocessorName = "similarity"

type SimilarityPostprocessor struct {
	Threshold float32
	KeepMin   int // KeepMin the top n documents, regardless of the threshold
}

func (s *SimilarityPostprocessor) Transform(ctx context.Context, response *types.RetrievalResponse) error {
	for i, resp := range response.Responses {
		docCount := len(resp.ResultDocuments)
		var filteredDocs []vs.Document
		for _, doc := range resp.ResultDocuments {
			if doc.SimilarityScore >= s.Threshold {
				filteredDocs = append(filteredDocs, doc)
			} else {
				if len(filteredDocs) < s.KeepMin {
					// Note: this is assuming that the documents are sorted by similarity score
					filteredDocs = append(filteredDocs, doc)
					slog.Debug("Keeping document below threshold", "docID", doc.ID, "score", doc.SimilarityScore, "threshold", s.Threshold)
				}
			}
		}
		response.Responses[i].ResultDocuments = filteredDocs
		slog.Debug("Filtered documents", "originalDocCount", docCount, "docsBelowThreshold", len(filteredDocs), "keepMin", s.KeepMin, "threshold", s.Threshold)
	}
	return nil
}

func (s *SimilarityPostprocessor) Name() string {
	return SimilarityPostprocessorName
}
