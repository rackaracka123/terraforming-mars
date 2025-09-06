package logger

import (
	"os"

	"go.uber.org/zap"
)

var globalLogger *zap.Logger

// Init initializes the global logger
func Init(logLevel *string) error {
	var err error

	// Create config based on GO_ENV for formatting
	env := os.Getenv("GO_ENV")
	var config zap.Config
	if env == "production" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
	}

	var appliedLogLevel string
	if logLevel != nil {
		appliedLogLevel = *logLevel
	} else {
		appliedLogLevel = "info"
	}

	// Set the log level based on TM_LOG_LEVEL
	switch appliedLogLevel {
	case "debug":
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		config.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		config.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	globalLogger, err = config.Build()
	if err != nil {
		return err
	}

	return nil
}

// Get returns the global logger
func Get() *zap.Logger {
	if globalLogger == nil {
		// Fallback to development logger if not initialized
		globalLogger, _ = zap.NewDevelopment()
	}
	return globalLogger
}

// Sync flushes the logger
func Sync() error {
	if globalLogger != nil {
		return globalLogger.Sync()
	}
	return nil
}

// Shutdown properly closes the logger
func Shutdown() error {
	return Sync()
}

// WithContext returns a logger with additional context fields
func WithContext(fields ...zap.Field) *zap.Logger {
	return Get().With(fields...)
}

// WithGameContext returns a logger with game-related context
func WithGameContext(gameID, playerID string) *zap.Logger {
	fields := make([]zap.Field, 0, 2)

	if gameID != "" {
		fields = append(fields, zap.String("game_id", gameID))
	}

	if playerID != "" {
		fields = append(fields, zap.String("player_id", playerID))
	}

	return Get().With(fields...)
}

// WithClientContext returns a logger with client-related context
func WithClientContext(clientID, playerID, gameID string) *zap.Logger {
	fields := make([]zap.Field, 0, 3)

	if clientID != "" {
		fields = append(fields, zap.String("client_id", clientID))
	}

	if playerID != "" {
		fields = append(fields, zap.String("player_id", playerID))
	}

	if gameID != "" {
		fields = append(fields, zap.String("game_id", gameID))
	}

	return Get().With(fields...)
}
