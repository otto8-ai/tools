package log

import (
	"context"
	"log/slog"
)

type contextKey string

const loggerKey = contextKey("logger")

func ToCtx(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

func FromCtx(ctx context.Context) *slog.Logger {
	v := ctx.Value(loggerKey)
	if v == nil {
		return slog.Default()
	}
	return v.(*slog.Logger)
}
