package pgvector

/* DISCLAIMER: Heavily inspired by https://github.com/tmc/langchaingo/tree/5e330db17991a2e259cd5fa4c1350a7e1e4787ab/vectorstores/pgvector (Thank you!) */

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/gptscript-ai/knowledge/pkg/env"
	vs "github.com/gptscript-ai/knowledge/pkg/vectorstore/types"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
	cg "github.com/philippgille/chromem-go"
)

const (
	// pgLockIDEmbeddingTable is used for advisor lock to fix issue arising from concurrent
	// creation of the embedding table.The same value represents the same lock.
	pgLockIDEmbeddingTable = 1573678846307946494
	// pgLockIDCollectionTable is used for advisor lock to fix issue arising from concurrent
	// creation of the collection table.The same value represents the same lock.
	pgLockIDCollectionTable = 1573678846307946495
	// pgLockIDExtension is used for advisor lock to fix issue arising from concurrent creation
	// of the vector extension. The value is deliberately set to the same as python langchain
	// https://github.com/langchain-ai/langchain/blob/v0.0.340/libs/langchain/langchain/vectorstores/pgvector.py#L167
	pgLockIDExtension = 1573678846307946496

	// pgLockIDCreateCollection is used for advisor lock to fix issue arising from concurrent
	// creation of the collection. The same value represents the same lock.
	pgLockIDCreateCollection = 1573678846307946497

	// VsPgvectorEmbeddingConcurrency can be set as an environment variable to control the number of parallel API calls to create embedding for documents. Default is 100
	VsPgvectorEmbeddingConcurrency = "VS_PGVECTOR_EMBEDDING_CONCURRENCY"
)

var (
	ErrEmbedderWrongNumberVectors = errors.New("number of vectors from embedder does not match number of documents")
	ErrInvalidScoreThreshold      = errors.New("score threshold must be between 0 and 1")
	ErrInvalidFilters             = errors.New("invalid filters")
	ErrUnsupportedOptions         = errors.New("unsupported options")
)

// PGXConn represents both a pgx.Conn and pgxpool.Pool conn.
type PGXConn interface {
	Ping(ctx context.Context) error
	Begin(ctx context.Context) (pgx.Tx, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, arguments ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, arguments ...any) pgx.Row
	SendBatch(ctx context.Context, batch *pgx.Batch) pgx.BatchResults
}

type CloseNoErr interface {
	Close()
}

type VectorStore struct {
	embeddingFunc        cg.EmbeddingFunc
	embeddingConcurrency int
	conn                 PGXConn
	embeddingTableName   string
	collectionTableName  string
	vectorDimensions     int
	hnswIndex            *HNSWIndex
}

// HNSWIndex lets you specify the HNSW index parameters.
// See here for more details: https://github.com/pgvector/pgvector#hnsw
//
// m: he max number of connections per layer (16 by default)
// efConstruction: the size of the dynamic candidate list for constructing the graph (64 by default)
// distanceFunction: the distance function to use (l2 by default).
type HNSWIndex struct {
	m                int
	efConstruction   int
	distanceFunction string
}

var DefaultHNSWIndex = &HNSWIndex{
	m:                16,
	efConstruction:   64,
	distanceFunction: "vector_l2_ops",
}

func New(ctx context.Context, dsn string, embeddingFunc cg.EmbeddingFunc) (*VectorStore, error) {
	dsn = "postgres://" + strings.TrimPrefix(dsn, "pgvector://")

	store := &VectorStore{
		embeddingTableName:   "knowledge_embeddings",
		collectionTableName:  "knowledge_collections",
		embeddingFunc:        embeddingFunc,
		embeddingConcurrency: env.GetIntFromEnvOrDefault(VsPgvectorEmbeddingConcurrency, 100),
		hnswIndex:            nil,
	}

	var err error
	store.conn, err = pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	if err = store.conn.Ping(ctx); err != nil {
		return nil, err
	}

	return store, store.init(ctx)
}

func (v VectorStore) init(ctx context.Context) error {
	tx, err := v.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) // rollback on error (noop after commit)
	if err := v.createVectorExtensionIfNotExists(ctx, tx); err != nil {
		return err
	}
	if err := v.createCollectionTableIfNotExists(ctx, tx); err != nil {
		return err
	}
	if err := v.createEmbeddingTableIfNotExists(ctx, tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (v VectorStore) createVectorExtensionIfNotExists(ctx context.Context, tx pgx.Tx) error {
	// inspired by
	// https://github.com/langchain-ai/langchain/blob/v0.0.340/libs/langchain/langchain/vectorstores/pgvector.py#L167
	// The advisor lock fixes issue arising from concurrent
	// creation of the vector extension.
	// https://github.com/langchain-ai/langchain/issues/12933
	// For more information see:
	// https://www.postgresql.org/docs/16/explicit-locking.html#ADVISORY-LOCKS
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", pgLockIDExtension); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, "CREATE EXTENSION IF NOT EXISTS vector"); err != nil {
		return err
	}
	return nil
}

