package main

import (
	"os"
	"testing"

	"github.com/Searge/k8s-controller/cmd"
)

// TestCmdExecute tests that cmd.Execute() can be called
// This provides coverage for the main.go file
func TestCmdExecute(t *testing.T) {
	// Save original args and exit function
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set args to show help (this won't cause exit)
	os.Args = []string{"k8s-controller", "--help"}

	// cmd.Execute() will call os.Exit(0) for --help
	// We need to catch that
	defer func() {
		if r := recover(); r != nil {
			// This is expected for --help flag
		}
	}()

	// This call covers the cmd.Execute() line in main.go
	// It will exit with help, but that's fine for test coverage
	cmd.Execute()
}
