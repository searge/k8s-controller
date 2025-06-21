package logger

import (
	"bytes"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

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

func BenchmarkInit(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Init("info")
	}
}

func BenchmarkGetLogger(b *testing.B) {
	Init("info")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		GetLogger()
	}
}
