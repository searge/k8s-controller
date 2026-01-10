// Package cmd contains tests for the CLI commands.
// This file tests the list command definition, flag configuration, and validation logic.
package cmd

import (
	"encoding/json"
	"io"
	"os"
	"testing"
	"time"

	"github.com/Searge/k8s-controller/pkg/k8s"
)

// Test constants to avoid string duplication
const (
	testImageNginx        = "nginx:1.21"
	testImageRedis        = "redis:6.2"
	testImagePostgres     = "postgres:13"
	testImageBusybox      = "busybox:latest"
	testDeploymentName    = "test-deployment"
	testNamespaceDefault  = "default"
	testNamespaceKube     = "kube-system"
	testMessageCloseError = "Failed to close client"
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
		{"selector", "l", true},
		{"kubeconfig", "", true},
		{"context", "", true},
		{"timeout", "", true},
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

			// Check shorthand if flag exists and expected
			if tt.shouldExist && flag != nil && tt.shorthand != "" && flag.Shorthand != tt.shorthand {
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
		tableFormat = "table"
		jsonFormat  = "json"
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
			args:              []string{"--namespace=" + testNamespaceDefault},
			expectedNamespace: testNamespaceDefault,
			expectedOutput:    tableFormat,
			shouldErr:         false,
		},
		{
			name:              "namespace short flag",
			args:              []string{"-n", testNamespaceKube},
			expectedNamespace: testNamespaceKube,
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
			args:              []string{"-n", testNamespaceDefault, "-o", jsonFormat},
			expectedNamespace: testNamespaceDefault,
			expectedOutput:    jsonFormat,
			shouldErr:         false,
		},
		{
			name:              "label selector flag",
			args:              []string{"-l", "app=nginx"},
			expectedNamespace: "",
			expectedOutput:    tableFormat,
			shouldErr:         false,
		},
		{
			name:              "timeout flag",
			args:              []string{"--timeout=60"},
			expectedNamespace: "",
			expectedOutput:    tableFormat,
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
	labelSelector = ""
	timeoutSeconds = 30

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
		{"valid yaml format", "yaml", false},
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
	tests := []struct {
		name      string
		namespace string
		shouldErr bool
	}{
		{"empty namespace", "", false}, // Empty means all namespaces
		{"valid namespace", testNamespaceDefault, false},
		{"valid namespace with hyphen", testNamespaceKube, false},
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

// TestFormatAge tests the age formatting function.
func TestFormatAge(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"seconds", 45 * time.Second, "45s"},
		{"one minute", 1 * time.Minute, "1m"},
		{"minutes", 30 * time.Minute, "30m"},
		{"one hour", 1 * time.Hour, "1h"},
		{"hours", 12 * time.Hour, "12h"},
		{"one day", 24 * time.Hour, "1d"},
		{"multiple days", 5 * 24 * time.Hour, "5d"},
		{"less than minute", 30 * time.Second, "30s"},
		{"exactly minute", 60 * time.Second, "1m"},
		{"exactly hour", 60 * time.Minute, "1h"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatAge(tt.duration)
			if result != tt.expected {
				t.Errorf("formatAge(%v) = %s, want %s", tt.duration, result, tt.expected)
			}
		})
	}
}

// TestFormatImages tests the image formatting function.
func TestFormatImages(t *testing.T) {
	tests := []struct {
		name     string
		images   []string
		expected string
	}{
		{
			name:     "no images",
			images:   []string{},
			expected: "<none>",
		},
		{
			name:     "single image",
			images:   []string{testImageNginx},
			expected: testImageNginx,
		},
		{
			name:     "many images",
			images:   []string{testImageNginx, testImageRedis, testImagePostgres, "mysql:8.0", "mongodb:4.4"},
			expected: testImageNginx + "," + testImageRedis + " +3 more",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatImages(tt.images)
			if result != tt.expected {
				t.Errorf("formatImages(%v) = %s, want %s", tt.images, result, tt.expected)
			}
		})
	}
}

