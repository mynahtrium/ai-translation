package logging

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

type ctxKey struct{}

var sessionKey = ctxKey{}

func New(level string) *slog.Logger {
	var lvl slog.Level
	switch strings.ToLower(level) {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level:     lvl,
		AddSource: lvl == slog.LevelDebug,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	return slog.New(handler)
}

func WithSession(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, sessionKey, sessionID)
}

func FromContext(ctx context.Context, logger *slog.Logger) *slog.Logger {
	if sessionID, ok := ctx.Value(sessionKey).(string); ok {
		return logger.With("session_id", sessionID)
	}
	return logger
}
