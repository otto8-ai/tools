package datastore

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gptscript-ai/knowledge/pkg/config"
	etypes "github.com/gptscript-ai/knowledge/pkg/datastore/embeddings/types"
	"github.com/gptscript-ai/knowledge/pkg/log"
	"github.com/gptscript-ai/knowledge/pkg/output"

	"github.com/adrg/xdg"
	"github.com/gptscript-ai/knowledge/pkg/index"
	"github.com/gptscript-ai/knowledge/pkg/llm"
	"github.com/gptscript-ai/knowledge/pkg/vectorstore"
	cg "github.com/philippgille/chromem-go"
)

type Datastore struct {
	LLM                    llm.LLM
	Index                  *index.DB
	Vectorstore            vectorstore.VectorStore
	EmbeddingConfig        config.EmbeddingsConfig
	EmbeddingModelProvider etypes.EmbeddingModelProvider
}

// GetDefaultDSNs returns the paths for the datastore and vectorstore databases.
// In addition, it returns a boolean indicating whether the datastore is an archive.
func GetDefaultDSNs(indexDSN, vectorDSN string) (string, string, bool, error) {
	var isArchive bool

	if strings.Contains(indexDSN, "://archive://") || strings.Contains(vectorDSN, "://archive://") {
		isArchive = true
	}

	if indexDSN == "" {
		var err error
		indexDSN, err = xdg.DataFile("gptscript/knowledge/knowledge.db")
		if err != nil {
			return "", "", isArchive, err
		}
		indexDSN = "sqlite://" + indexDSN
		slog.Debug("Using default Index DSN", "indexDSN", indexDSN)
	}

	if vectorDSN == "" {
		path, err := xdg.DataFile("gptscript/knowledge/vector.db")
		if err != nil {
			return "", "", isArchive, err
		}

		vectorDSN = "chromem://" + path

		slog.Debug("Using default Vector DSN", "vectorDSN", vectorDSN)
	}

	return indexDSN, vectorDSN, isArchive, nil
}

func LogEmbeddingFunc(embeddingFunc cg.EmbeddingFunc) cg.EmbeddingFunc {
	return func(ctx context.Context, text string) ([]float32, error) {
		l := log.FromCtx(ctx).With("stage", "embedding")

		l.With("status", "starting").Info("Creating embedding")

		embedding, err := embeddingFunc(ctx, text)
		if err != nil {
			l.With("status", "failed").Error("Failed to create embedding", "error", err)
			return nil, err
		}

		l.With("status", "completed").Info("Created embedding")
		return embedding, nil
	}
}

func NewDatastore(ctx context.Context, indexDSN string, automigrate bool, vectorDSN string, embeddingProvider etypes.EmbeddingModelProvider) (*Datastore, error) {
	indexDSN, vectorDSN, isArchive, err := GetDefaultDSNs(indexDSN, vectorDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to determine datastore paths: %w", err)
	}

	idx, err := index.New(ctx, indexDSN, automigrate)
	if err != nil {
		return nil, err
	}

	if err := idx.AutoMigrate(); err != nil {
		return nil, fmt.Errorf("failed to auto-migrate index: %w", err)
	}

	slog.Debug("Using embedding model provider", "provider", embeddingProvider.Name(), "config", output.RedactSensitive(embeddingProvider.Config()))

	vsdb, err := vectorstore.New(ctx, vectorDSN, embeddingProvider)
	if err != nil {
		return nil, err
	}

	ds := &Datastore{
		Index:                  idx,
		Vectorstore:            vsdb,
		EmbeddingModelProvider: embeddingProvider,
	}

	// If loaded from archive, do not create a default dataset
	if isArchive {
		return ds, nil
	}

	// Ensure default dataset exists
	defaultDS, err := ds.GetDataset(context.Background(), "default")
	if err != nil {
		return nil, fmt.Errorf("failed to ensure default dataset: %w", err)
	}

	if defaultDS == nil {
		err = ds.NewDataset(context.Background(), index.Dataset{ID: "default"})
		if err != nil {
			return nil, fmt.Errorf("failed to create default dataset: %w", err)
		}
	}

	return ds, nil
}

func (s *Datastore) ExportDatasetsToFile(ctx context.Context, path string, datasets ...string) error {
	tmpDir, err := os.MkdirTemp(os.TempDir(), "knowledge-export-")
	if err != nil {
		return err
	}

	defer os.RemoveAll(tmpDir)

	if err = s.Index.ExportDatasetsToFile(ctx, tmpDir, datasets...); err != nil {
		return err
	}

	if err = s.Vectorstore.ExportCollectionsToFile(ctx, tmpDir, datasets...); err != nil {
		return err
	}

	finfo, err := os.Stat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		if err = os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
	}

	// make sure target path is a file
	if finfo != nil && finfo.IsDir() {
		path = filepath.Join(path, fmt.Sprintf("knowledge-export-%s.zip", time.Now().Format("2006-01-02-15-04-05")))
	}

	// zip it up
	if err = zipDir(tmpDir, path); err != nil {
		return err
	}

	return nil
}

func (s *Datastore) ImportDatasetsFromFile(ctx context.Context, path string, datasets ...string) error {
	tmpDir, err := os.MkdirTemp(os.TempDir(), "knowledge-import-")
	if err != nil {
		return err
	}

	defer os.RemoveAll(tmpDir)

	r, err := zip.OpenReader(path)
	if err != nil {
		return err
	}
	defer r.Close()

	if len(r.File) != 2 {
		return fmt.Errorf("knowledge archive must contain exactly two files, found %d", len(r.File))
	}

	dbFile := ""
	vectorStoreFile := ""
	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		path := filepath.Join(tmpDir, f.Name)
		if f.FileInfo().IsDir() {
			continue
		}

		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err := io.Copy(f, rc); err != nil {
			return err
		}
		_ = f.Close()
		_ = rc.Close()

		// FIXME: this should not be static as we may support multiple (vector) DBs at some point
		if filepath.Ext(f.Name()) == ".db" {
			dbFile = path
		} else if filepath.Ext(f.Name()) == ".gob" {
			vectorStoreFile = path
		}
	}

	if dbFile == "" || vectorStoreFile == "" {
		return fmt.Errorf("knowledge archive must contain exactly one .db and one .gob file")
	}

	if err = s.Index.ImportDatasetsFromFile(ctx, dbFile); err != nil {
		return err
	}

	if err = s.Vectorstore.ImportCollectionsFromFile(ctx, vectorStoreFile, datasets...); err != nil {
		return err
	}

	return nil
}

func zipDir(src, dst string) error {
	zipfile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	// Create a new zip archive.
	w := zip.NewWriter(zipfile)
	defer w.Close()

	// Walk the file tree and add files to the zip archive.
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get the file information
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		if !info.IsDir() {
			// Update the header name
			header.Name = filepath.Base(path)

			// Write the header
			writer, err := w.CreateHeader(header)
			if err != nil {
				return err
			}

			// If the file is not a directory, write the file to the archive
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(writer, file)
			if err != nil {
				return err
			}
		}
		return nil
	})
}
