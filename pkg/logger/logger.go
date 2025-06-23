// Package logger provides structured logging functionality using zerolog.
// It offers a simple interface for initializing and configuring application-wide logging.
package logger

import (
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Init initializes the global logger with the specified level.
// Supported levels: debug, info, warn/warning, error, fatal, panic.
// If an invalid level is provided, defaults to info level.
// The logger is configured to use console output for better readability.
func Init(level string) {
	// Configure zerolog to use console writer for better readability
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Set log level
	switch strings.ToLower(level) {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn", "warning":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "panic":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	log.Debug().Str("level", level).Msg("Logger initialized")
}

// GetLogger returns the configured logger instance.
// This logger inherits the global configuration set by Init().
// It's safe to call this function multiple times and from multiple goroutines.
func GetLogger() zerolog.Logger {
	return log.Logger
}
