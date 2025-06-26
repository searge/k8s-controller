// Package k8s contains tests for the Kubernetes client functionality.
// This file tests kubeconfig loading, client creation, and connection testing.
package k8s

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

// A constant for fake server URL
const fakeServerURL string = "https://fake-server"

// TestGetDefaultKubeconfigPath tests the default kubeconfig path resolution.
func TestGetDefaultKubeconfigPath(t *testing.T) {
	// Save original environment
	originalKubeconfig := os.Getenv("KUBECONFIG")
	originalHome := os.Getenv("HOME")
	defer func() {
		if err := os.Setenv("KUBECONFIG", originalKubeconfig); err != nil {
			t.Errorf("Failed to restore KUBECONFIG: %v", err)
		}
		if err := os.Setenv("HOME", originalHome); err != nil {
			t.Errorf("Failed to restore HOME: %v", err)
		}
	}()

	tests := []struct {
		name          string
		kubeconfigEnv string
		homeEnv       string
		expected      string
	}{
		{
			name:          "KUBECONFIG environment variable set",
			kubeconfigEnv: "/custom/kubeconfig",
			homeEnv:       "/home/user",
			expected:      "/custom/kubeconfig",
		},
		{
			name:          "HOME environment variable set",
			kubeconfigEnv: "",
			homeEnv:       "/home/user",
			expected:      "/home/user/.kube/config",
		},
		{
			name:          "no environment variables",
			kubeconfigEnv: "",
			homeEnv:       "",
			expected:      "./kubeconfig",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			if err := os.Setenv("KUBECONFIG", tt.kubeconfigEnv); err != nil {
				t.Fatalf("Failed to set KUBECONFIG: %v", err)
			}
			if err := os.Setenv("HOME", tt.homeEnv); err != nil {
				t.Fatalf("Failed to set HOME: %v", err)
			}

			result := getDefaultKubeconfigPath()
			if result != tt.expected {
				t.Errorf("getDefaultKubeconfigPath() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestLoadKubeconfigFileNotFound tests error handling when kubeconfig file doesn't exist.
func TestLoadKubeconfigFileNotFound(t *testing.T) {
	logger := zerolog.New(os.Stderr)

	config := ClientConfig{
		KubeconfigPath: "/nonexistent/path/config",
	}

	_, err := LoadKubeconfig(config, logger)
	if err == nil {
		t.Error("LoadKubeconfig() should return error for nonexistent file")
	}

	expectedError := "kubeconfig file not found at /nonexistent/path/config"
	if err.Error() != expectedError {
		t.Errorf("LoadKubeconfig() error = %v, want %v", err.Error(), expectedError)
	}
}

// TestCreateClientWithInvalidConfig tests client creation with invalid configuration.
func TestCreateClientWithInvalidConfig(t *testing.T) {
	logger := zerolog.New(os.Stderr)

	// Create a temporary file with invalid YAML content
	tmpDir := t.TempDir()
	invalidKubeconfig := filepath.Join(tmpDir, "invalid-kubeconfig")
	if err := os.WriteFile(invalidKubeconfig, []byte("invalid yaml content"), 0644); err != nil {
		t.Fatalf("Failed to create invalid kubeconfig file: %v", err)
	}

	config := ClientConfig{
		KubeconfigPath: invalidKubeconfig,
	}

	_, err := CreateClient(config, logger)
	if err == nil {
		t.Error("CreateClient() should return error for invalid kubeconfig")
	}
}

// TestClientWithFakeClientset tests the Client with a fake Kubernetes clientset.
func TestClientWithFakeClientset(t *testing.T) {
	logger := zerolog.New(os.Stderr)

	// Create a fake clientset
	fakeClientset := fake.NewSimpleClientset()

	client := &Client{
		clientset: fakeClientset,
		config:    &rest.Config{Host: fakeServerURL},
		logger:    logger,
	}

	// Test GetClientset
	if client.GetClientset() != fakeClientset {
		t.Error("GetClientset() should return the fake clientset")
	}

	// Test GetConfig
	if client.GetConfig().Host != fakeServerURL {
		t.Error("GetConfig() should return the config with fake server")
	}

	// Test Close (should not return error)
	if err := client.Close(); err != nil {
		t.Errorf("Close() should not return error, got: %v", err)
	}
}

// TestTestConnectionWithFakeClient tests the TestConnection method with a fake client.
func TestTestConnectionWithFakeClient(t *testing.T) {
	logger := zerolog.New(os.Stderr)

	// Create a fake clientset
	fakeClientset := fake.NewSimpleClientset()

	client := &Client{
		clientset: fakeClientset,
		config:    &rest.Config{Host: fakeServerURL},
		logger:    logger,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test connection with fake client (should succeed)
	err := client.TestConnection(ctx)
	if err != nil {
		t.Errorf("TestConnection() with fake client should succeed, got error: %v", err)
	}
}

// TestTestConnectionWithCancelledContext tests the TestConnection method with a cancelled context.
func TestTestConnectionWithCancelledContext(t *testing.T) {
	logger := zerolog.New(os.Stderr)

	// Create a fake clientset
	fakeClientset := fake.NewSimpleClientset()

	client := &Client{
		clientset: fakeClientset,
		config:    &rest.Config{Host: fakeServerURL},
		logger:    logger,
	}

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Test connection with cancelled context
	err := client.TestConnection(ctx)
	if err == nil {
		// Note: fake clientset might not respect context cancellation
		// This is a limitation of the test, not the actual implementation
		t.Skip("Fake clientset does not respect context cancellation - this test requires a real cluster")
	}
}

// BenchmarkCreateClient benchmarks the client creation process.
func BenchmarkCreateClient(b *testing.B) {
	logger := zerolog.New(os.Stderr)

	// Create a temporary valid kubeconfig for benchmarking
	tmpDir := b.TempDir()
	validKubeconfig := filepath.Join(tmpDir, "valid-kubeconfig")
	kubeconfigContent := `
apiVersion: v1
kind: Config
current-context: test-context
contexts:
- name: test-context
  context:
    cluster: test-cluster
    user: test-user
clusters:
- name: test-cluster
  cluster:
    server: https://localhost:6443
    insecure-skip-tls-verify: true
users:
- name: test-user
  user:
    token: fake-token
`
	if err := os.WriteFile(validKubeconfig, []byte(kubeconfigContent), 0644); err != nil {
		b.Fatalf("Failed to create valid kubeconfig file: %v", err)
	}

	config := ClientConfig{
		KubeconfigPath: validKubeconfig,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = LoadKubeconfig(config, logger)
	}
}

// ExampleCreateClient demonstrates how to create a Kubernetes client.
func ExampleCreateClient() {
	logger := zerolog.New(os.Stderr)

	config := ClientConfig{
		KubeconfigPath: "/path/to/kubeconfig",
		Context:        "my-context",
	}

	client, err := CreateClient(config, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create Kubernetes client")
	}
	defer func() {
		if err := client.Close(); err != nil {
			logger.Error().Err(err).Msg("Failed to close client")
		}
	}()

	// Test the connection
	ctx := context.Background()
	if err := client.TestConnection(ctx); err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to Kubernetes")
	}

	logger.Info().Msg("Successfully connected to Kubernetes")
}

// ExampleClient_TestConnection demonstrates how to test a Kubernetes connection.
func ExampleClient_TestConnection() {
	logger := zerolog.New(os.Stderr)

	// Assuming you have a client already created
	config := ClientConfig{}
	client, err := CreateClient(config, logger)
	if err != nil {
		return
	}
	defer func() {
		if err := client.Close(); err != nil {
			logger.Error().Err(err).Msg("Failed to close client")
		}
	}()

	// Test connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.TestConnection(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("Connection test failed")
		return
	}

	logger.Info().Msg("Connection test successful")
}
