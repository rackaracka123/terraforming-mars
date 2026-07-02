package logger

import (
	"log/slog"
	"os"
)

// globalLevel holds the active log level; adjusting it reconfigures the default
// logger without rebuilding the handler.
var globalLevel = new(slog.LevelVar)

// Init initializes the process-wide slog default logger. In production it emits
// JSON; otherwise it uses the pretty colored console handler. The whole backend
// — domain, action, delivery, service — logs through slog.Default().
func Init(logLevel *string) error {
	appliedLogLevel := "info"
	if logLevel != nil {
		appliedLogLevel = *logLevel
	}

	switch appliedLogLevel {
	case "debug":
		globalLevel.Set(slog.LevelDebug)
	case "warn":
		globalLevel.Set(slog.LevelWarn)
	case "error":
		globalLevel.Set(slog.LevelError)
	default:
		globalLevel.Set(slog.LevelInfo)
	}

	var handler slog.Handler
	if os.Getenv("GO_ENV") == "production" {
		handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: globalLevel, AddSource: true})
	} else {
		handler = newPrettyHandler(os.Stderr, globalLevel)
	}

	slog.SetDefault(slog.New(handler))
	return nil
}

// Get returns the process-wide logger. Before Init runs it falls back to the
// slog default (a plain text handler), so callers never get a nil logger.
func Get() *slog.Logger {
	return slog.Default()
}

// Sync is a no-op: slog handlers write synchronously to stderr. Retained so the
// server shutdown path and tests compile unchanged.
func Sync() error {
	return nil
}

// Shutdown is a no-op; see Sync.
func Shutdown() error {
	return nil
}

// WithContext returns a logger with additional context attributes.
func WithContext(args ...any) *slog.Logger {
	return slog.Default().With(args...)
}

// WithGameContext returns a logger with game-related context.
func WithGameContext(gameID, playerID string) *slog.Logger {
	args := make([]any, 0, 2)
	if gameID != "" {
		args = append(args, slog.String("game_id", gameID))
	}
	if playerID != "" {
		args = append(args, slog.String("player_id", playerID))
	}
	return slog.Default().With(args...)
}

// WithClientContext returns a logger with client-related context.
func WithClientContext(clientID, playerID, gameID string) *slog.Logger {
	args := make([]any, 0, 3)
	if clientID != "" {
		args = append(args, slog.String("client_id", clientID))
	}
	if playerID != "" {
		args = append(args, slog.String("player_id", playerID))
	}
	if gameID != "" {
		args = append(args, slog.String("game_id", gameID))
	}
	return slog.Default().With(args...)
}
