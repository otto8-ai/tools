package chromem

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gptscript-ai/knowledge/pkg/datastore/types"
	"github.com/gptscript-ai/knowledge/pkg/env"
	"github.com/gptscript-ai/knowledge/pkg/log"
	"github.com/gptscript-ai/knowledge/pkg/vectorstore/errors"
	vs "github.com/gptscript-ai/knowledge/pkg/vectorstore/types"
	"github.com/philippgille/chromem-go"
)

// VsChromemEmbeddingParallelThread can be set as an environment variable to control the number of parallel API calls to create embedding for documents. Default is 100
const VsChromemEmbeddingParallelThread = "VS_CHROMEM_EMBEDDING_PARALLEL_THREAD"

type ChromemStore struct {
	db            *chromem.DB
	embeddingFunc chromem.EmbeddingFunc
}

// New creates a new Chromem vector store.
// Three types are supported:
// 1. In-memory: chromem://:memory:
// 2. Persistent: chromem://path/to/db-file
// 3. In-memory, loaded from archive: chromem://archive://path/to/archive-file
func New(dsn string, embeddingFunc chromem.EmbeddingFunc) (*ChromemStore, error) {
	dsn = strings.TrimPrefix(dsn, "chromem://")

	if dsn == ":memory:" {
		return &ChromemStore{
			db:            chromem.NewDB(),
			embeddingFunc: embeddingFunc,
		}, nil
	}

	var vsdb *chromem.DB
	var err error
	if strings.HasPrefix(dsn, types.ArchivePrefix) {
		// Import from archive -> in-memory DB, not persisted back to the archive
		vsdb = chromem.NewDB()
		if err = vsdb.ImportFromFile(strings.TrimPrefix(dsn, types.ArchivePrefix), ""); err != nil {
			return nil, fmt.Errorf("failed to import vector database: %w", err)
		}
	} else {
		vsdb, err = chromem.NewPersistentDB(dsn, false, chromem.WithOnCorruptedCollectionBehavior(chromem.OnCorruptedDelete))
		if err != nil {
			return nil, err
		}
	}

	return &ChromemStore{
		db:            vsdb,
		embeddingFunc: embeddingFunc,
	}, nil
}

func (s *ChromemStore) CreateCollection(_ context.Context, name string) error {
	_, err := s.db.CreateCollection(name, nil, s.embeddingFunc)
	if err != nil {
		return err
	}

	return nil
}

func (s *ChromemStore) AddDocuments(ctx context.Context, docs []vs.Document, collection string) ([]string, error) {
	l := log.FromCtx(ctx).With("stage", "vectorstore").With("vectorstore", "chromem-go")

	l.With("status", "starting").Info("Adding documents to collection (generating embeddings)")

	ids := make([]string, len(docs))
	chromemDocs := make([]chromem.Document, len(docs))
	for docIdx, doc := range docs {
		ids[docIdx] = doc.ID
		mc := make(map[string]any)
		maps.Copy(mc, doc.Metadata)
		if len(doc.Content) == 0 {
			slog.Debug("Document has no content", "id", ids[docIdx], "index", docIdx)
			doc.Content = "<no content>"
		}
		chromemDocs[docIdx] = chromem.Document{
			ID:        ids[docIdx],
			Metadata:  anyMapToStringMap(mc),
			Embedding: nil, // Embeddings will be computed downstream
			Content:   doc.Content,
		}
	}

	col := s.db.GetCollection(collection, s.embeddingFunc)
	if col == nil {
		l.With("status", "failed").With("error", errors.ErrCollectionNotFound.Error()).Error("Collection not found", "collection", collection)
		return nil, fmt.Errorf("%w: %q", errors.ErrCollectionNotFound, collection)
	}

	concurrency := env.GetIntFromEnvOrDefault(VsChromemEmbeddingParallelThread, 100)

	err := col.AddDocuments(ctx, chromemDocs, concurrency)
	if err != nil {
		l.With("status", "failed").With("error", err.Error()).Error("Failed to add documents to collection (generate embeddings)")
		return nil, err
	}

	l.With("status", "completed").Info("Added documents to collection (generated embeddings)")

	return ids, nil
}

