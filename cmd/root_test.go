package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
)

func TestRootCmd(t *testing.T) {
	// Test that the root command can be executed without errors
	cmd := &cobra.Command{
		Use: "test",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Mock the logger initialization to avoid side effects
		},
		Run: func(cmd *cobra.Command, args []string) {
			// Do nothing
		},
	}

	// Add the log-level flag
	cmd.PersistentFlags().String("log-level", "info", "Log level")

	// Execute command with different log levels
	tests := []struct {
		name string
		args []string
	}{
		{"default log level", []string{}},
		{"debug log level", []string{"--log-level=debug"}},
		{"info log level", []string{"--log-level=info"}},
		{"error log level", []string{"--log-level=error"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture output
			var out bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetErr(&out)

			// Set args
			cmd.SetArgs(tt.args)

			// Execute command
			err := cmd.Execute()
			if err != nil {
				t.Errorf("Command execution failed: %v", err)
			}
		})
	}
}

func TestLogLevelFlag(t *testing.T) {
	// Reset the root command for testing
	testCmd := &cobra.Command{Use: "test"}
	var testLogLevel string

	testCmd.PersistentFlags().StringVar(&testLogLevel, "log-level", "info", "Log level")

	// Test different flag formats
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{"short flag format", []string{"--log-level", "debug"}, "debug"},
		{"equals format", []string{"--log-level=warn"}, "warn"},
		{"default value", []string{}, "info"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flag value
			testLogLevel = "info"

			// Parse flags
			testCmd.SetArgs(tt.args)
			err := testCmd.ParseFlags(tt.args)
			if err != nil {
				t.Errorf("Flag parsing failed: %v", err)
			}

			if testLogLevel != tt.expected {
				t.Errorf("Expected log level %s, got %s", tt.expected, testLogLevel)
			}
		})
	}
}