// TestTruncateString tests the string truncation function.
func TestTruncateString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "short string",
			input:    "hello",
			maxLen:   10,
			expected: "hello",
		},
		{
			name:     "long string",
			input:    "this-is-a-very-long-string-that-needs-truncating",
			maxLen:   20,
			expected: "this-is-a-very-lo...",
		},
		{
			name:     "empty string",
			input:    "",
			maxLen:   10,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateString(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncateString(%s, %d) = %s, want %s", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

// TestFormatDeploymentOutput tests the deployment output formatting.
func TestFormatDeploymentOutput(t *testing.T) {
	// Create test deployments
	testDeployments := []k8s.DeploymentInfo{
		{
			Name:      "nginx-deployment",
			Namespace: testNamespaceDefault,
			Replicas: struct {
				Desired   int32 `json:"desired"`
				Available int32 `json:"available"`
				Ready     int32 `json:"ready"`
				Updated   int32 `json:"updated"`
			}{
				Desired:   3,
				Available: 3,
				Ready:     3,
				Updated:   3,
			},
			Age:       24 * time.Hour,
			Images:    []string{testImageNginx},
			CreatedAt: time.Now().Add(-24 * time.Hour),
		},
	}

	tests := []struct {
		name        string
		format      string
		shouldError bool
	}{
		{"table format", "table", false},
		{"json format", "json", false},
		{"yaml format", "yaml", false},
		{"invalid format", "xml", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := formatDeploymentOutput(testDeployments, tt.format)
			if tt.shouldError && err == nil {
				t.Errorf("formatDeploymentOutput() should return error for format %s", tt.format)
			}
			if !tt.shouldError && err != nil {
				t.Errorf("formatDeploymentOutput() should not return error for format %s, got: %v", tt.format, err)
			}
		})
	}
}

// TestFormatDeploymentJSON tests JSON output formatting.
func TestFormatDeploymentJSON(t *testing.T) {
	testDeployments := []k8s.DeploymentInfo{
		{
			Name:      testDeploymentName,
			Namespace: testNamespaceDefault,
			CreatedAt: time.Now(),
		},
	}

	// Capture stdout to validate JSON structure
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	oldStdout := os.Stdout
	os.Stdout = w

	// Execute the function
	formatErr := formatDeploymentJSON(testDeployments)

	// Restore stdout and read captured output
	w.Close()
	os.Stdout = oldStdout

	if formatErr != nil {
		t.Errorf("formatDeploymentJSON() should not return error, got: %v", formatErr)
	}

	// Read and validate JSON output
	outputBytes, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("Failed to read captured output: %v", err)
	}

	// Validate JSON structure
	var output struct {
		Kind       string               `json:"kind"`
		APIVersion string               `json:"apiVersion"`
		Items      []k8s.DeploymentInfo `json:"items"`
		Count      int                  `json:"count"`
	}

	if err := json.Unmarshal(outputBytes, &output); err != nil {
		t.Errorf("Failed to parse JSON output: %v", err)
	}

	// Validate expected fields
	if output.Kind != "DeploymentList" {
		t.Errorf("Expected kind 'DeploymentList', got '%s'", output.Kind)
	}
	if output.APIVersion != "apps/v1" {
		t.Errorf("Expected apiVersion 'apps/v1', got '%s'", output.APIVersion)
	}
	if output.Count != len(testDeployments) {
		t.Errorf("Expected count %d, got %d", len(testDeployments), output.Count)
	}
	if len(output.Items) != len(testDeployments) {
		t.Errorf("Expected %d items, got %d", len(testDeployments), len(output.Items))
	}
}

// TestFormatDeploymentTable tests table output formatting.
func TestFormatDeploymentTable(t *testing.T) {
	tests := []struct {
		name        string
		deployments []k8s.DeploymentInfo
		namespace   string
	}{
		{
			name:        "empty deployments",
			deployments: []k8s.DeploymentInfo{},
			namespace:   "",
		},
		{
			name: "single deployment",
			deployments: []k8s.DeploymentInfo{
				{
					Name:      testDeploymentName,
					Namespace: testNamespaceDefault,
					Replicas: struct {
						Desired   int32 `json:"desired"`
						Available int32 `json:"available"`
						Ready     int32 `json:"ready"`
						Updated   int32 `json:"updated"`
					}{
						Desired:   1,
						Available: 1,
						Ready:     1,
						Updated:   1,
					},
					Age:    time.Hour,
					Images: []string{"nginx:latest"},
				},
			},
			namespace: testNamespaceDefault,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalNamespace := namespace
			namespace = tt.namespace
			defer func() {
				namespace = originalNamespace
			}()

			err := formatDeploymentTable(tt.deployments)
			if err != nil {
				t.Errorf("formatDeploymentTable() should not return error, got: %v", err)
			}
		})
	}
}
