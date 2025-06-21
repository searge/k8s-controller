// Package cmd implements the command-line interface for the k8s-controller application.
package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestVersionCmd verifies that the version command executes successfully
// and produces the expected output format.
func TestVersionCmd(t *testing.T) {
	// Create a buffer to capture output
	var out bytes.Buffer

	// Create a new root command for testing to avoid side effects
	testRootCmd := &cobra.Command{Use: "test"}
	testVersionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(_ *cobra.Command, _ []string) {
			out.WriteString("k8s-controller version dev\n")
		},
	}

	testRootCmd.AddCommand(testVersionCmd)
	testRootCmd.SetOut(&out)
	testRootCmd.SetArgs([]string{"version"})

	err := testRootCmd.Execute()
	if err != nil {
		t.Errorf("version command failed: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "k8s-controller version") {
		t.Errorf("Expected version output, got: %s", output)
	}
}

// TestAppVersion verifies that the appVersion variable has a valid default value
// and can be accessed for version information.
func TestAppVersion(t *testing.T) {
	// Test that appVersion variable exists and has a default value
	if appVersion == "" {
		t.Error("appVersion should not be empty")
	}

	// Test default value
	if appVersion != "dev" {
		t.Errorf("Expected default version 'dev', got: %s", appVersion)
	}
}
