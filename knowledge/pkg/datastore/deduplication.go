package datastore

import (
	"context"
	"errors"
	"log/slog"

	"github.com/gptscript-ai/knowledge/pkg/index/types"
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
	searchMeta := types.FileMetadata{
		AbsolutePath: opts.FileMetadata.AbsolutePath,
		Size:         opts.FileMetadata.Size,
		ModifiedAt:   opts.FileMetadata.ModifiedAt,
	}

	res, err := d.Index.FindFileByMetadata(ctx, datasetID, searchMeta, false)
	if err != nil && !errors.Is(err, types.ErrDBFileNotFound) {
		return false, err
	}

	if res == nil || res.ID == "" {
		return false, nil
	}

	return true, nil
}

func DedupeUpsert(ctx context.Context, d *Datastore, datasetID string, content []byte, opts IngestOpts) (bool, error) {
	searchMeta := types.FileMetadata{
		AbsolutePath: opts.FileMetadata.AbsolutePath,
	}

	res, err := d.Index.FindFileByMetadata(ctx, datasetID, searchMeta, false)
	if err != nil && !errors.Is(err, types.ErrDBFileNotFound) {
		return false, err
	}

	if res == nil || res.ID == "" {
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
