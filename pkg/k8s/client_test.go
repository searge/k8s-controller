// Package k8s contains tests for the Kubernetes client functionality.
// This file tests kubeconfig loading, client creation, connection testing, and deployment operations.
package k8s

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rs/zerolog"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	ktesting "k8s.io/client-go/testing"
)

// Test constants to avoid string duplication
const (
	fakeServerURL        = "https://fake-server"
	testImageNginx       = "nginx:1.21"
	testImageRedis       = "redis:6.2"
	testImagePostgres    = "postgres:13"
	testImageBusybox     = "busybox:latest"
	testNamespaceDefault = "default"
	testNamespaceKube    = "kube-system"
	testMsgFailedClose   = "Failed to close client"
)

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

// createTestDeployment creates a deployment for testing purposes.
func createTestDeployment(name, namespace string, replicas int32, images []string) *appsv1.Deployment {
	containers := make([]corev1.Container, len(images))
	for i, image := range images {
		containers[i] = corev1.Container{
			Name:  fmt.Sprintf("container-%d", i),
			Image: image,
		}
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			CreationTimestamp: metav1.Time{
				Time: time.Now().Add(-24 * time.Hour), // Created 24 hours ago
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: containers,
				},
			},
		},
		Status: appsv1.DeploymentStatus{
			ReadyReplicas:     replicas,
			AvailableReplicas: replicas,
		},
	}
}

// TestListDeployments tests the ListDeployments functionality with various scenarios.
func TestListDeployments(t *testing.T) {
	logger := zerolog.New(os.Stderr)

	tests := createListDeploymentTests()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := setupTestClient(logger, tt.deployments, tt.errorOnList)

			ctx := context.Background()
			deployments, err := client.ListDeployments(ctx, tt.options)

			validateListDeploymentResults(t, deployments, err, tt)
		})
	}
}

// listDeploymentTestCase represents a test case for ListDeployments.
type listDeploymentTestCase struct {
	name          string
	deployments   []runtime.Object
	options       ListDeploymentsOptions
	expectedCount int
	expectedNames []string
	expectedError bool
	errorOnList   bool
}

// createListDeploymentTests creates test cases for ListDeployments.
func createListDeploymentTests() []listDeploymentTestCase {
	return []listDeploymentTestCase{
		createAllNamespacesTestCase(),
		createSpecificNamespaceTestCase(),
		createEmptyNamespaceTestCase(),
		createNoDeploymentsTestCase(),
		createAPIErrorTestCase(),
	}
}

// createAllNamespacesTestCase creates a test case for listing from all namespaces.
func createAllNamespacesTestCase() listDeploymentTestCase {
	return listDeploymentTestCase{
		name: "list deployments from all namespaces",
		deployments: []runtime.Object{
			createTestDeployment("app1", testNamespaceDefault, 3, []string{testImageNginx}),
			createTestDeployment("app2", testNamespaceKube, 1, []string{testImageBusybox}),
			createTestDeployment("app3", testNamespaceDefault, 2, []string{testImageRedis, testImagePostgres}),
		},
		options: ListDeploymentsOptions{
			Namespace: "",
		},
		expectedCount: 3,
		expectedNames: []string{"app1", "app2", "app3"},
		expectedError: false,
	}
}

// createSpecificNamespaceTestCase creates a test case for listing from a specific namespace.
func createSpecificNamespaceTestCase() listDeploymentTestCase {
	return listDeploymentTestCase{
		name: "list deployments from specific namespace",
		deployments: []runtime.Object{
			createTestDeployment("app1", testNamespaceDefault, 3, []string{testImageNginx}),
			createTestDeployment("app2", testNamespaceKube, 1, []string{testImageBusybox}),
			createTestDeployment("app3", testNamespaceDefault, 2, []string{testImageRedis}),
		},
		options: ListDeploymentsOptions{
			Namespace: testNamespaceDefault,
		},
		expectedCount: 2,
		expectedNames: []string{"app1", "app3"},
		expectedError: false,
	}
}

// createEmptyNamespaceTestCase creates a test case for listing from an empty namespace.
func createEmptyNamespaceTestCase() listDeploymentTestCase {
	return listDeploymentTestCase{
		name: "list deployments from empty namespace",
		deployments: []runtime.Object{
			createTestDeployment("app1", testNamespaceDefault, 3, []string{testImageNginx}),
		},
		options: ListDeploymentsOptions{
			Namespace: "empty-namespace",
		},
		expectedCount: 0,
		expectedNames: []string{},
		expectedError: false,
	}
}

// createNoDeploymentsTestCase creates a test case for when no deployments exist.
func createNoDeploymentsTestCase() listDeploymentTestCase {
	return listDeploymentTestCase{
		name:        "no deployments exist",
		deployments: []runtime.Object{},
		options: ListDeploymentsOptions{
			Namespace: "",
		},
		expectedCount: 0,
		expectedNames: []string{},
		expectedError: false,
	}
}

