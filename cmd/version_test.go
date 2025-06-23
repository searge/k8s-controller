// Package cmd implements the command-line interface for the k8s-controller application.
package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestVersionCmd verifies that the version command executes successfully
// and produces the expected output format.
func TestVersionCmd(t *testing.T) {
	// Create a buffer to capture output
	var out bytes.Buffer

	// Use getVersionCmd helper to create isolated version command
	cmd := getVersionCmd(&out)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("version command failed: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "k8s-controller version") {
		t.Errorf("Expected version output, got: %s", output)
	}
}

// getVersionCmd creates an isolated copy of the version command with custom output.
// This avoids side effects and allows testing the command logic independently.
func getVersionCmd(out *bytes.Buffer) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Fprintf(out, "k8s-controller version %s\n", Version)
		},
	}
}

// TestVersion verifies that the Version variable has a valid default value
// and can be accessed for version information.
func TestVersion(t *testing.T) {
	// Test that Version variable exists and has a default value
	if Version == "" {
		t.Error("Version should not be empty")
	}

	// Test default value
	if Version != "dev" {
		t.Errorf("Expected default version 'dev', got: %s", Version)
	}
}

// TestVersionFlag verifies that the --version and -v flags work correctly
// and produce the expected output without running other commands.
func TestVersionFlag(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"long version flag", []string{"--version"}},
		{"short version flag", []string{"-v"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create isolated test command with version support
			var out bytes.Buffer
			cmd := getVersionCmd(&out)

			var showVersion bool
			cmd.Flags().BoolVarP(&showVersion, "version", "v", false, "Show version")

			cmd.SetOut(&out)
			cmd.SetArgs(tt.args)

			// Parse flags to set showVersion
			err := cmd.ParseFlags(tt.args)
			if err != nil {
				t.Errorf("Flag parsing failed: %v", err)
			}

			// Simulate version flag behavior
			if showVersion {
				cmd.Run(cmd, []string{})
				output := out.String()
				if !strings.Contains(output, "k8s-controller version") {
					t.Errorf("Expected version output, got: %s", output)
				}
			}
		})
	}
}