func (v VectorStore) createCollectionTableIfNotExists(ctx context.Context, tx pgx.Tx) error {
	// inspired by
	// https://github.com/langchain-ai/langchain/blob/v0.0.340/libs/langchain/langchain/vectorstores/pgvector.py#L167
	// The advisor lock fixes issue arising from concurrent
	// creation of the vector extension.
	// https://github.com/langchain-ai/langchain/issues/12933
	// For more information see:
	// https://www.postgresql.org/docs/16/explicit-locking.html#ADVISORY-LOCKS
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", pgLockIDCollectionTable); err != nil {
		return err
	}
	sql := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
	name varchar,
	cmetadata json,
	"uuid" uuid NOT NULL,
	UNIQUE (name),
	PRIMARY KEY (uuid))`, v.collectionTableName)
	if _, err := tx.Exec(ctx, sql); err != nil {
		return err
	}
	return nil
}

func (v VectorStore) createEmbeddingTableIfNotExists(ctx context.Context, tx pgx.Tx) error {
	// inspired by
	// https://github.com/langchain-ai/langchain/blob/v0.0.340/libs/langchain/langchain/vectorstores/pgvector.py#L167
	// The advisor lock fixes issue arising from concurrent
	// creation of the vector extension.
	// https://github.com/langchain-ai/langchain/issues/12933
	// For more information see:
	// https://www.postgresql.org/docs/16/explicit-locking.html#ADVISORY-LOCKS
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", pgLockIDEmbeddingTable); err != nil {
		return err
	}

	vectorDimensions := ""
	if v.vectorDimensions > 0 {
		vectorDimensions = fmt.Sprintf("(%d)", v.vectorDimensions)
	}

	sql := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
	collection_id uuid,
	embedding vector%s,
	document bytea,
	cmetadata json,
	"uuid" uuid NOT NULL,
	CONSTRAINT knowledge_pg_embedding_collection_id_fkey
	FOREIGN KEY (collection_id) REFERENCES %s (uuid) ON DELETE CASCADE,
	PRIMARY KEY (uuid))`, v.embeddingTableName, vectorDimensions, v.collectionTableName)
	if _, err := tx.Exec(ctx, sql); err != nil {
		return err
	}
	sql = fmt.Sprintf(`CREATE INDEX IF NOT EXISTS %s_collection_id ON %s (collection_id)`, v.embeddingTableName, v.embeddingTableName)
	if _, err := tx.Exec(ctx, sql); err != nil {
		return err
	}

	// See this for more details on HNSW indexes: https://github.com/pgvector/pgvector#hnsw
	if v.hnswIndex != nil {
		sql = fmt.Sprintf(
			`CREATE INDEX IF NOT EXISTS %s_embedding_hnsw ON %s USING hnsw (embedding %s)`,
			v.embeddingTableName, v.embeddingTableName, v.hnswIndex.distanceFunction,
		)
		if v.hnswIndex.m > 0 && v.hnswIndex.efConstruction > 0 {
			sql = fmt.Sprintf("%s WITH (m=%d, ef_construction = %d)", sql, v.hnswIndex.m, v.hnswIndex.efConstruction)
		}
		if _, err := tx.Exec(ctx, sql); err != nil {
			return err
		}
	}

	return nil
}

func (v VectorStore) getCollectionUUID(ctx context.Context, collection string) (string, error) {
	var cuuid string
	err := v.conn.QueryRow(ctx, fmt.Sprintf(`SELECT uuid FROM %s WHERE name=$1`, v.collectionTableName), collection).Scan(&cuuid)
	if err != nil {
		return "", err
	}
	return cuuid, nil
}

