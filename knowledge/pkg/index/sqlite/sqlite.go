package sqlite

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func New(ctx context.Context, dsn string, gormCfg *gorm.Config) (*gorm.DB, *sql.DB, error) {
	db, err := gorm.Open(sqlite.Open(strings.TrimPrefix(dsn, "sqlite://")), gormCfg)
	if err != nil {
		return nil, nil, err
	}

	// Enable foreign key constraint to make sure that deletes cascade
	tx := db.Exec("PRAGMA foreign_keys = ON")
	if tx.Error != nil {
		return nil, nil, tx.Error
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, err
	}

	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	sqlDB.SetConnMaxIdleTime(3 * time.Minute)
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetMaxOpenConns(1)

	return db, sqlDB, nil
}