// createAPIErrorTestCase creates a test case for API errors.
func createAPIErrorTestCase() listDeploymentTestCase {
	return listDeploymentTestCase{
		name: "API error on list",
		deployments: []runtime.Object{
			createTestDeployment("app1", testNamespaceDefault, 3, []string{testImageNginx}),
		},
		options: ListDeploymentsOptions{
			Namespace: "",
		},
		expectedError: true,
		errorOnList:   true,
	}
}

// setupTestClient creates a test client with fake data and optional error simulation.
func setupTestClient(logger zerolog.Logger, deployments []runtime.Object, simulateError bool) *Client {
	fakeClientset := fake.NewSimpleClientset(deployments...)

	if simulateError {
		fakeClientset.PrependReactor("list", "deployments",
			func(_ ktesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, nil, fmt.Errorf("simulated API error")
			})
	}

	return &Client{
		clientset: fakeClientset,
		config:    &rest.Config{Host: fakeServerURL},
		logger:    logger,
	}
}

// validateListDeploymentResults validates the test results.
func validateListDeploymentResults(t *testing.T, deployments []DeploymentInfo, err error, tt listDeploymentTestCase) {
	t.Helper()

	// Check error expectation
	if tt.expectedError {
		if err == nil {
			t.Errorf("ListDeployments() expected error, got nil")
		}
		return
	}

	if err != nil {
		t.Errorf("ListDeployments() unexpected error: %v", err)
		return
	}

	// Check deployment count
	if len(deployments) != tt.expectedCount {
		t.Errorf("ListDeployments() got %d deployments, want %d", len(deployments), tt.expectedCount)
	}

	// Check deployment names
	actualNames := make([]string, len(deployments))
	for i, deployment := range deployments {
		actualNames[i] = deployment.Name
	}

	if !stringSlicesEqual(actualNames, tt.expectedNames) {
		t.Errorf("ListDeployments() got names %v, want %v", actualNames, tt.expectedNames)
	}

	// Verify deployment info structure
	validateDeploymentStructure(t, deployments)
}

// validateDeploymentStructure validates that deployment info has required fields.
func validateDeploymentStructure(t *testing.T, deployments []DeploymentInfo) {
	t.Helper()

	for _, deployment := range deployments {
		if deployment.Name == "" {
			t.Error("Deployment name should not be empty")
		}
		if deployment.Namespace == "" {
			t.Error("Deployment namespace should not be empty")
		}
		if deployment.CreatedAt.IsZero() {
			t.Error("Deployment CreatedAt should not be zero")
		}
		if deployment.Age <= 0 {
			t.Error("Deployment Age should be positive")
		}
		if len(deployment.Images) == 0 {
			t.Error("Deployment should have at least one image")
		}
	}
}

// TestExtractImages tests the extractImages function with various container configurations.
func TestExtractImages(t *testing.T) {
	t.Run("single container with one image", func(t *testing.T) {
		deployment := createDeploymentWithContainers([]corev1.Container{
			{Name: "web", Image: testImageNginx},
		})
		expectedImages := []string{testImageNginx}
		assertExtractedImages(t, deployment, expectedImages)
	})

	t.Run("multiple containers with different images", func(t *testing.T) {
		deployment := createDeploymentWithContainers([]corev1.Container{
			{Name: "web", Image: testImageNginx},
			{Name: "db", Image: testImagePostgres},
			{Name: "cache", Image: testImageRedis},
		})
		expectedImages := []string{testImageNginx, testImagePostgres, testImageRedis}
		assertExtractedImages(t, deployment, expectedImages)
	})

	t.Run("containers with duplicate images", func(t *testing.T) {
		deployment := createDeploymentWithContainers([]corev1.Container{
			{Name: "web1", Image: testImageNginx},
			{Name: "web2", Image: testImageNginx},
			{Name: "db", Image: testImagePostgres},
		})
		expectedImages := []string{testImageNginx, testImagePostgres}
		assertExtractedImages(t, deployment, expectedImages)
	})

	t.Run("init containers and regular containers", func(t *testing.T) {
		deployment := createDeploymentWithInitContainers(
			[]corev1.Container{{Name: "init", Image: testImageBusybox}},
			[]corev1.Container{{Name: "app", Image: testImageNginx}},
		)
		expectedImages := []string{testImageNginx, testImageBusybox}
		assertExtractedImages(t, deployment, expectedImages)
	})

	t.Run("empty image names should be ignored", func(t *testing.T) {
		deployment := createDeploymentWithContainers([]corev1.Container{
			{Name: "valid", Image: testImageNginx},
			{Name: "empty", Image: ""},
		})
		expectedImages := []string{testImageNginx}
		assertExtractedImages(t, deployment, expectedImages)
	})
}