func (v VectorStore) CreateCollection(ctx context.Context, collection string) error {
	slog.Debug("Creating collection", "collection", collection, "store", "pgvector")
	tx, err := v.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) // rollback on error (noop after commit)

	// Acquire an advisory lock
	_, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", pgLockIDCreateCollection)
	if err != nil {
		return fmt.Errorf("failed to acquire advisory lock: %w", err)
	}

	_, err = tx.Exec(ctx, fmt.Sprintf(`INSERT INTO %s (uuid, name) VALUES($1, $2)`, v.collectionTableName), uuid.New().String(), collection)
	if err != nil {
		return fmt.Errorf("failed to create collection %s: %w", collection, err)
	}

	return tx.Commit(ctx)
}

func (v VectorStore) AddDocuments(ctx context.Context, docs []vs.Document, collection string) ([]string, error) {
	cid, err := v.getCollectionUUID(ctx, collection)
	if err != nil {
		return nil, err
	}

	texts := make([]string, 0, len(docs))
	for _, doc := range docs {
		texts = append(texts, doc.Content)
	}

	b := &pgx.Batch{}
	ids := make([]string, len(docs))

	var sharedErr error
	sharedErrLock := sync.Mutex{}
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)
	setSharedErr := func(err error) {
		sharedErrLock.Lock()
		defer sharedErrLock.Unlock()
		// Another goroutine might have already set the error.
		if sharedErr == nil {
			sharedErr = err
			// Cancel the operation for all other goroutines.
			cancel(sharedErr)
		}
	}

	sql := fmt.Sprintf(`INSERT INTO %s (uuid, document, embedding, cmetadata, collection_id)
		VALUES($1, $2, $3, $4, $5)`, v.embeddingTableName)

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, v.embeddingConcurrency)
	for docIdx, doc := range docs {
		id := uuid.New().String()
		ids[docIdx] = id
		doc.ID = id

		wg.Add(1)
		go func(doc vs.Document) {
			defer wg.Done()

			// Don't even start if another goroutine already failed.
			if ctx.Err() != nil {
				return
			}

			// Wait here while $concurrency other goroutines are creating documents.
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			vec, err := v.embeddingFunc(ctx, doc.Content)
			if err != nil {
				setSharedErr(fmt.Errorf("failed to embed document %s: %w", doc.ID, err))
				return
			}

			b.Queue(sql, doc.ID, []byte(doc.Content), pgvector.NewVector(vec), doc.Metadata, cid)
		}(doc)

		docs[docIdx] = doc
	}
	wg.Wait()

	if sharedErr != nil {
		return nil, sharedErr
	}

	return ids, v.conn.SendBatch(ctx, b).Close()
}

/*
SimilaritySearch performs a similarity search on the given query and returns the most similar documents.
* pgvector supports different distance functions: https://github.com/pgvector/pgvector/blob/master/README.md#querying
* Supported distance functions are (more have been added since writing this):
*   - `<->` - L2 distance
*   - `<#>` - (negative) inner product
*   - `<=>` - cosine distance (for cosine similarity, use 1 - cosine distance)
*   - `<+>` - L1 distance (added in 0.7.0)
*   - `<~>` - Hamming distance (binary vectors, added in 0.7.0)
*   - `<%>` - Jaccard distance (binary vectors, added in 0.7.0)
*/
func (v VectorStore) SimilaritySearch(ctx context.Context, query string, numDocuments int, collection string, where map[string]string, whereDocument []cg.WhereDocument) ([]vs.Document, error) {
	slog.Debug("Similarity search", "query", query, "numDocuments", numDocuments, "collection", collection, "where", where, "whereDocument", whereDocument, "store", "pgvector")

	if len(whereDocument) > 0 {
		return nil, fmt.Errorf("pgvector does not support whereDocument")
	}

	queryEmbedding, err := v.embeddingFunc(ctx, query)
	if err != nil {
		return nil, err
	}
	whereQuerys := make([]string, 0)
	for k, v := range where {
		whereQuerys = append(whereQuerys, fmt.Sprintf("(data.cmetadata ->> '%s') = '%s'", k, v))
	}
	whereQuery := strings.Join(whereQuerys, " AND ")
	if len(whereQuery) == 0 {
		whereQuery = "TRUE"
	}
	dims := len(queryEmbedding)
	sql := fmt.Sprintf(`WITH filtered_embedding_dims AS MATERIALIZED (
    SELECT
        *
    FROM
        %s
    WHERE
        vector_dims (
                embedding
        ) = $1
)
SELECT
	data.uuid,
	data.document,
	data.cmetadata,
	data.similarity
FROM (
	SELECT
		filtered_embedding_dims.*,
		1 - (embedding <=> $2) AS similarity
	FROM
		filtered_embedding_dims
		JOIN %s ON filtered_embedding_dims.collection_id=%s.uuid WHERE %s.name='%s') AS data
WHERE %s
ORDER BY
	data.similarity DESC
LIMIT $3`, v.embeddingTableName,
		v.collectionTableName, v.collectionTableName, v.collectionTableName, collection,
		whereQuery)
	rows, err := v.conn.Query(ctx, sql, dims, pgvector.NewVector(queryEmbedding), numDocuments)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	docs := make([]vs.Document, 0)
	for rows.Next() {
		doc := vs.Document{}
		var contentB []byte
		if err := rows.Scan(&doc.ID, &contentB, &doc.Metadata, &doc.SimilarityScore); err != nil {
			return nil, err
		}
		doc.Content = string(contentB)
		docs = append(docs, doc)
	}
	return docs, rows.Err()
}

