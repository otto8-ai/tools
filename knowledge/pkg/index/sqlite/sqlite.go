package sqlite

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/gptscript-ai/knowledge/pkg/index/types"
	"gorm.io/gorm"
)

type Index struct {
	types.DB
}

func New(ctx context.Context, dsn string, gormCfg *gorm.Config, autoMigrate bool) (*Index, error) {
	db, err := gorm.Open(sqlite.Open(strings.TrimPrefix(dsn, "sqlite://")), gormCfg)
	if err != nil {
		return nil, err
	}

	// Enable PRAGMAs
	// - foreign key constraint to make sure that deletes cascade
	// - busy_timeout (ms) to prevent db lockups as we're accessing the DB from multiple separate processes in otto8
	// - journal_mode to WAL for better concurrency performance
	tx := db.Exec(`
PRAGMA foreign_keys = ON;
PRAGMA busy_timeout = 5000;
PRAGMA journal_mode = WAL;
`)
	if tx.Error != nil {
		return nil, tx.Error
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	sqlDB.SetConnMaxIdleTime(3 * time.Minute)

	// These only have an effect on this one process (including goroutines), but not on other processes e.g. as part
	// of parallel ingestion in otto8, so they won't improve the performance on that end
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetMaxOpenConns(1)

	return &Index{
		DB: types.DB{
			GormDB:      db,
			SqlDB:       sqlDB,
			AutoMigrate: autoMigrate,
		},
	}, nil
}

func (i *Index) AutoMigrate() error {
	return i.DB.DoAutoMigrate()
}

func (i *Index) ExportDatasetsToFile(ctx context.Context, path string, ids ...string) error {
	gdb := i.DB.GormDB.WithContext(ctx)

	var datasets []types.Dataset
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

	ndb, err := New(ctx, "sqlite://"+path, nil, true)
	if err != nil {
		return err
	}
	if err := ndb.AutoMigrate(); err != nil {
		return err
	}
	ngdb := ndb.GormDB.WithContext(ctx)

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

func (i *Index) ImportDatasetsFromFile(ctx context.Context, path string) error {
	gdb := i.DB.GormDB.WithContext(ctx)

	ndb, err := New(ctx, "sqlite://"+strings.TrimPrefix(path, "sqlite://"), nil, false)
	if err != nil {
		return err
	}
	ngdb := ndb.GormDB.WithContext(ctx)

	defer ndb.DB.Close()

	var datasets []types.Dataset
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

func (i *Index) UpdateDataset(ctx context.Context, dataset types.Dataset) error {
	return i.DB.UpdateDataset(ctx, dataset)
}

func (i *Index) CreateDataset(ctx context.Context, dataset types.Dataset, opts *types.DatasetCreateOpts) error {
	return i.DB.CreateDataset(ctx, dataset, opts)
}

func (i *Index) GetDataset(ctx context.Context, datasetID string) (*types.Dataset, error) {
	return i.DB.GetDataset(ctx, datasetID)
}

func (i *Index) ListDatasets(ctx context.Context) ([]types.Dataset, error) {
	return i.DB.ListDatasets()
}

func (i *Index) DeleteDataset(ctx context.Context, datasetID string) error {
	return i.DB.DeleteDataset(ctx, datasetID)
}

func (i *Index) DeleteFile(ctx context.Context, datasetID, fileID string) error {
	return i.DB.DeleteFile(ctx, datasetID, fileID)
}

func (i *Index) FindFile(ctx context.Context, searchFile types.File) (*types.File, error) {
	return i.DB.FindFile(ctx, searchFile)
}

func (i *Index) PruneFiles(ctx context.Context, datasetID string, pathPrefix string, keep []string) ([]types.File, error) {
	return i.DB.PruneFiles(ctx, datasetID, pathPrefix, keep)
}

func (i *Index) FindFileByMetadata(ctx context.Context, dataset string, metadata types.FileMetadata, includeDocuments bool) (*types.File, error) {
	return i.DB.FindFileByMetadata(ctx, dataset, metadata, includeDocuments)
}

func (i *Index) DeleteDocument(ctx context.Context, documentID, datasetID string) error {
	return i.DB.DeleteDocument(ctx, documentID, datasetID)
}