func anyMapToStringMap(m map[string]any) map[string]string {
	convertedMap := make(map[string]string)

	for key, value := range m {
		switch v := value.(type) {
		case string:
			convertedMap[key] = v
		case int:
			convertedMap[key] = strconv.Itoa(v)
		case bool:
			convertedMap[key] = strconv.FormatBool(v)
		case float64:
			convertedMap[key] = strconv.FormatFloat(v, 'f', -1, 64)
		default:
			x, ok := value.(string)
			if ok {
				convertedMap[key] = x
			}
			// skip unsupported types for now
		}
	}
	return convertedMap
}

func convertStringMapToAnyMap(m map[string]string) map[string]any {
	convertedMap := make(map[string]any)

	for key, value := range m {
		convertedMap[key] = value
	}
	return convertedMap
}

func (s *ChromemStore) SimilaritySearch(ctx context.Context, query string, numDocuments int, collection string, where map[string]string, whereDocument []chromem.WhereDocument) ([]vs.Document, error) {
	col := s.db.GetCollection(collection, s.embeddingFunc)
	if col == nil {
		return nil, fmt.Errorf("%w: %q", errors.ErrCollectionNotFound, collection)
	}

	if col.Count() == 0 {
		return nil, fmt.Errorf("%w: %q", errors.ErrCollectionEmpty, collection)
	}

	if numDocuments > col.Count() {
		numDocuments = col.Count()
		slog.Debug("Reduced number of documents to search for", "numDocuments", numDocuments)
	}

	slog.Debug("filtering documents", "where", where, "whereDocument", whereDocument)

	qr, err := col.Query(ctx, query, numDocuments, where, whereDocument)
	if err != nil {
		return nil, err
	}

	if len(qr) == 0 {
		return nil, nil
	}

	var sDocs []vs.Document

	for _, qrd := range qr {
		sDocs = append(sDocs, vs.Document{
			ID:              qrd.ID,
			Metadata:        convertStringMapToAnyMap(qrd.Metadata),
			SimilarityScore: qrd.Similarity,
			Content:         qrd.Content,
		})
	}

	return sDocs, nil
}

func (s *ChromemStore) RemoveCollection(_ context.Context, collection string) error {
	return s.db.DeleteCollection(collection)
}

func (s *ChromemStore) RemoveDocument(ctx context.Context, documentID string, collection string, where map[string]string, whereDocument []chromem.WhereDocument) error {
	col := s.db.GetCollection(collection, s.embeddingFunc)
	if col == nil {
		return fmt.Errorf("%w: %q", errors.ErrCollectionNotFound, collection)
	}
	return col.Delete(ctx, where, whereDocument, documentID)
}

func (s *ChromemStore) ImportCollectionsFromFile(ctx context.Context, path string, collections ...string) error {
	finfo, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("couldn't stat file %q: %w", path, err)
	}
	if finfo.IsDir() {
		return fmt.Errorf("path %q is a directory", path)
	}
	slog.Debug("Importing collections from file", "path", path)
	return s.db.ImportFromFile(path, "", collections...)
}

func (s *ChromemStore) ExportCollectionsToFile(ctx context.Context, path string, collections ...string) error {
	finfo, err := os.Stat(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("couldn't stat file %q: %w", path, err)
	}
	if finfo.IsDir() {
		path = filepath.Join(path, "chromem-export.gob")
	}
	slog.Debug("Exporting collections to file", "path", path)
	return s.db.ExportToFile(path, false, "", collections...)
}

func (s *ChromemStore) GetDocuments(ctx context.Context, collection string, where map[string]string, whereDocument []chromem.WhereDocument) ([]vs.Document, error) {
	col := s.db.GetCollection(collection, s.embeddingFunc)
	if col == nil {
		return nil, fmt.Errorf("%w: %q", errors.ErrCollectionNotFound, collection)
	}

	cdocs, err := col.GetDocuments(ctx, nil, nil)
	if err != nil {
		return nil, err
	}

	var docs []vs.Document
	for _, doc := range cdocs {
		docs = append(docs, vs.Document{
			ID:       doc.ID,
			Metadata: convertStringMapToAnyMap(doc.Metadata),
			Content:  doc.Content,
		})
	}

	return docs, nil
}
