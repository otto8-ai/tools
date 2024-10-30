package datastore

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gptscript-ai/knowledge/pkg/index"
	"github.com/gptscript-ai/knowledge/pkg/vectorstore/types"
	"github.com/philippgille/chromem-go"
)

func (s *Datastore) DeleteDocument(ctx context.Context, documentID, datasetID string) error {
	// Find in Database
	var document index.Document
	tx := s.Index.WithContext(ctx).First(&document, "id = ? AND dataset = ?", documentID, datasetID)
	if tx.Error != nil {
		return ErrDBDocumentNotFound
	}

	// Remove from VectorStore
	if err := s.Vectorstore.RemoveDocument(ctx, documentID, datasetID, nil, nil); err != nil {
		return fmt.Errorf("failed to remove document from VectorStore: %w", err)
	}

	// Remove from Database
	tx = s.Index.WithContext(ctx).Delete(&document)
	if tx.Error != nil {
		return fmt.Errorf("failed to delete document from DB: %w", tx.Error)
	}

	// Check if owning file should be removed
	var count int64
	tx = s.Index.WithContext(ctx).Model(&index.Document{}).Where("file_id = ?", document.FileID).Count(&count)
	if tx.Error != nil {
		return tx.Error
	}

	if count == 0 {
		slog.Info("Removing file, because all associated documents are gone", "file", document.FileID)
		tx = s.Index.WithContext(ctx).Delete(&index.File{}, "id = ?", document.FileID)
		if tx.Error != nil {
			return fmt.Errorf("failed to delete owning file from DB: %w", tx.Error)
		}
	}

	return nil
}

func (s *Datastore) GetDocuments(ctx context.Context, datasetID string, where map[string]string, whereDocument []chromem.WhereDocument) ([]types.Document, error) {
	return s.Vectorstore.GetDocuments(ctx, datasetID, where, whereDocument)
}
