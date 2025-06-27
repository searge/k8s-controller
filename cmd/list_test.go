// Package cmd contains tests for the CLI commands.
// This file tests the list command definition, flag configuration, and validation logic.
package cmd

import (
	"testing"
)

// TestListCommandDefined verifies that the list command is properly defined
// and configured with the expected properties.
func TestListCommandDefined(t *testing.T) {
	if listCmd == nil {
		t.Fatal("listCmd should be defined")
	}

	if listCmd.Use != "list" {
		t.Errorf("expected command use 'list', got %s", listCmd.Use)
	}

	// Verify that the deployments subcommand is registered
	deploymentsCmdFound := false
	for _, subCmd := range listCmd.Commands() {
		if subCmd.Use == "deployments" {
			deploymentsCmdFound = true
			break
		}
	}

	if !deploymentsCmdFound {
		t.Error("deployments subcommand should be registered with list command")
	}
}

// TestListDeploymentsCommandDefined verifies that the list deployments command
// is properly defined and configured with the expected flags.
func TestListDeploymentsCommandDefined(t *testing.T) {
	if listDeploymentsCmd == nil {
		t.Fatal("listDeploymentsCmd should be defined")
	}

	if listDeploymentsCmd.Use != "deployments" {
		t.Errorf("expected command use 'deployments', got %s", listDeploymentsCmd.Use)
	}

	// Verify required flags are properly configured
	tests := []struct {
		flagName    string
		shorthand   string
		shouldExist bool
	}{
		{"namespace", "n", true},
		{"output", "o", true},
	}

	for _, tt := range tests {
		t.Run("flag_"+tt.flagName, func(t *testing.T) {
			flag := listDeploymentsCmd.Flags().Lookup(tt.flagName)
			if tt.shouldExist && flag == nil {
				t.Errorf("expected '%s' flag to be defined", tt.flagName)
			}
			if !tt.shouldExist && flag != nil {
				t.Errorf("expected '%s' flag not to be defined", tt.flagName)
			}

			// Check shorthand if flag exists
			if tt.shouldExist && flag != nil && flag.Shorthand != tt.shorthand {
				t.Errorf("expected '%s' flag shorthand to be '%s', got '%s'",
					tt.flagName, tt.shorthand, flag.Shorthand)
			}
		})
	}
}

// TestListDeploymentsFlagParsing verifies that the list deployments command
// correctly parses flag values.
func TestListDeploymentsFlagParsing(t *testing.T) {
	tests := createFlagParsingTestCases()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runFlagParsingTest(t, tt.args, tt.expectedNamespace, tt.expectedOutput, tt.shouldErr)
		})
	}
}

// createFlagParsingTestCases creates test cases for flag parsing to reduce function length.
func createFlagParsingTestCases() []struct {
	name              string
	args              []string
	expectedNamespace string
	expectedOutput    string
	shouldErr         bool
} {
	// Define test constants to avoid duplication
	const (
		defaultNS    = "default"
		kubeSystemNS = "kube-system"
		tableFormat  = "table"
		jsonFormat   = "json"
	)

	return []struct {
		name              string
		args              []string
		expectedNamespace string
		expectedOutput    string
		shouldErr         bool
	}{
		{
			name:              "default values",
			args:              []string{},
			expectedNamespace: "",
			expectedOutput:    tableFormat,
			shouldErr:         false,
		},
		{
			name:              "namespace flag",
			args:              []string{"--namespace=" + defaultNS},
			expectedNamespace: defaultNS,
			expectedOutput:    tableFormat,
			shouldErr:         false,
		},
		{
			name:              "namespace short flag",
			args:              []string{"-n", kubeSystemNS},
			expectedNamespace: kubeSystemNS,
			expectedOutput:    tableFormat,
			shouldErr:         false,
		},
		{
			name:              "output flag",
			args:              []string{"--output=" + jsonFormat},
			expectedNamespace: "",
			expectedOutput:    jsonFormat,
			shouldErr:         false,
		},
		{
			name:              "output short flag",
			args:              []string{"-o", jsonFormat},
			expectedNamespace: "",
			expectedOutput:    jsonFormat,
			shouldErr:         false,
		},
		{
			name:              "both flags",
			args:              []string{"-n", defaultNS, "-o", jsonFormat},
			expectedNamespace: defaultNS,
			expectedOutput:    jsonFormat,
			shouldErr:         false,
		},
	}
}

// runFlagParsingTest is a helper function to reduce cognitive complexity.
func runFlagParsingTest(t *testing.T, args []string, expectedNamespace, expectedOutput string, shouldErr bool) {
	t.Helper()

	// Reset variables
	namespace = ""
	outputFormat = "table"

	// Parse flags
	err := listDeploymentsCmd.ParseFlags(args)
	if shouldErr && err == nil {
		t.Error("expected error but got none")
	}
	if !shouldErr && err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check values if no error expected
	if !shouldErr {
		if namespace != expectedNamespace {
			t.Errorf("expected namespace %s, got %s", expectedNamespace, namespace)
		}
		if outputFormat != expectedOutput {
			t.Errorf("expected output %s, got %s", expectedOutput, outputFormat)
		}
	}
}

// TestValidateOutputFormat tests the output format validation function.
func TestValidateOutputFormat(t *testing.T) {
	tests := []struct {
		name      string
		format    string
		shouldErr bool
	}{
		{"valid table format", "table", false},
		{"valid json format", "json", false},
		{"invalid format", "yaml", true},
		{"invalid format xml", "xml", true},
		{"empty format", "", true},
		{"case sensitive", "Table", true}, // Should be lowercase
		{"case sensitive json", "JSON", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateOutputFormat(tt.format)
			if tt.shouldErr && err == nil {
				t.Errorf("validateOutputFormat(%s) should return error, got nil", tt.format)
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("validateOutputFormat(%s) should not return error, got: %v", tt.format, err)
			}
		})
	}
}

// TestValidateNamespace tests the namespace validation function.
func TestValidateNamespace(t *testing.T) {
	const kubeSystemNS = "kube-system" // Define constant to avoid duplication

	tests := []struct {
		name      string
		namespace string
		shouldErr bool
	}{
		{"empty namespace", "", false}, // Empty means all namespaces
		{"valid namespace", "default", false},
		{"valid namespace with hyphen", kubeSystemNS, false},
		{"valid namespace with numbers", "test123", false},
		{"valid namespace with mixed", "app-v2", false},
		{
			"too long namespace",
			"this-is-a-very-long-namespace-name-that-exceeds-the-maximum-length-allowed-by-kubernetes",
			true,
		},
		{"uppercase characters", "MyNamespace", true},
		{"contains underscore", "my_namespace", true},
		{"contains dot", "my.namespace", true},
		{"starts with hyphen", "-myns", true},
		{"ends with hyphen", "myns-", true},
		{"contains space", "my namespace", true},
		{"special characters", "my@namespace", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateNamespace(tt.namespace)
			if tt.shouldErr && err == nil {
				t.Errorf("validateNamespace(%s) should return error, got nil", tt.namespace)
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("validateNamespace(%s) should not return error, got: %v", tt.namespace, err)
			}
		})
	}
}
