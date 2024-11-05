package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/docker/docker-credential-helpers/credentials"
	"github.com/gptscript-ai/gptscript-helper-sqlite/pkg/common"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	p, err := NewPostgres(context.Background())
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error creating postgres: %v\n", err)
		os.Exit(1)
	}
	credentials.Serve(p)
}

func NewPostgres(ctx context.Context) (common.Database, error) {
	dsn := os.Getenv("GPTSCRIPT_POSTGRES_DSN")
	if dsn == "" {
		return common.Database{}, fmt.Errorf("missing GPTSCRIPT_POSTGRES_DSN")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
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
