package sqlite_vec

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"strings"

	sqlitevec "github.com/asg017/sqlite-vec-go-bindings/ncruces"
	vs "github.com/gptscript-ai/knowledge/pkg/vectorstore/types"
	"github.com/ncruces/go-sqlite3"
	cg "github.com/philippgille/chromem-go"
)

type VectorStore struct {
	embeddingFunc       cg.EmbeddingFunc
	db                  *sqlite3.Conn
	embeddingsTableName string
}

func New(ctx context.Context, dsn string, embeddingFunc cg.EmbeddingFunc) (*VectorStore, error) {
	dsn = "file:" + strings.TrimPrefix(dsn, "sqlite-vec://")

	slog.Debug("sqlite-vec", "dsn", dsn)
	db, err := sqlite3.Open(dsn)
	if err != nil {
		return nil, err
	}

	store := &VectorStore{
		embeddingFunc:       embeddingFunc,
		db:                  db,
		embeddingsTableName: "knowledge_embeddings",
	}

	stmt, _, err := db.Prepare(`SELECT sqlite_version(), vec_version()`)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize sqlite-vec: %w", err)
	}

	stmt.Step()
	slog.Debug("sqlite-vec info", "sqlite_version", stmt.ColumnText(0), "vec_version", stmt.ColumnText(1))
	err = stmt.Close()
	if err != nil {
		return nil, err
	}

	return store, store.prepareTables(ctx)
}

func (v *VectorStore) Close() error {
	return v.db.Close()
}

