// Package cmd contains tests for the CLI commands.
// This file tests the version command.
//
// It used to hold a full copy of the serve command's tests under different
// names; those live in serve_test.go, where the command they test lives.
package cmd

import (
	"testing"
)

// TestVersionCommandDefined verifies the version command is registered and the
// build-time default is in place.
func TestVersionCommandDefined(t *testing.T) {
	if versionCmd == nil {
		t.Fatal("versionCmd should be defined")
	}

	if versionCmd.Use != "version" {
		t.Errorf("expected command use 'version', got %s", versionCmd.Use)
	}

	if Version == "" {
		t.Error("Version must not be empty: the ldflags override depends on a default being present")
	}

	if rootCmd.Version != Version {
		t.Errorf("rootCmd.Version = %q, want %q: --version and the version command should agree", rootCmd.Version, Version)
	}
}
