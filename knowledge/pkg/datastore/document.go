package datastore

import (
	"context"
	"fmt"

	"github.com/gptscript-ai/knowledge/pkg/vectorstore/types"
	"github.com/philippgille/chromem-go"
)

func (s *Datastore) DeleteDocument(ctx context.Context, documentID, datasetID string) error {
	// Remove from Index
	if err := s.Index.DeleteDocument(ctx, documentID, datasetID); err != nil {
		return fmt.Errorf("failed to remove document from Index: %w", err)
	}

	// Remove from VectorStore
	if err := s.Vectorstore.RemoveDocument(ctx, documentID, datasetID, nil, nil); err != nil {
		return fmt.Errorf("failed to remove document from VectorStore: %w", err)
	}

	return nil
}

func (s *Datastore) GetDocuments(ctx context.Context, datasetID string, where map[string]string, whereDocument []chromem.WhereDocument) ([]types.Document, error) {
	return s.Vectorstore.GetDocuments(ctx, datasetID, where, whereDocument)
}