func (v VectorStore) RemoveCollection(ctx context.Context, collection string) error {
	slog.Debug("Removing collection", "collection", collection, "store", "pgvector")

	tx, err := v.conn.Begin(ctx)
	if err != nil {
		return err
	}

	// Deletion from the collection table will cascade to the embedding table
	_, err = tx.Exec(ctx, fmt.Sprintf(`DELETE FROM %s WHERE name = $1`, v.collectionTableName), collection)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (v VectorStore) RemoveDocument(ctx context.Context, documentID string, collection string, where map[string]string, whereDocument []cg.WhereDocument) error {
	if len(whereDocument) > 0 {
		return fmt.Errorf("pgvector does not support whereDocument")
	}

	cid, err := v.getCollectionUUID(ctx, collection)
	if err != nil {
		return fmt.Errorf("collection %s not found: %w", collection, err)
	}

	// query to check if there are any docs at all
	var count int
	err = v.conn.QueryRow(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE collection_id = $1`, v.embeddingTableName), cid).Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		return nil
	}
	slog.Info("Removing document", "documentID", documentID, "collection", collection, "where", where, "existingDocs", count)

	// Where clause takes precedence over documentID for consistency with chromem-go's behavior, as that was the default before
	if len(where) > 0 {
		whereQuerys := make([]string, 0)
		for k, v := range where {
			whereQuerys = append(whereQuerys, fmt.Sprintf("(cmetadata ->> '%s') = '%s'", k, v))
		}
		whereQuery := strings.Join(whereQuerys, " AND ")
		if len(whereQuery) == 0 {
			whereQuery = "TRUE"
		}
		_, err = v.conn.Exec(ctx, fmt.Sprintf(`DELETE FROM %s WHERE collection_id = $1 AND  %s`, v.embeddingTableName, whereQuery), cid)
		return err
	}

	_, err = v.conn.Exec(ctx, fmt.Sprintf(`DELETE FROM %s WHERE uuid = $1 AND collection_id = $2`, v.embeddingTableName), documentID, cid)
	return err
}

func (v VectorStore) GetDocuments(ctx context.Context, collection string, where map[string]string, whereDocument []cg.WhereDocument) ([]vs.Document, error) {
	if len(whereDocument) > 0 {
		return nil, fmt.Errorf("pgvector does not support whereDocument")
	}

	cid, err := v.getCollectionUUID(ctx, collection)
	if err != nil {
		return nil, err
	}

	whereQuerys := make([]string, 0)
	for k, v := range where {
		whereQuerys = append(whereQuerys, fmt.Sprintf("(cmetadata ->> '%s') = '%s'", k, v))
	}
	whereQuery := strings.Join(whereQuerys, " AND ")
	if len(whereQuery) == 0 {
		whereQuery = "TRUE"
	}

	sql := fmt.Sprintf(`SELECT uuid, document, cmetadata FROM %s WHERE collection_id = $1 AND %s`, v.embeddingTableName, whereQuery)
	rows, err := v.conn.Query(ctx, sql, cid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	docs := make([]vs.Document, 0)
	for rows.Next() {
		doc := vs.Document{}
		if err := rows.Scan(&doc.ID, &doc.Content, &doc.Metadata); err != nil {
			return nil, err
		}
		docs = append(docs, doc)
	}
	return docs, rows.Err()
}

func (v VectorStore) ImportCollectionsFromFile(ctx context.Context, path string, collections ...string) error {
	return fmt.Errorf("function ImportCollectionsFromFile not implemented for vectorstore pgvector")
}

func (v VectorStore) ExportCollectionsToFile(ctx context.Context, path string, collections ...string) error {
	return fmt.Errorf("function ExportCollectionsToFile not implemented for vectorstore pgvector")
}
