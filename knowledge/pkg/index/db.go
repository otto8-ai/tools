package index

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gptscript-ai/knowledge/pkg/index/postgres"
	"github.com/gptscript-ai/knowledge/pkg/index/sqlite"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DB struct {
	gormDB      *gorm.DB
	sqlDB       *sql.DB
	autoMigrate bool
}

func New(ctx context.Context, dsn string, autoMigrate bool) (*DB, error) {
	var (
		db      *gorm.DB
		sqlDB   *sql.DB
		err     error
		gormCfg = &gorm.Config{
			Logger: logger.New(log.Default(), logger.Config{
				SlowThreshold: 200 * time.Millisecond,
				Colorful:      true,
				LogLevel:      logger.Silent,
			}),
		}
	)

	dialect := strings.Split(dsn, "://")[0]

	slog.Debug("indexdb", "dialect", dialect, "dsn", dsn)

	switch dialect {
	case "sqlite":
		db, sqlDB, err = sqlite.New(ctx, dsn, gormCfg)
	case "postgres":
		db, sqlDB, err = postgres.New(ctx, dsn, gormCfg)
	default:
		err = fmt.Errorf("unsupported dialect: %q", dialect)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to open index DB: %w", err)
	}

	return &DB{
		gormDB:      db,
		sqlDB:       sqlDB,
		autoMigrate: autoMigrate,
	}, nil
}

func (db *DB) AutoMigrate() error {
	if !db.autoMigrate {
		return nil
	}

	return db.gormDB.AutoMigrate(
		&Dataset{},
		&File{},
		&Document{},
	)
}

func (db *DB) Check(w http.ResponseWriter, _ *http.Request) {
	if err := db.sqlDB.Ping(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	_, _ = w.Write([]byte(`{"status": "ok"}`))
}

func (db *DB) Close() error {
	return db.sqlDB.Close()
}

func (db *DB) WithContext(ctx context.Context) *gorm.DB {
	return db.gormDB.WithContext(ctx)
}
