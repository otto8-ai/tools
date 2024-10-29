package datastore

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"log/slog"

	"github.com/gptscript-ai/knowledge/pkg/index"
)

// IsDuplicateFunc is a function that determines whether a document is a duplicate or if it should be ingested.
// The function should return true if the document is a duplicate (and thus should not be ingested) and false otherwise.
type IsDuplicateFunc func(ctx context.Context, d *Datastore, datasetID string, content []byte, opts IngestOpts) (bool, error)

// IsDuplicateFuncs is a map of deduplication functions by name.
var IsDuplicateFuncs = map[string]IsDuplicateFunc{
	"file_metadata": DedupeByFileMetadata,
	"dummy":         DummyDedupe,
	"none":          DummyDedupe,
	"ignore":        DummyDedupe,
	"upsert":        DedupeUpsert,
}

// DedupeByFileMetadata is a deduplication function that checks if the document is a duplicate based on the file metadata.
func DedupeByFileMetadata(ctx context.Context, d *Datastore, datasetID string, content []byte, opts IngestOpts) (bool, error) {
	var count int64
	err := d.Index.WithContext(ctx).Model(&index.File{}).
		Where("dataset = ?", datasetID).
		Where("absolute_path = ?", opts.FileMetadata.AbsolutePath).
		Where("size = ?", opts.FileMetadata.Size).
		Where("modified_at = ?", opts.FileMetadata.ModifiedAt).
		Count(&count).Error

	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func DedupeUpsert(ctx context.Context, d *Datastore, datasetID string, content []byte, opts IngestOpts) (bool, error) {
	var res index.File
	err := d.Index.WithContext(ctx).Model(&index.File{}).
		Where("dataset = ?", datasetID).
		Where("absolute_path = ?", opts.FileMetadata.AbsolutePath).
		First(&res).Error

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, err
	}

	if res.ID == "" {
		return false, nil
	}

	// If incoming file is newer than the existing file, delete the existing file
	if res.ModifiedAt.Before(opts.FileMetadata.ModifiedAt) {
		slog.Debug("Upserting by deleting existing file", "file", res.ID, "absPath", res.AbsolutePath, "modified_at", res.ModifiedAt, "new_modified_at", opts.FileMetadata.ModifiedAt)
		err = d.DeleteFile(ctx, datasetID, res.ID)
		if err != nil {
			return false, err
		}
		return false, nil
	}

	slog.Debug("Not upserting: incoming file is not newer", "file", res.ID, "absPath", res.AbsolutePath, "modified_at", res.ModifiedAt, "new_modified_at", opts.FileMetadata.ModifiedAt)

	return true, nil
}

// DummyDedupe is a dummy deduplication function that always returns false (i.e. "No Duplicate").
func DummyDedupe(ctx context.Context, d *Datastore, datasetID string, content []byte, opts IngestOpts) (bool, error) {
	return false, nil
}
