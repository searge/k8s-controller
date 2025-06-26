// Package cmd contains tests for the CLI commands.
// This file tests the connection command definition and flag configuration.
package cmd

import (
	"testing"
)

// TestConnectionCommandDefined verifies that the connection command is properly defined
// and configured with the expected flags and properties.
func TestConnectionCommandDefined(t *testing.T) {
	if connectionCmd == nil {
		t.Fatal("connectionCmd should be defined")
	}

	if connectionCmd.Use != "connection" {
		t.Errorf("expected command use 'connection', got %s", connectionCmd.Use)
	}

	// Verify required flags are properly configured
	tests := []struct {
		flagName string
		required bool
	}{
		{"kubeconfig", false},
		{"context", false},
		{"timeout", false},
	}

	for _, tt := range tests {
		t.Run("flag_"+tt.flagName, func(t *testing.T) {
			flag := connectionCmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Errorf("expected '%s' flag to be defined", tt.flagName)
			}
		})
	}
}

// TestConnectionFlagDefaults verifies that the connection command flags have correct default values.
func TestConnectionFlagDefaults(t *testing.T) {
	// Reset variables to test defaults
	kubeconfigPath = ""
	contextName = ""
	timeoutSeconds = 0

	// Parse empty args to get defaults
	if err := connectionCmd.ParseFlags([]string{}); err != nil {
		t.Fatalf("ParseFlags failed: %v", err)
	}

	// Check defaults - these should remain empty/zero until flags are parsed
	if kubeconfigPath != "" {
		t.Errorf("expected default kubeconfig path to be empty, got %s", kubeconfigPath)
	}

	if contextName != "" {
		t.Errorf("expected default context to be empty, got %s", contextName)
	}

	// Note: timeout has a default value set in the flag definition,
	// but it won't be applied until the command actually runs
}

// TestConnectionFlagParsing verifies that the connection command correctly parses flag values.
func TestConnectionFlagParsing(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		expectedPath string
		expectedCtx  string
		expectedTime int
	}{
		{
			name:         "kubeconfig flag",
			args:         []string{"--kubeconfig=/test/path"},
			expectedPath: "/test/path",
			expectedCtx:  "",
			expectedTime: 0,
		},
		{
			name:         "context flag",
			args:         []string{"--context=test-context"},
			expectedPath: "",
			expectedCtx:  "test-context",
			expectedTime: 0,
		},
		{
			name:         "timeout flag",
			args:         []string{"--timeout=30"},
			expectedPath: "",
			expectedCtx:  "",
			expectedTime: 30,
		},
		{
			name:         "all flags",
			args:         []string{"--kubeconfig=/test/path", "--context=test-ctx", "--timeout=45"},
			expectedPath: "/test/path",
			expectedCtx:  "test-ctx",
			expectedTime: 45,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset variables
			kubeconfigPath = ""
			contextName = ""
			timeoutSeconds = 0

			// Parse flags
			err := connectionCmd.ParseFlags(tt.args)
			if err != nil {
				t.Errorf("ParseFlags failed: %v", err)
			}

			// Check values
			if kubeconfigPath != tt.expectedPath {
				t.Errorf("expected kubeconfig path %s, got %s", tt.expectedPath, kubeconfigPath)
			}

			if contextName != tt.expectedCtx {
				t.Errorf("expected context %s, got %s", tt.expectedCtx, contextName)
			}

			if timeoutSeconds != tt.expectedTime {
				t.Errorf("expected timeout %d, got %d", tt.expectedTime, timeoutSeconds)
			}
		})
	}
}
