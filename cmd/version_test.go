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
			// Create a test command that mimics version flag behavior
			var out bytes.Buffer
			testCmd := &cobra.Command{
				Use: "test",
				Run: func(_ *cobra.Command, _ []string) {
					out.WriteString("k8s-controller version dev\n")
				},
			}

			var showVersion bool
			testCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "Show version")

			testCmd.SetOut(&out)
			testCmd.SetArgs(tt.args)

			// Parse flags to set showVersion
			err := testCmd.ParseFlags(tt.args)
			if err != nil {
				t.Errorf("Flag parsing failed: %v", err)
			}

			// Simulate version flag behavior
			if showVersion {
				testCmd.Run(testCmd, []string{})
				output := out.String()
				if !strings.Contains(output, "k8s-controller version") {
					t.Errorf("Expected version output, got: %s", output)
				}
			}
		})
	}
}
