package client

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/gptscript-ai/knowledge/pkg/datastore"
	"github.com/gptscript-ai/knowledge/pkg/datastore/documentloader"
	remotes "github.com/gptscript-ai/knowledge/pkg/datastore/documentloader/remote"
	dstypes "github.com/gptscript-ai/knowledge/pkg/datastore/types"
	"github.com/gptscript-ai/knowledge/pkg/index"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

func ingestPaths(ctx context.Context, c Client, opts *IngestPathsOpts, datasetID string, ingestionFunc func(path string, metadata map[string]any) error, paths ...string) (int, int, error) {
	ingestedFilesCount := 0
	skippedUnsupportedFilesCount := 0

	var ignoreFilePatterns []gitignore.Pattern
	var err error
	if opts.IgnoreFile != "" {
		ignoreFilePatterns, err = readIgnoreFile(opts.IgnoreFile)
		if err != nil {
			return ingestedFilesCount, skippedUnsupportedFilesCount, fmt.Errorf("failed to read ignore file %q: %w", opts.IgnoreFile, err)
		}
	}

	var ignoreExtensionsPatterns []gitignore.Pattern
	if len(opts.IgnoreExtensions) > 0 {
		for _, ext := range opts.IgnoreExtensions {
			if ext != "" {
				p := "*." + strings.TrimPrefix(ext, ".")
				ignoreExtensionsPatterns = append(ignoreExtensionsPatterns, gitignore.ParsePattern(p, nil))
			}
		}
	}

	if opts.Concurrency < 1 {
		opts.Concurrency = 10
	}
	sem := semaphore.NewWeighted(int64(opts.Concurrency)) // limit max. concurrency

	g, ctx := errgroup.WithContext(ctx)

	// Stack to store metadata when entering nested directories
	var metadataStack []Metadata

	for _, p := range paths {
		path := p

		// Build ignore matcher using patterns in increasing priority
		// 1. Default ignore file
		// 2. User-provided ignore file
		// 3. User-provided ignore extensions
		// 4. Default ignore patterns
		var currentIgnorePatterns []gitignore.Pattern
		defaultIgnoreFilePatterns, err := useDefaultIgnoreFileIfExists(path)
		if err != nil {
			return ingestedFilesCount, skippedUnsupportedFilesCount, fmt.Errorf("failed to use default ignore file: %w", err)
		}
		currentIgnorePatterns = append(defaultIgnoreFilePatterns, ignoreFilePatterns...)
		currentIgnorePatterns = append(currentIgnorePatterns, ignoreExtensionsPatterns...)
		currentIgnorePatterns = append(currentIgnorePatterns, DefaultIgnorePatterns...)

		ignore := gitignore.NewMatcher(currentIgnorePatterns)

		var touchedFilePaths []string

		if strings.HasPrefix(filepath.Base(filepath.Clean(path)), ".") {
			if !opts.IncludeHidden {
				slog.Debug("Ignoring hidden path", "path", path)
				continue
			}
		}

		if remotes.IsRemote(path) {
			// Load remote files
			remotePath, err := remotes.LoadRemote(path)
			if err != nil {
				return ingestedFilesCount, skippedUnsupportedFilesCount, fmt.Errorf("failed to load from remote %q: %w", path, err)
			}
			path = remotePath
		}

		fileInfo, err := os.Stat(path)
		if err != nil {
			return ingestedFilesCount, skippedUnsupportedFilesCount, fmt.Errorf("failed to get file info for %s: %w", path, err)
		}

		if fileInfo.IsDir() {
			directoryMetadata, err := loadDirMetadata(path)
			if err != nil {
				return ingestedFilesCount, skippedUnsupportedFilesCount, err
			}
			if directoryMetadata != nil {
				metadataStack = append(metadataStack, *directoryMetadata)
			}

			// Process directory
			err = filepath.WalkDir(path, func(subPath string, d os.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() {
					if subPath == path {
						return nil // Always process the top-level directory
					}
					if !opts.Recursive {
						return filepath.SkipDir // Skip subdirectories if not recursive
					}

					// One dir level deeper -> load new metadata
					newMetadata, err := loadDirMetadata(subPath)
					if err != nil {
						return err
					}
					if newMetadata != nil {
						metadataStack = append(metadataStack, *newMetadata)
					}
					return nil
				}

				rel, err := filepath.Rel(path, subPath)
				if err != nil {
					return fmt.Errorf("failed to get rel path, error: %w", err)
				}
				if isIgnored(ignore, rel) {
					slog.Debug("Ignoring file", "path", subPath, "ignorefile", opts.IgnoreFile, "ignoreExtensions", opts.IgnoreExtensions)
					return nil
				}

				// Process the file
				sp := subPath
				absPath, err := filepath.Abs(sp)
				if err != nil {
					return fmt.Errorf("failed to get absolute path for %s: %w", sp, err)
				}
				touchedFilePaths = append(touchedFilePaths, absPath)

				g.Go(func() error {
					if err := sem.Acquire(ctx, 1); err != nil {
						return err
					}
					defer sem.Release(1)

					fileMeta, err := findMetadata(absPath, metadataStack, opts.Metadata)
					if err != nil {
						return fmt.Errorf("failed to find metadata for %s: %w", absPath, err)
					}

					slog.Debug("Ingesting file", "absPath", absPath, "metadata", fileMeta)

					err = ingestionFunc(sp, fileMeta)
					if err != nil && !opts.ErrOnUnsupportedFile && errors.Is(err, &documentloader.UnsupportedFileTypeError{}) {
						skippedUnsupportedFilesCount++
						err = nil
					} else if err == nil {
						ingestedFilesCount++
					}
					return err
				})
				return nil
			})
			if err != nil {
				return ingestedFilesCount, skippedUnsupportedFilesCount, err
			}
		} else {
			if isIgnored(ignore, path) {
				slog.Debug("Ignoring file", "path", path, "ignorefile", opts.IgnoreFile, "ignoreExtensions", opts.IgnoreExtensions)
				continue
			}
			absPath, err := filepath.Abs(path)
			if err != nil {
				return ingestedFilesCount, skippedUnsupportedFilesCount, fmt.Errorf("failed to get absolute path for %s: %w", path, err)
			}
			touchedFilePaths = append(touchedFilePaths, absPath)

			// Process a file directly
			g.Go(func() error {
				if err := sem.Acquire(ctx, 1); err != nil {
					return err
				}
				defer sem.Release(1)

				fileMeta, err := findMetadata(absPath, metadataStack, opts.Metadata)
				if err != nil {
					return fmt.Errorf("failed to find metadata for %s: %w", absPath, err)
				}

				err = ingestionFunc(path, fileMeta)
				if err != nil && !opts.ErrOnUnsupportedFile && errors.Is(err, &documentloader.UnsupportedFileTypeError{}) {
					skippedUnsupportedFilesCount++
					err = nil
				} else if err == nil {
					ingestedFilesCount++
				}
				return err
			})
		}

		// Prune files for this basePath
		if opts.Prune && fileInfo.IsDir() {
			g.Go(func() error {
				pruned, err := c.PrunePath(ctx, datasetID, path, touchedFilePaths)
				if err != nil {
					return fmt.Errorf("failed to prune files: %w", err)
				}
				slog.Info("Pruned files", "count", len(pruned), "basePath", path)
				return nil
			})
		}
	}

	// Wait for all goroutines to finish
	return ingestedFilesCount, skippedUnsupportedFilesCount, g.Wait()
}

