// Package cmd contains tests for the CLI commands.
// This file tests the serve command definition, flag configuration, and validation logic.
package cmd

import (
	"testing"
)

// TestServeCommand verifies that the serve command is properly defined
// and configured with the expected flags and properties.
func TestServeCommand(t *testing.T) {
	if serveCmd == nil {
		t.Fatal("serveCmd should be defined")
	}

	if serveCmd.Use != "serve" {
		t.Errorf("expected command use 'serve', got %s", serveCmd.Use)
	}

	// Verify the port flag is properly configured
	portFlag := serveCmd.Flags().Lookup("port")
	if portFlag == nil {
		t.Error("expected 'port' flag to be defined")
	}
}

// TestValidatePort tests the port validation function with basic cases.
// This ensures the validation works for common valid and invalid scenarios.
func TestValidatePort(t *testing.T) {
	tests := []struct {
		name      string
		port      int
		shouldErr bool
	}{
		{"valid port 8080", 8080, false},
		{"valid port 1", 1, false},
		{"valid port 65535", 65535, false},
		{"invalid port 0", 0, true},
		{"invalid negative port", -1, true},
		{"invalid port too high", 65536, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePort(tt.port)
			if tt.shouldErr && err == nil {
				t.Errorf("validatePort(%d) should return error, got nil", tt.port)
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("validatePort(%d) should not return error, got: %v", tt.port, err)
			}
		})
	}
}
