package datastore

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gptscript-ai/knowledge/pkg/index/types"
)

type UpdateDatasetOpts struct {
	ReplaceMedata bool
}

func (s *Datastore) CreateDataset(ctx context.Context, dataset types.Dataset, opts *types.DatasetCreateOpts) error {
	// Create dataset
	if err := s.Index.CreateDataset(ctx, dataset, opts); err != nil {
		return err
	}

	// Create collection
	err := s.Vectorstore.CreateCollection(ctx, dataset.ID, opts)
	if err != nil {
		return err
	}
	slog.Info("Created dataset", "id", dataset.ID)
	return nil
}

func (s *Datastore) DeleteDataset(ctx context.Context, datasetID string) error {
	// Delete dataset
	if err := s.Index.DeleteDataset(ctx, datasetID); err != nil {
		return err
	}

	// Delete collection
	err := s.Vectorstore.RemoveCollection(ctx, datasetID)
	if err != nil {
		return err
	}
	return nil
}

func (s *Datastore) GetDataset(ctx context.Context, datasetID string) (*types.Dataset, error) {
	return s.Index.GetDataset(ctx, datasetID)
}

func (s *Datastore) ListDatasets(ctx context.Context) ([]types.Dataset, error) {
	return s.Index.ListDatasets(ctx)
}

func (s *Datastore) UpdateDataset(ctx context.Context, updatedDataset types.Dataset, opts *UpdateDatasetOpts) (*types.Dataset, error) {
	if opts == nil {
		opts = &UpdateDatasetOpts{}
	}

	var origDS *types.Dataset
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