func HashPath(path string) string {
	hasher := sha1.New()
	hasher.Write([]byte(path))
	hashBytes := hasher.Sum(nil)
	return hex.EncodeToString(hashBytes)
}

func AskDir(ctx context.Context, c Client, path string, query string, opts *IngestPathsOpts, ropts *datastore.RetrieveOpts) (*dstypes.RetrievalResponse, error) {
	abspath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path from %q: %w", path, err)
	}

	finfo, err := os.Stat(abspath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("path %q does not exist", abspath)
		}
		return nil, fmt.Errorf("failed to get file info for %q: %w", abspath, err)
	}
	if !finfo.IsDir() {
		return nil, fmt.Errorf("path %q is not a directory", abspath)
	}

	datasetID := HashPath(abspath)
	slog.Debug("Directory Dataset ID hashed", "path", abspath, "id", datasetID)

	_, err = getOrCreateDataset(ctx, c, datasetID, true)
	if err != nil {
		return nil, err
	}

	// ingest files
	if opts == nil {
		opts = &IngestPathsOpts{
			Prune: true,
		}
	}

	ingested, skippedUnsupported, err := c.IngestPaths(ctx, datasetID, opts, path)
	if err != nil {
		return nil, fmt.Errorf("failed to ingest files: %w", err)
	}
	slog.Debug("Ingested files", "ingestedCount", ingested, "skippedUnsupported", skippedUnsupported, "path", abspath)

	// retrieve documents
	return c.Retrieve(ctx, []string{datasetID}, query, *ropts)
}

func getOrCreateDataset(ctx context.Context, c Client, datasetID string, create bool) (*index.Dataset, error) {
	var ds *index.Dataset
	var err error
	ds, err = c.GetDataset(ctx, datasetID)
	if err != nil {
		return nil, err
	}
	if ds == nil {
		if create {
			ds, err = c.CreateDataset(ctx, datasetID)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("dataset %q not found", datasetID)
		}
	}
	return ds, nil
}
