// Package logger provides structured logging functionality using zerolog.
package logger

import (
	"bytes"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// TestInit verifies that the Init function correctly sets the global log level
// for various input values including valid levels, invalid levels, and edge cases.
func TestInit(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		expected zerolog.Level
	}{
		{"debug level", "debug", zerolog.DebugLevel},
		{"info level", "info", zerolog.InfoLevel},
		{"warn level", "warn", zerolog.WarnLevel},
		{"warning level", "warning", zerolog.WarnLevel},
		{"error level", "error", zerolog.ErrorLevel},
		{"fatal level", "fatal", zerolog.FatalLevel},
		{"panic level", "panic", zerolog.PanicLevel},
		{"invalid level defaults to info", "invalid", zerolog.InfoLevel},
		{"empty level defaults to info", "", zerolog.InfoLevel},
		{"uppercase level", "DEBUG", zerolog.DebugLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture log output
			var buf bytes.Buffer
			log.Logger = log.Output(&buf)

			// Test the Init function
			Init(tt.level)

			// Check if the global level was set correctly
			if zerolog.GlobalLevel() != tt.expected {
				t.Errorf("Init(%s) set level to %v, want %v",
					tt.level, zerolog.GlobalLevel(), tt.expected)
			}
		})
	}
}

// TestGetLogger verifies that GetLogger returns a valid logger instance
// and that the returned logger can be used for logging without panicking.
func TestGetLogger(t *testing.T) {
	// Initialize logger
	Init("info")

	// Get logger instance
	logger := GetLogger()

	// Test that we can log without panicking
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("GetLogger() caused panic: %v", r)
		}
	}()

	logger.Info().Msg("test message")
}

// BenchmarkInit measures the performance of the Init function.
// This helps ensure that logger initialization doesn't become a bottleneck.
func BenchmarkInit(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Init("info")
	}
}

// BenchmarkGetLogger measures the performance of the GetLogger function.
// This is important since GetLogger might be called frequently throughout the application.
func BenchmarkGetLogger(b *testing.B) {
	Init("info")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		GetLogger()
	}
}

// ExampleInit demonstrates basic usage of the Init function
// with different log levels.
func ExampleInit() {
	// Initialize logger with info level
	Init("info")

	// Initialize logger with debug level for development
	Init("debug")

	// Initialize logger with error level for production
	Init("error")

	// Output:
}

// ExampleGetLogger demonstrates how to get and use a logger instance.
func ExampleGetLogger() {
	// First initialize the logger
	Init("info")

	// Get a logger instance
	logger := GetLogger()

	// Use the logger
	logger.Info().Str("component", "example").Msg("Application started")
	logger.Debug().Int("count", 42).Msg("Processing items")

	// Output:
}

// ExampleInit_withInvalidLevel demonstrates that invalid log levels
// default to info level gracefully.
func ExampleInit_withInvalidLevel() {
	// Invalid levels default to info
	Init("invalid-level")

	logger := GetLogger()
	logger.Info().Msg("This will be logged at info level")

	// Output:
}
