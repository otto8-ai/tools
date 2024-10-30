package postgres

import (
	"context"
	"database/sql"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func New(ctx context.Context, dsn string, gormCfg *gorm.Config) (*gorm.DB, *sql.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), gormCfg)
	if err != nil {
		return nil, nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetConnMaxIdleTime(30 * time.Minute)
	sqlDB.SetConnMaxLifetime(1 * time.Hour)

	return db, sqlDB, nil
}
