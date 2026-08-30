// Package logger provides structured logging functionality using zerolog.
// It offers a simple interface for initializing and configuring application-wide logging.
package logger

import (
	"io"
	"os"
	"strings"

	"github.com/go-logr/zerologr"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"k8s.io/klog/v2"
)

// Init initializes the global logger with the specified level.
// Supported levels: debug, info, warn/warning, error, fatal, panic.
// If an invalid level is provided, defaults to info level.
// The logger is configured to use console output for better readability.
func Init(level string) {
	InitWithWriter(level, os.Stderr)
}

// InitWithWriter is Init with the output destination as a parameter, which is
// what lets tests capture the log — including the klog bridge below, since
// zerologr snapshots the logger value when the bridge is installed and a later
// swap of log.Logger does not re-route klog.
func InitWithWriter(level string, out io.Writer) {
	// Configure zerolog to use console writer for better readability
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: out})

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

	// Route client-go's logging through zerolog. client-go reports its runtime
	// failures -- a reflector that cannot reach the API server, a cache that will
	// not sync -- via klog, not via anything this application passes it. Without
	// this bridge those errors bypass structured logging entirely: a smoke test
	// against an unreachable cluster produced sixty seconds of silent retries
	// and a single raw klog line, invisible to anything parsing our output.
	klog.SetLogger(zerologr.New(&log.Logger))

	log.Debug().Str("level", level).Msg("Logger initialized")
}

// GetLogger returns the configured logger instance.
// This logger inherits the global configuration set by Init().
// It's safe to call this function multiple times and from multiple goroutines.
func GetLogger() zerolog.Logger {
	return log.Logger
}
