package datastore

import (
	"context"
	"log/slog"

	"github.com/gptscript-ai/knowledge/pkg/datastore/types"
	types2 "github.com/gptscript-ai/knowledge/pkg/vectorstore/types"
	"github.com/philippgille/chromem-go"

	"github.com/gptscript-ai/knowledge/pkg/datastore/defaults"
	"github.com/gptscript-ai/knowledge/pkg/flows"
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
	return s.Vectorstore.SimilaritySearch(ctx, query, numDocuments, datasetID, where, whereDocument)
}
