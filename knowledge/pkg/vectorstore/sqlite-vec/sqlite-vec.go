package sqlite_vec

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	sqlitevec "github.com/asg017/sqlite-vec-go-bindings/ncruces"
	dbtypes "github.com/gptscript-ai/knowledge/pkg/index/types"
	vs "github.com/gptscript-ai/knowledge/pkg/vectorstore/types"
	cg "github.com/philippgille/chromem-go"
	"gorm.io/gorm"

	"github.com/ncruces/go-sqlite3/gormlite"
)

type VectorStore struct {
	embeddingFunc       cg.EmbeddingFunc
	db                  *gorm.DB
	embeddingsTableName string
}

func New(ctx context.Context, dsn string, embeddingFunc cg.EmbeddingFunc) (*VectorStore, error) {
	dsn = "file:" + strings.TrimPrefix(dsn, "sqlite-vec://")

	slog.Debug("sqlite-vec", "dsn", dsn)
	db, err := gorm.Open(gormlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Enable PRAGMAs
	// - busy_timeout (ms) to prevent db lockups as we're accessing the DB from multiple separate processes in acorn
	tx := db.Exec(`
PRAGMA busy_timeout = 5000;
`)
	if tx.Error != nil {
		return nil, tx.Error
	}

	store := &VectorStore{
		embeddingFunc:       embeddingFunc,
		db:                  db,
		embeddingsTableName: "knowledge_embeddings",
	}

	var sqliteVersion, vecVersion string
	res := db.Raw("SELECT sqlite_version(), vec_version()").Row()
	if err := res.Scan(&sqliteVersion, &vecVersion); err != nil {
		return nil, fmt.Errorf("failed to scan results: %w", err)
	}
	slog.Debug("sqlite-vec info", "sqlite_version", sqliteVersion, "vec_version", vecVersion)

	return store, store.prepareTables(ctx)
}

func (v *VectorStore) Close() error {
	sqldb, err := v.db.DB()
	if err != nil {
		return err
	}
	return sqldb.Close()
}

func (v *VectorStore) prepareTables(ctx context.Context) error {
	err := v.db.Exec(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS [%s]
		(
			id TEXT PRIMARY KEY,
			collection_id TEXT NOT NULL,
			content TEXT,
			metadata JSON
		)
		;
	`, v.embeddingsTableName)).Error
	if err != nil {
		return fmt.Errorf("failed to create embeddings table %q: %w", v.embeddingsTableName, err)
	}

	return nil
}

func (v *VectorStore) CreateCollection(ctx context.Context, collection string, opts *dbtypes.DatasetCreateOpts) error {
	emb, err := v.embeddingFunc(ctx, "dummy text")
	if err != nil {
		return fmt.Errorf("failed to get embedding: %w", err)
	}
	dimensionality := len(emb) // FIXME: somehow allow to pass this in or set it globally

	err = v.db.Exec(fmt.Sprintf(`CREATE VIRTUAL TABLE IF NOT EXISTS [%s_vec] USING
	vec0(
		document_id TEXT PRIMARY KEY,
		embedding float[%d] distance_metric=cosine
	)
    `, collection, dimensionality)).Error

	if err != nil {
		return fmt.Errorf("failed to create vector table: %w", err)
	}

	return nil
}

func (v *VectorStore) AddDocuments(ctx context.Context, docs []vs.Document, collection string) ([]string, error) {
	ids := make([]string, len(docs))

	err := v.db.Transaction(func(tx *gorm.DB) error {
		if len(docs) > 0 {
			valuePlaceholders := make([]string, len(docs))
			args := make([]interface{}, 0, len(docs)*2) // 2 args per doc: document_id and embedding

			for i, doc := range docs {
				emb, err := v.embeddingFunc(ctx, doc.Content)
				if err != nil {
					return fmt.Errorf("failed to compute embedding for document %s: %w", doc.ID, err)
				}

				serializedEmb, err := sqlitevec.SerializeFloat32(emb)
				if err != nil {
					return fmt.Errorf("failed to serialize embedding for document %s: %w", doc.ID, err)
				}

				valuePlaceholders[i] = "(?, ?)"
				args = append(args, doc.ID, serializedEmb)

				ids[i] = doc.ID
			}

			// Raw query for *_vec as gorm doesn't support virtual tables
			query := fmt.Sprintf(`
				INSERT INTO [%s_vec] (document_id, embedding)
				VALUES %s
			`, collection, strings.Join(valuePlaceholders, ", "))

			if err := tx.Exec(query, args...).Error; err != nil {
				return fmt.Errorf("failed to batch insert into vector table: %w", err)
			}
		}

		embs := make([]map[string]interface{}, len(docs))
		for i, doc := range docs {
			metadataJson, err := json.Marshal(doc.Metadata)
			if err != nil {
				return fmt.Errorf("failed to marshal metadata for document %s: %w", doc.ID, err)
			}
			embs[i] = map[string]interface{}{
				"id":            doc.ID,
				"collection_id": collection,
				"content":       doc.Content,
				"metadata":      metadataJson,
			}
		}

		// Use GORM's Create for the embeddings table
		if err := tx.Table(v.embeddingsTableName).Create(embs).Error; err != nil {
			return fmt.Errorf("failed to batch insert into embeddings table: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return ids, nil
}

func (v *VectorStore) SimilaritySearch(ctx context.Context, query string, numDocuments int, collection string, where map[string]string, whereDocument []cg.WhereDocument, embeddingFunc cg.EmbeddingFunc) ([]vs.Document, error) {
	ef := v.embeddingFunc
	if embeddingFunc != nil {
		ef = embeddingFunc
	}

	q, err := ef(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to compute embedding: %w", err)
	}

	qv, err := sqlitevec.SerializeFloat32(q)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize query embedding: %w", err)
	}

	var docs []vs.Document
	err = v.db.Transaction(func(tx *gorm.DB) error {
		// Query matching document IDs and distances
		rows, err := tx.Raw(fmt.Sprintf(`
            SELECT document_id, distance 
            FROM [%s_vec]
            WHERE embedding MATCH ? 
            ORDER BY distance 
            LIMIT ?
        `, collection), qv, numDocuments).Rows()
		if err != nil {
			return fmt.Errorf("failed to query vector table: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var docID string
			var distance float32
			if err := rows.Scan(&docID, &distance); err != nil {
				return fmt.Errorf("failed to scan row: %w", err)
			}
			docs = append(docs, vs.Document{
				ID:              docID,
				SimilarityScore: 1 - distance, // Higher score means closer match
			})
		}

		// Fetch content and metadata for each document
		for i, doc := range docs {
			var content string
			var metadataJSON []byte
			err := tx.Raw(fmt.Sprintf(`
                SELECT content, metadata 
                FROM [%s]
                WHERE id = ?
            `, v.embeddingsTableName), doc.ID).Row().Scan(&content, &metadataJSON)
			if err != nil {
				return fmt.Errorf("failed to query embeddings table for document %s: %w", doc.ID, err)
			}

			var metadata map[string]interface{}
			if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
				return fmt.Errorf("failed to parse metadata for document %s: %w", doc.ID, err)
			}

			docs[i].Content = content
			docs[i].Metadata = metadata
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return docs, nil
}

func (v *VectorStore) RemoveCollection(ctx context.Context, collection string) error {
	err := v.db.Exec(fmt.Sprintf(`DROP TABLE IF EXISTS [%s_vec]`, collection)).Error
	if err != nil {
		return fmt.Errorf("failed to drop table: %w", err)
	}

	err = v.db.Exec(fmt.Sprintf(`DELETE FROM [%s] WHERE collection_id = ?`, v.embeddingsTableName), collection).Error
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

	err := v.db.Transaction(func(tx *gorm.DB) error {
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

			err := tx.Raw(fmt.Sprintf(`
                SELECT id 
                FROM [%s]
                WHERE collection_id = ? AND %s
            `, v.embeddingsTableName, whereQuery), collection).Scan(&ids).Error
			if err != nil {
				return fmt.Errorf("failed to query IDs: %w", err)
			}
		} else {
			ids = []string{documentID}
		}

		if len(ids) == 0 {
			return nil // No documents to delete
		}

		slog.Debug("deleting documents from sqlite-vec", "ids", ids)

		if err := tx.Table(fmt.Sprintf("%s_vec", collection)).Where("document_id IN ?", ids).Delete(nil).Error; err != nil {
			return fmt.Errorf("failed to delete documents from vector table: %w", err)
		}

		if err := tx.Table(v.embeddingsTableName).Where("id IN ?", ids).Delete(nil).Error; err != nil {
			return fmt.Errorf("failed to delete documents from embeddings table: %w", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (v *VectorStore) GetDocuments(ctx context.Context, collection string, where map[string]string, whereDocument []cg.WhereDocument) ([]vs.Document, error) {
	if len(whereDocument) > 0 {
		return nil, fmt.Errorf("sqlite-vec does not support whereDocument")
	}

	var docs []vs.Document

	// Build metadata filter query
	whereQueries := []string{}
	args := []interface{}{collection}

	for k, v := range where {
		if strings.TrimSpace(k) == "" || strings.TrimSpace(v) == "" {
			continue
		}
		whereQueries = append(whereQueries, fmt.Sprintf("(metadata ->> '$.%s') = ?", k))
		args = append(args, v)
	}

	whereQuery := strings.Join(whereQueries, " AND ")
	if len(whereQuery) > 0 {
		whereQuery = " AND " + whereQuery
	}

	query := fmt.Sprintf(`
        SELECT id, content, metadata
        FROM [%s]
        WHERE collection_id = ?%s
    `, v.embeddingsTableName, whereQuery)

	rows, err := v.db.Raw(query, args...).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query documents: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var doc vs.Document
		var metadata string

		if err := rows.Scan(&doc.ID, &doc.Content, &metadata); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Parse metadata as JSON
		if err := json.Unmarshal([]byte(metadata), &doc.Metadata); err != nil {
			return nil, fmt.Errorf("failed to parse metadata for document %s: %w", doc.ID, err)
		}

		docs = append(docs, doc)
	}

	return docs, nil
}

func (v *VectorStore) ImportCollectionsFromFile(ctx context.Context, path string, collections ...string) error {
	return fmt.Errorf("not implemented")
}

func (v *VectorStore) ExportCollectionsToFile(ctx context.Context, path string, collections ...string) error {
	return fmt.Errorf("not implemented")
}
