package datastore

import (
	"context"
	"errors"
	"fmt"

	"github.com/gptscript-ai/knowledge/pkg/index/types"
)

// ErrDBFileNotFound is returned when a file is not found.
var ErrDBFileNotFound = errors.New("file not found in database")

func (s *Datastore) DeleteFile(ctx context.Context, datasetID, fileID string) error {
	// Find file
	search := types.File{ID: fileID, Dataset: datasetID}
	file, err := s.Index.FindFile(ctx, search)
	if err != nil {
		return fmt.Errorf("failed to find file in DB: %w", err)
	}

	// Remove owned documents from VectorStore and Database
	for _, doc := range file.Documents {
		if err := s.Vectorstore.RemoveDocument(ctx, doc.ID, datasetID, nil, nil); err != nil {
			return fmt.Errorf("failed to remove document from VectorStore: %w", err)
		}
	}

	// Remove file DB
	return s.Index.DeleteFile(ctx, datasetID, fileID)
}

func (s *Datastore) PruneFiles(ctx context.Context, datasetID string, pathPrefix string, keep []string) ([]types.File, error) {
	return s.Index.PruneFiles(ctx, datasetID, pathPrefix, keep)
}

func (s *Datastore) FindFile(ctx context.Context, searchFile types.File) (*types.File, error) {
	return s.Index.FindFile(ctx, searchFile)
}