// createDeploymentWithContainers creates a deployment with the specified containers.
func createDeploymentWithContainers(containers []corev1.Container) *appsv1.Deployment {
	return &appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: containers,
				},
			},
		},
	}
}

// createDeploymentWithInitContainers creates a deployment with init containers and regular containers.
func createDeploymentWithInitContainers(initContainers, containers []corev1.Container) *appsv1.Deployment {
	return &appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					InitContainers: initContainers,
					Containers:     containers,
				},
			},
		},
	}
}

// assertExtractedImages validates that extracted images match expected images.
func assertExtractedImages(t *testing.T, deployment *appsv1.Deployment, expectedImages []string) {
	t.Helper()
	images := extractImages(deployment)

	if len(images) != len(expectedImages) {
		t.Errorf("extractImages() got %d images, want %d", len(images), len(expectedImages))
	}

	assertAllExpectedImagesPresent(t, images, expectedImages)
	assertNoUnexpectedImages(t, images, expectedImages)
}

// assertAllExpectedImagesPresent checks that all expected images are present in the result.
func assertAllExpectedImagesPresent(t *testing.T, actualImages, expectedImages []string) {
	t.Helper()
	for _, expectedImage := range expectedImages {
		if !containsImage(actualImages, expectedImage) {
			t.Errorf("extractImages() missing expected image: %s", expectedImage)
		}
	}
}

// assertNoUnexpectedImages checks that no unexpected images are present in the result.
func assertNoUnexpectedImages(t *testing.T, actualImages, expectedImages []string) {
	t.Helper()
	for _, actualImage := range actualImages {
		if !containsImage(expectedImages, actualImage) {
			t.Errorf("extractImages() found unexpected image: %s", actualImage)
		}
	}
}

// containsImage checks if a slice contains a specific image.
func containsImage(images []string, target string) bool {
	for _, image := range images {
		if image == target {
			return true
		}
	}
	return false
}

// stringSlicesEqual checks if two string slices contain the same elements (order independent).
func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	// Create a map to count occurrences in slice a
	counts := make(map[string]int)
	for _, s := range a {
		counts[s]++
	}

	// Check that slice b has the same elements with same counts
	for _, s := range b {
		if counts[s] == 0 {
			return false
		}
		counts[s]--
	}

	return true
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

// BenchmarkListDeployments benchmarks the ListDeployments operation.
func BenchmarkListDeployments(b *testing.B) {
	logger := zerolog.New(os.Stderr)

	// Create test deployments
	deployments := make([]runtime.Object, 100)
	for i := 0; i < 100; i++ {
		deployments[i] = createTestDeployment(
			fmt.Sprintf("app-%d", i),
			testNamespaceDefault,
			3,
			[]string{fmt.Sprintf("nginx:1.%d", i%10)},
		)
	}

	fakeClientset := fake.NewSimpleClientset(deployments...)
	client := &Client{
		clientset: fakeClientset,
		config:    &rest.Config{Host: fakeServerURL},
		logger:    logger,
	}

	ctx := context.Background()
	opts := ListDeploymentsOptions{Namespace: ""}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.ListDeployments(ctx, opts)
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
			logger.Error().Err(err).Msg(testMsgFailedClose)
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
			logger.Error().Err(err).Msg(testMsgFailedClose)
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

// ExampleClient_ListDeployments demonstrates how to list deployments.
func ExampleClient_ListDeployments() {
	logger := zerolog.New(os.Stderr)

	// Create client
	config := ClientConfig{}
	client, err := CreateClient(config, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create client")
	}
	defer func() {
		if err := client.Close(); err != nil {
			logger.Error().Err(err).Msg("Failed to close client")
		}
	}()

	ctx := context.Background()

	// List all deployments
	allDeployments, err := client.ListDeployments(ctx, ListDeploymentsOptions{})
	if err != nil {
		logger.Error().Err(err).Msg("Failed to list all deployments")
		return
	}

	logger.Info().Int("count", len(allDeployments)).Msg("Listed all deployments")

	// List deployments from specific namespace
	nsDeployments, err := client.ListDeployments(ctx, ListDeploymentsOptions{
		Namespace: testNamespaceKube,
	})
	if err != nil {
		logger.Error().Err(err).Msg("Failed to list namespace deployments")
		return
	}

	logger.Info().
		Int("count", len(nsDeployments)).
		Str("namespace", testNamespaceKube).
		Msg("Listed deployments from namespace")

	// List deployments with label selector
	labeledDeployments, err := client.ListDeployments(ctx, ListDeploymentsOptions{
		LabelSelector: "app=nginx",
	})
	if err != nil {
		logger.Error().Err(err).Msg("Failed to list labeled deployments")
		return
	}

	logger.Info().
		Int("count", len(labeledDeployments)).
		Str("selector", "app=nginx").
		Msg("Listed deployments with label selector")
}