func (v *VectorStore) prepareTables(ctx context.Context) error {
	err := v.db.Exec(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s
		(
			id TEXT PRIMARY KEY,
			collection_id TEXT NOT NULL,
			content TEXT,
			metadata JSON
		)
		;
	`, v.embeddingsTableName))
	if err != nil {
		return fmt.Errorf("failed to create %s table: %w", v.embeddingsTableName, err)
	}

	return nil
}

func (v *VectorStore) CreateCollection(ctx context.Context, collection string) error {
	emb, err := v.embeddingFunc(ctx, "dummy text")
	if err != nil {
		return fmt.Errorf("failed to get embedding: %w", err)
	}
	dimensionality := len(emb) // FIXME: somehow allow to pass this in or set it globally

	return v.db.Exec(fmt.Sprintf(`CREATE VIRTUAL TABLE IF NOT EXISTS %s_vec USING
	vec0(
		document_id TEXT PRIMARY KEY,
		embedding float[%d] distance_metric=cosine
	)
    `, collection, dimensionality))
}

func (v *VectorStore) AddDocuments(ctx context.Context, docs []vs.Document, collection string) ([]string, error) {
	stmt, _, err := v.db.Prepare(fmt.Sprintf(`INSERT INTO %s_vec(document_id, embedding) VALUES (?, ?)`, collection))
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	ids := make([]string, len(docs))
	for docIdx, doc := range docs {
		emb, err := v.embeddingFunc(ctx, doc.Content)
		if err != nil {
			return nil, fmt.Errorf("failed to get embedding: %w", err)
		}
		v, err := sqlitevec.SerializeFloat32(emb)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize embedding: %w", err)
		}
		if err := stmt.BindText(1, doc.ID); err != nil {
			return nil, fmt.Errorf("failed to bind document_id: %w", err)
		}
		if err := stmt.BindBlob(2, v); err != nil {
			return nil, fmt.Errorf("failed to bind embedding: %w", err)
		}
		if err := stmt.Exec(); err != nil {
			return nil, fmt.Errorf("failed to insert document (vector): %w", err)
		}
		if err := stmt.Reset(); err != nil {
			return nil, fmt.Errorf("failed to reset statement: %w", err)
		}
		ids[docIdx] = doc.ID
	}

	if err := stmt.Close(); err != nil {
		return nil, fmt.Errorf("failed to close statement: %w", err)
	}

	// add to embeddings table
	stmt, _, err = v.db.Prepare(fmt.Sprintf(`INSERT INTO %s(id, collection_id, content, metadata) VALUES (?, ?, ?, ?)`, v.embeddingsTableName))
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	for _, doc := range docs {
		if err := stmt.BindText(1, doc.ID); err != nil {
			return nil, fmt.Errorf("failed to bind document_id: %w", err)
		}
		if err := stmt.BindText(2, collection); err != nil {
			return nil, fmt.Errorf("failed to bind collection_id: %w", err)
		}
		if err := stmt.BindText(3, doc.Content); err != nil {
			return nil, fmt.Errorf("failed to bind content: %w", err)
		}
		if err := stmt.BindJSON(4, doc.Metadata); err != nil {
			return nil, fmt.Errorf("failed to bind metadata: %w", err)
		}
		if err := stmt.Exec(); err != nil {
			return nil, fmt.Errorf("failed to insert document (embeddings table): %w", err)
		}
		if err := stmt.Reset(); err != nil {
			return nil, fmt.Errorf("failed to reset statement: %w", err)
		}
	}

	return ids, nil
}

func (v *VectorStore) SimilaritySearch(ctx context.Context, query string, numDocuments int, collection string, where map[string]string, whereDocument []cg.WhereDocument) ([]vs.Document, error) {
	q, err := v.embeddingFunc(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get embedding: %w", err)
	}

	stmt, _, err := v.db.Prepare(fmt.Sprintf(`SELECT document_id, distance FROM %s_vec WHERE embedding MATCH ? ORDER BY distance LIMIT %d`, collection, numDocuments))
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	qv, err := sqlitevec.SerializeFloat32(q)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize query embedding: %w", err)
	}

	if err := stmt.BindBlob(1, qv); err != nil {
		return nil, fmt.Errorf("failed to bind query embedding: %w", err)
	}

	var docs []vs.Document
	for stmt.Step() {
		docID := stmt.ColumnText(0)
		distance := stmt.ColumnFloat(1)
		docs = append(docs, vs.Document{ID: docID, SimilarityScore: float32(1 - distance)})
	}
	if stmt.Err() != nil {
		return nil, fmt.Errorf("failed to execute statement: %w", stmt.Err())
	}
	if err := stmt.Close(); err != nil {
		return nil, fmt.Errorf("failed to close statement: %w", err)
	}

	nstmt, _, err := v.db.Prepare(fmt.Sprintf(`SELECT content, metadata FROM %s WHERE id = ?`, v.embeddingsTableName))
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}
	for i, doc := range docs {
		if err := nstmt.BindText(1, doc.ID); err != nil {
			return nil, fmt.Errorf("failed to bind document_id : %w", err)
		}
		if nstmt.Step() {
			doc.Content = nstmt.ColumnText(0)
			err = nstmt.ColumnJSON(1, &doc.Metadata)
			if err != nil {
				return nil, fmt.Errorf("failed to get metadata: %w", err)
			}
		}
		if nstmt.Err() != nil {
			return nil, fmt.Errorf("failed to execute statement: %w", stmt.Err())
		}
		docs[i] = doc
		if err = nstmt.Reset(); err != nil {
			return nil, fmt.Errorf("failed to reset statement: %w", err)
		}
	}

	if err := nstmt.Close(); err != nil {
		return nil, fmt.Errorf("failed to close statement: %w", err)
	}

	return docs, nil
}

func (v *VectorStore) RemoveCollection(ctx context.Context, collection string) error {
	err := v.db.Exec(fmt.Sprintf(`DROP TABLE IF EXISTS %s_vec`, collection))
	if err != nil {
		return fmt.Errorf("failed to drop table: %w", err)
	}

	err = v.db.Exec(fmt.Sprintf(`DELETE FROM %s WHERE collection_id = %s`, v.embeddingsTableName, collection))
	if err != nil {
		return fmt.Errorf("failed to delete documents: %w", err)
	}

	return nil
}

func (v *VectorStore) RemoveDocument(ctx context.Context, documentID string, collection string, where map[string]string, whereDocument []cg.WhereDocument) error {
	if len(whereDocument) > 0 {
		return fmt.Errorf("sqlite-vec does not support whereDocument")
	}

	var ids []string

	// delete by metadata filter
	if len(where) > 0 {
		whereQueries := make([]string, 0)
		for k, v := range where {
			if strings.TrimSpace(k) == "" || strings.TrimSpace(v) == "" {
				continue
			}
			whereQueries = append(whereQueries, fmt.Sprintf("(metadata ->> '$.%s') = '%s'", k, v))
		}
		whereQuery := strings.Join(whereQueries, " AND ")
		if len(whereQuery) == 0 {
			whereQuery = "TRUE"
		}

		stmt, _, err := v.db.Prepare(fmt.Sprintf(`SELECT id FROM %s WHERE collection_id = '%s' AND %s`, v.embeddingsTableName, collection, whereQuery))
		if err != nil {
			return fmt.Errorf("failed to prepare statement: %w", err)
		}

		for stmt.Step() {
			ids = append(ids, stmt.ColumnText(0))
		}

		if stmt.Err() != nil {
			return fmt.Errorf("failed to execute statement: %w", stmt.Err())
		}
	} else {
		ids = []string{documentID}
	}

	slog.Debug("deleting documents from sqlite-vec", "ids", ids)

	// delete by ID
	embStmt, _, err := v.db.Prepare(fmt.Sprintf(`DELETE FROM %s_vec WHERE document_id = ?`, collection))
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}

	colStmt, _, err := v.db.Prepare(fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, v.embeddingsTableName))
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}

	for _, id := range ids {
		slog.Debug("deleting document from sqlite-vec", "id", id)

		if err := embStmt.BindText(1, id); err != nil {
			return fmt.Errorf("failed to bind document_id: %w", err)
		}

		if err := colStmt.BindText(1, id); err != nil {
			return fmt.Errorf("failed to bind document_id: %w", err)
		}

		if err := embStmt.Exec(); err != nil {
			return fmt.Errorf("failed to delete document (vector): %w", err)
		}

		if err := colStmt.Exec(); err != nil {
			return fmt.Errorf("failed to delete document (embeddings table): %w", err)
		}

		if err := embStmt.Reset(); err != nil {
			return fmt.Errorf("failed to reset statement: %w", err)
		}

		if err := colStmt.Reset(); err != nil {
			return fmt.Errorf("failed to reset statement: %w", err)
		}
	}

	return nil
}

func (v *VectorStore) GetDocuments(ctx context.Context, collection string, where map[string]string, whereDocument []cg.WhereDocument) ([]vs.Document, error) {
	if len(whereDocument) > 0 {
		return nil, fmt.Errorf("sqlite-vec does not support whereDocument")
	}

	var docs []vs.Document

	// delete by metadata filter
	if len(where) > 0 {
		whereQueries := make([]string, 0)
		for k, v := range where {
			if strings.TrimSpace(k) == "" || strings.TrimSpace(v) == "" {
				continue
			}
			whereQueries = append(whereQueries, fmt.Sprintf("(metadata ->> '$.%s') = '%s'", k, v))
		}
		whereQuery := strings.Join(whereQueries, " AND ")
		if len(whereQuery) == 0 {
			whereQuery = "TRUE"
		}

		stmt, _, err := v.db.Prepare(fmt.Sprintf(`SELECT id, content, metadata FROM %s WHERE collection_id = '%s' AND %s`, v.embeddingsTableName, collection, whereQuery))
		if err != nil {
			return nil, fmt.Errorf("failed to prepare statement: %w", err)
		}

		for stmt.Step() {
			doc := vs.Document{
				ID:      stmt.ColumnText(0),
				Content: stmt.ColumnText(1),
			}

			err = stmt.ColumnJSON(2, &doc.Metadata)
			if err != nil {
				return nil, fmt.Errorf("failed to get metadata: %w", err)
			}
			docs = append(docs, doc)
		}

		if stmt.Err() != nil {
			return nil, fmt.Errorf("failed to execute statement: %w", stmt.Err())
		}
	}
	return docs, nil
}

func (v *VectorStore) ImportCollectionsFromFile(ctx context.Context, path string, collections ...string) error {
	return fmt.Errorf("not implemented")
}

func (v *VectorStore) ExportCollectionsToFile(ctx context.Context, path string, collections ...string) error {
	return fmt.Errorf("not implemented")
}
