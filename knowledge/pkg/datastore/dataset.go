package datastore

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/gptscript-ai/knowledge/pkg/index"
	"gorm.io/gorm"
)

type UpdateDatasetOpts struct {
	ReplaceMedata bool
}

func (s *Datastore) NewDataset(ctx context.Context, dataset index.Dataset) error {
	// Create dataset
	tx := s.Index.WithContext(ctx).Create(&dataset)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrDuplicatedKey) {
			return fmt.Errorf("dataset already exists: %w", tx.Error)
		}
		return tx.Error
	}
	tx.Commit()

	// Create collection
	err := s.Vectorstore.CreateCollection(ctx, dataset.ID)
	if err != nil {
		return err
	}
	slog.Info("Created dataset", "id", dataset.ID)
	return nil
}

func (s *Datastore) DeleteDataset(ctx context.Context, datasetID string) error {
	// Delete dataset
	slog.Info("Deleting dataset", "id", datasetID)
	tx := s.Index.WithContext(ctx).Delete(&index.Dataset{}, "id = ?", datasetID)
	if tx.Error != nil {
		return tx.Error
	}
	tx.Commit()

	// Delete collection
	err := s.Vectorstore.RemoveCollection(ctx, datasetID)
	if err != nil {
		return err
	}
	return nil
}

func (s *Datastore) GetDataset(ctx context.Context, datasetID string) (*index.Dataset, error) {
	// Get dataset with files and associated documents preloaded
	dataset := &index.Dataset{}
	tx := s.Index.WithContext(ctx).Preload("Files.Documents").First(dataset, "id = ?", datasetID)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get dataset %q from DB: %w", datasetID, tx.Error)
	}

	return dataset, nil
}

func (s *Datastore) ListDatasets(ctx context.Context) ([]index.Dataset, error) {
	tx := s.Index.WithContext(ctx).Find(&[]index.Dataset{})
	if tx.Error != nil {
		return nil, tx.Error
	}

	var datasets []index.Dataset
	if err := tx.Scan(&datasets).Error; err != nil {
		return nil, err
	}

	return datasets, nil
}

func (s *Datastore) UpdateDataset(ctx context.Context, updatedDataset index.Dataset, opts *UpdateDatasetOpts) (*index.Dataset, error) {
	if opts == nil {
		opts = &UpdateDatasetOpts{}
	}

	var origDS *index.Dataset
	var err error

	if updatedDataset.ID == "" {
		return origDS, fmt.Errorf("dataset ID is required")
	}

	origDS, err = s.GetDataset(ctx, updatedDataset.ID)
	if err != nil {
		return origDS, err
	}
	if origDS == nil {
		return origDS, fmt.Errorf("dataset not found: %s", updatedDataset.ID)
	}

	// Update Metadata
	if opts.ReplaceMedata {
		origDS.ReplaceMetadata(updatedDataset.Metadata)
	} else {
		origDS.UpdateMetadata(updatedDataset.Metadata)
	}

	if updatedDataset.EmbeddingsProviderConfig != nil {
		origDS.EmbeddingsProviderConfig = updatedDataset.EmbeddingsProviderConfig
	}

	// Check if there is any other non-null field in the updatedDataset
	if updatedDataset.Files != nil {
		return origDS, fmt.Errorf("files cannot be updated")
	}

	slog.Debug("Updating dataset", "id", updatedDataset.ID, "metadata", updatedDataset.Metadata, "embeddingsConfig", updatedDataset.EmbeddingsProviderConfig)

	return origDS, s.Index.UpdateDataset(ctx, *origDS)
}
