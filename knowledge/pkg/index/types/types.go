package types

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"gorm.io/gorm"
)

type DB struct {
	GormDB      *gorm.DB
	SqlDB       *sql.DB
	AutoMigrate bool
}

func (db *DB) DoAutoMigrate() error {
	if !db.AutoMigrate {
		return nil
	}

	return db.GormDB.AutoMigrate(
		&Dataset{},
		&File{},
		&Document{},
	)
}

func (db *DB) UpdateDataset(ctx context.Context, dataset Dataset) error {
	gdb := db.GormDB.WithContext(ctx)

	slog.Debug("Updating dataset in DB", "id", dataset.ID, "metadata", dataset.Metadata)
	err := gdb.Save(dataset).Error
	if err != nil {
		return err
	}

	gdb.Commit()
	return nil
}

func (db *DB) Close() error {
	return db.SqlDB.Close()
}

func (db *DB) WithContext(ctx context.Context) *gorm.DB {
	return db.GormDB.WithContext(ctx)
}

func (db *DB) CreateDataset(ctx context.Context, dataset Dataset, opts *DatasetCreateOpts) error {
	gdb := db.GormDB.WithContext(ctx)

	slog.Debug("Creating dataset in DB", "id", dataset.ID, "metadata", dataset.Metadata)
	err := gdb.Create(&dataset).Error
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			if opts != nil && opts.ErrOnExists {
				return fmt.Errorf("dataset already exists: %w", err)
			}
		} else {
			return err
		}
	} else {
		gdb.Commit()
	}
	return nil
}

func (db *DB) DeleteDataset(ctx context.Context, datasetID string) error {
	gdb := db.GormDB.WithContext(ctx)

	slog.Debug("Deleting dataset from DB", "id", datasetID)
	tx := gdb.Delete(&Dataset{}, "id = ?", datasetID)
	if tx.Error != nil {
		return tx.Error
	}
	tx.Commit()

	return nil
}

func (db *DB) GetDataset(ctx context.Context, datasetID string) (*Dataset, error) {
	dataset := &Dataset{}
	tx := db.WithContext(ctx).Preload("Files.Documents").First(dataset, "id = ?", datasetID)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get dataset %q from DB: %w", datasetID, tx.Error)
	}

	return dataset, nil
}

func (db *DB) ListDatasets() ([]Dataset, error) {
	var datasets []Dataset
	tx := db.GormDB.Find(&datasets)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return datasets, nil
}

func (db *DB) DeleteFile(ctx context.Context, datasetID, fileID string) error {

	// Find file in database with associated documents
	var file File
	tx := db.WithContext(ctx).Preload("Documents").Where("id = ? AND dataset = ?", fileID, datasetID).First(&file)
	if tx.Error != nil {
		return ErrDBFileNotFound
	}

	// Remove owned documents from VectorStore and Database
	for _, doc := range file.Documents {

		tx = db.WithContext(ctx).Delete(&doc)
		if tx.Error != nil {
			return fmt.Errorf("failed to delete document from DB: %w", tx.Error)
		}
	}

	// Remove file DB
	tx = db.WithContext(ctx).Delete(&file)
	if tx.Error != nil {
		return fmt.Errorf("failed to delete file from DB: %w", tx.Error)
	}

	return nil
}

func (db *DB) PruneFiles(ctx context.Context, datasetID string, pathPrefix string, keep []string) ([]File, error) {
	var files []File
	tx := db.WithContext(ctx).
		Where("dataset = ?", datasetID).
		Where("absolute_path LIKE ?", pathPrefix+"%").
		Not("absolute_path IN ?", keep).
		Find(&files)
	if tx.Error != nil {
		return nil, tx.Error
	}

	slog.Debug("Pruning files", "count", len(files), "dataset", datasetID, "path_prefix", pathPrefix, "keep", keep)

	for _, file := range files {
		if err := db.DeleteFile(ctx, datasetID, file.ID); err != nil {
			return nil, err
		}
	}

	return files, nil
}

func (db *DB) FindFile(ctx context.Context, searchFile File) (*File, error) {
	if searchFile.Dataset == "" {
		return nil, fmt.Errorf("dataset must be provided")
	}

	var file File
	var tx *gorm.DB
	if searchFile.ID != "" {
		tx = db.WithContext(ctx).Preload("Documents").Where("dataset = ? AND id = ?", searchFile.Dataset, searchFile.ID).First(&file)
	} else if searchFile.AbsolutePath != "" {
		tx = db.WithContext(ctx).Preload("Documents").Where("dataset = ? AND absolute_path = ?", searchFile.Dataset, searchFile.AbsolutePath).First(&file)
	} else {
		return nil, fmt.Errorf("either fileID or fileAbsPath must be provided")
	}
	if tx.Error != nil {
		return nil, ErrDBFileNotFound
	}

	return &file, nil
}

func (db *DB) FindFileByMetadata(ctx context.Context, dataset string, metadata FileMetadata, includeDocuments bool) (*File, error) {
	var file File
	tx := db.WithContext(ctx)
	if includeDocuments {
		tx = tx.Preload("Documents")
	}
	tx = tx.Where("dataset = ?", dataset)

	if metadata.Name != "" {
		tx = tx.Where("name = ?", metadata.Name)
	}
	if metadata.AbsolutePath != "" {
		tx = tx.Where("absolute_path = ?", metadata.AbsolutePath)
	}
	if metadata.Size > 0 {
		tx = tx.Where("size = ?", metadata.Size)
	}
	if !metadata.ModifiedAt.IsZero() {
		tx = tx.Where("modified_at = ?", metadata.ModifiedAt)
	}

	err := tx.First(&file).Error
	if err != nil {
		return nil, ErrDBFileNotFound
	}

	return &file, nil
}

func (db *DB) DeleteDocument(ctx context.Context, documentID, datasetID string) error {
	// Find in Database
	var document Document
	tx := db.WithContext(ctx).First(&document, "id = ? AND dataset = ?", documentID, datasetID)
	if tx.Error != nil {
		return ErrDBDocumentNotFound
	}

	// Remove from Database
	tx = db.WithContext(ctx).Delete(&document)
	if tx.Error != nil {
		return fmt.Errorf("failed to delete document from DB: %w", tx.Error)
	}

	// Check if owning file should be removed
	var count int64
	tx = db.WithContext(ctx).Model(&Document{}).Where("file_id = ?", document.FileID).Count(&count)
	if tx.Error != nil {
		return tx.Error
	}

	if count == 0 {
		slog.Info("Removing file, because all associated documents are gone", "file", document.FileID)
		tx = db.WithContext(ctx).Delete(&File{}, "id = ?", document.FileID)
		if tx.Error != nil {
			return fmt.Errorf("failed to delete owning file from DB: %w", tx.Error)
		}
	}

	return nil
}

func (db *DB) CreateFile(ctx context.Context, file File) error {
	gdb := db.GormDB.WithContext(ctx)

	slog.Debug("Creating file in DB", "id", file.ID, "metadata", file.FileMetadata)
	err := gdb.Create(&file).Error
	if err != nil {
		return err
	}

	gdb.Commit()
	return nil
}
