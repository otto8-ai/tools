package index

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"gorm.io/gorm"
)

func GetDataset(db *gorm.DB, id string) (*Dataset, error) {
	var dataset Dataset
	err := db.First(&dataset, "id = ?", id).Error
	return &dataset, err
}

// TODO: Index should become an interface to support this for different databases

func (db *DB) ExportDatasetsToFile(ctx context.Context, path string, ids ...string) error {
	gdb := db.gormDB.WithContext(ctx)

	var datasets []Dataset
	err := gdb.Preload("Files.Documents").Find(&datasets, "id IN ?", ids).Error
	if err != nil {
		return err
	}

	slog.Debug("Exporting datasets", "ids", ids, "count", len(datasets))

	finfo, err := os.Stat(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if finfo.IsDir() {
		path = filepath.Join(path, "knowledge-export.db")
	}

	slog.Debug("Exporting datasets to file", "path", path)

	ndb, err := New(ctx, "sqlite://"+path, true)
	if err != nil {
		return err
	}
	if err := ndb.AutoMigrate(); err != nil {
		return err
	}
	ngdb := ndb.gormDB.WithContext(ctx)

	defer ndb.Close()

	// fill new database with exported datasets
	for _, dataset := range datasets {
		if err := ngdb.Create(&dataset).Error; err != nil {
			return err
		}
	}
	ngdb.Commit()

	return nil
}

func (db *DB) ImportDatasetsFromFile(ctx context.Context, path string) error {
	gdb := db.gormDB.WithContext(ctx)

	ndb, err := New(ctx, "sqlite://"+strings.TrimPrefix(path, "sqlite://"), false)
	if err != nil {
		return err
	}
	ngdb := ndb.gormDB.WithContext(ctx)

	defer ndb.Close()

	var datasets []Dataset
	err = ngdb.Find(&datasets).Error
	if err != nil {
		return err
	}

	// fill new database with exported datasets
	for _, dataset := range datasets {
		if err := gdb.Create(&dataset).Error; err != nil {
			return err
		}
	}
	gdb.Commit()

	return nil
}

func (db *DB) UpdateDataset(ctx context.Context, dataset Dataset) error {
	gdb := db.gormDB.WithContext(ctx)

	slog.Debug("Updating dataset in DB", "id", dataset.ID, "metadata", dataset.Metadata)
	err := gdb.Save(dataset).Error
	if err != nil {
		return err
	}

	gdb.Commit()
	return nil
}
