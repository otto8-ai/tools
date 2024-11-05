package index

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"strings"
	"time"

	"github.com/gptscript-ai/knowledge/pkg/index/postgres"
	"github.com/gptscript-ai/knowledge/pkg/index/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func New(ctx context.Context, dsn string, autoMigrate bool) (Index, error) {
	var (
		indexDB Index
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
		indexDB, err = sqlite.New(ctx, dsn, gormCfg, autoMigrate)
	case "postgres":
		indexDB, err = postgres.New(ctx, dsn, gormCfg, autoMigrate)
	default:
		err = fmt.Errorf("unsupported dialect: %q", dialect)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to open index DB: %w", err)
	}

	return indexDB, nil
}
