package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/gptscript-ai/knowledge/pkg/index/types"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Index struct {
	types.DB
}

func New(ctx context.Context, dsn string, gormCfg *gorm.Config, autoMigrate bool) (*Index, error) {
	db, err := gorm.Open(postgres.Open(dsn), gormCfg)
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetConnMaxIdleTime(30 * time.Minute)
	sqlDB.SetConnMaxLifetime(1 * time.Hour)

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
	return fmt.Errorf("postgres: ExportDatasetsToFile not implemented")
}

func (i *Index) ImportDatasetsFromFile(ctx context.Context, path string) error {
	return fmt.Errorf("postgres: ImportDatasetsFromFile not implemented")
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
