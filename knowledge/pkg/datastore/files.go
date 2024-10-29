package datastore

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/gptscript-ai/knowledge/pkg/index"
	"gorm.io/gorm"
)

// ErrDBFileNotFound is returned when a file is not found.
var ErrDBFileNotFound = errors.New("file not found in database")

func (s *Datastore) DeleteFile(ctx context.Context, datasetID, fileID string) error {
	// Find file in database with associated documents
	var file index.File
	tx := s.Index.WithContext(ctx).Preload("Documents").Where("id = ? AND dataset = ?", fileID, datasetID).First(&file)
	if tx.Error != nil {
		return ErrDBFileNotFound
	}

	// Remove owned documents from VectorStore and Database
	for _, doc := range file.Documents {
		if err := s.Vectorstore.RemoveDocument(ctx, doc.ID, datasetID, nil, nil); err != nil {
			return fmt.Errorf("failed to remove document from VectorStore: %w", err)
		}

		tx = s.Index.WithContext(ctx).Delete(&doc)
		if tx.Error != nil {
			return fmt.Errorf("failed to delete document from DB: %w", tx.Error)
		}
	}

	// Remove file DB
	tx = s.Index.WithContext(ctx).Delete(&file)
	if tx.Error != nil {
		return fmt.Errorf("failed to delete file from DB: %w", tx.Error)
	}

	return nil
}

func (s *Datastore) PruneFiles(ctx context.Context, datasetID string, pathPrefix string, keep []string) ([]index.File, error) {
	var files []index.File
	tx := s.Index.WithContext(ctx).
		Where("dataset = ?", datasetID).
		Where("absolute_path LIKE ?", pathPrefix+"%").
		Not("absolute_path IN ?", keep).
		Find(&files)
	if tx.Error != nil {
		return nil, tx.Error
	}

	slog.Debug("Pruning files", "count", len(files), "dataset", datasetID, "path_prefix", pathPrefix, "keep", keep)

	for _, file := range files {
		if err := s.DeleteFile(ctx, datasetID, file.ID); err != nil {
			return nil, err
		}
	}

	return files, nil
}

func (s *Datastore) FindFile(ctx context.Context, searchFile index.File) (*index.File, error) {
	if searchFile.Dataset == "" {
		return nil, fmt.Errorf("dataset must be provided")
	}

	var file index.File
	var tx *gorm.DB
	if searchFile.ID != "" {
		tx = s.Index.WithContext(ctx).Preload("Documents").Where("dataset = ? AND id = ?", searchFile.Dataset, searchFile.ID).First(&file)
	} else if searchFile.AbsolutePath != "" {
		tx = s.Index.WithContext(ctx).Preload("Documents").Where("dataset = ? AND absolute_path = ?", searchFile.Dataset, searchFile.AbsolutePath).First(&file)
	} else {
		return nil, fmt.Errorf("either fileID or fileAbsPath must be provided")
	}
	if tx.Error != nil {
		return nil, ErrDBFileNotFound
	}

	return &file, nil
}
