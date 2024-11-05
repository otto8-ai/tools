package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/adrg/xdg"
	"github.com/docker/docker-credential-helpers/credentials"
	"github.com/glebarez/sqlite"
	"github.com/gptscript-ai/gptscript-helper-sqlite/pkg/common"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	s, err := NewSqlite(context.Background())
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error creating sqlite: %v\n", err)
		os.Exit(1)
	}
	credentials.Serve(s)
}

func NewSqlite(ctx context.Context) (common.Database, error) {
	var (
		dbPath string
		err    error
	)
	if os.Getenv("GPTSCRIPT_SQLITE_FILE") != "" {
		dbPath = os.Getenv("GPTSCRIPT_SQLITE_FILE")
	} else {
		dbPath, err = xdg.ConfigFile("gptscript/credentials.db")
		if err != nil {
			return common.Database{}, fmt.Errorf("failed to get credentials db path: %w", err)
		}
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
			LogLevel:                  logger.Error,
			IgnoreRecordNotFoundError: true,
		}),
	})
	if err != nil {
		return common.Database{}, fmt.Errorf("failed to open database: %w", err)
	}

	return common.NewDatabase(ctx, db)
}
