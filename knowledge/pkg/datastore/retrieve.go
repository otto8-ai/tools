package datastore

import (
	"context"
	"log/slog"
	"os"

	"github.com/gptscript-ai/knowledge/pkg/datastore/embeddings"
	etypes "github.com/gptscript-ai/knowledge/pkg/datastore/embeddings/types"
	"github.com/gptscript-ai/knowledge/pkg/datastore/types"
	"github.com/gptscript-ai/knowledge/pkg/output"
	types2 "github.com/gptscript-ai/knowledge/pkg/vectorstore/types"
	"github.com/mitchellh/copystructure"
	"github.com/philippgille/chromem-go"

	"github.com/gptscript-ai/knowledge/pkg/datastore/defaults"
	"github.com/gptscript-ai/knowledge/pkg/flows"

	cg "github.com/philippgille/chromem-go"
)

type RetrieveOpts struct {
	TopK          int
	Keywords      []string
	RetrievalFlow *flows.RetrievalFlow
}

func (s *Datastore) Retrieve(ctx context.Context, datasetIDs []string, query string, opts RetrieveOpts) (*types.RetrievalResponse, error) {
	slog.Debug("Retrieving content from dataset", "dataset", datasetIDs, "query", query)

	retrievalFlow := opts.RetrievalFlow
	if retrievalFlow == nil {
		retrievalFlow = &flows.RetrievalFlow{}
	}
	topK := defaults.TopK
	if opts.TopK > 0 {
		topK = opts.TopK
	}
	retrievalFlow.FillDefaults(topK)

	var whereDocs []chromem.WhereDocument
	if len(opts.Keywords) > 0 {
		whereDoc := chromem.WhereDocument{
			Operator:       chromem.WhereDocumentOperatorOr,
			WhereDocuments: []chromem.WhereDocument{},
		}
		whereDocNot := chromem.WhereDocument{
			Operator:       chromem.WhereDocumentOperatorAnd,
			WhereDocuments: []chromem.WhereDocument{},
		}
		for _, kw := range opts.Keywords {
			if kw[0] == '-' {
				whereDocNot.WhereDocuments = append(whereDocNot.WhereDocuments, chromem.WhereDocument{
					Operator: chromem.WhereDocumentOperatorNotContains,
					Value:    kw[1:],
				})
			} else {
				whereDoc.WhereDocuments = append(whereDoc.WhereDocuments, chromem.WhereDocument{
					Operator: chromem.WhereDocumentOperatorContains,
					Value:    kw,
				})
			}
		}
		if len(whereDoc.WhereDocuments) > 0 {
			whereDocs = append(whereDocs, whereDoc)
		}
		if len(whereDocNot.WhereDocuments) > 0 {
			whereDocs = append(whereDocs, whereDocNot)
		}
	}

	return retrievalFlow.Run(ctx, s, query, datasetIDs, &flows.RetrievalFlowOpts{Where: nil, WhereDocument: whereDocs})
}

func (s *Datastore) SimilaritySearch(ctx context.Context, query string, numDocuments int, datasetID string, where map[string]string, whereDocument []chromem.WhereDocument) ([]types2.Document, error) {
	ds, err := s.GetDataset(ctx, datasetID)
	if err != nil {
		return nil, err
	}
	var ef cg.EmbeddingFunc
	if ds.EmbeddingsProviderConfig != nil {
		dsEmbeddingProvider, err := embeddings.ProviderFromConfig(*ds.EmbeddingsProviderConfig)
		if err != nil {
			return nil, err
		}
		if s.EmbeddingModelProvider.EmbeddingModelName() != dsEmbeddingProvider.EmbeddingModelName() {
			slog.Warn("Embeddings model mismatch", "dataset", datasetID, "attached", dsEmbeddingProvider.EmbeddingModelName(), "configured", s.EmbeddingModelProvider.EmbeddingModelName())
			if os.Getenv("KNOW_PREFER_NEW_EMBEDDING_MODEL") == "" {
				slog.Info("Using dataset's embeddings model", "model", dsEmbeddingProvider.EmbeddingModelName())
				copied, err := copystructure.Copy(s.EmbeddingModelProvider)
				if err != nil {
					return nil, err
				}
				copied.(etypes.EmbeddingModelProvider).UseEmbeddingModel(dsEmbeddingProvider.EmbeddingModelName())
				ef, err = copied.(etypes.EmbeddingModelProvider).EmbeddingFunc()
				if err != nil {
					return nil, err
				}
				slog.Debug("Using dataset specific embedding function", "dataset", datasetID, "model", dsEmbeddingProvider.Name(), "newProviderConfig", output.RedactSensitive(copied.(etypes.EmbeddingModelProvider)))
			}
		}
	}
	return s.Vectorstore.SimilaritySearch(ctx, query, numDocuments, datasetID, where, whereDocument, ef)
}
