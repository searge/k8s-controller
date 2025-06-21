// Package cmd contains tests for the CLI commands.
// This file tests the serve command definition and flag configuration.
package cmd

import (
	"testing"
)

// TestServeCommandDefined verifies that the serve command is properly defined
// and configured with the expected flags and properties.
func TestServeCommandDefined(t *testing.T) {
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
