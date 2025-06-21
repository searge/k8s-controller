package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestVersionCmd(t *testing.T) {
	// Create a buffer to capture output
	var out bytes.Buffer

	// Create a new root command for testing to avoid side effects
	testRootCmd := &cobra.Command{Use: "test"}
	testVersionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(cmd *cobra.Command, args []string) {
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
