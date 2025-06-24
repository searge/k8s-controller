// Package k8s provides Kubernetes client functionality for the k8s-controller application.
// It handles kubeconfig loading, client creation, and connection testing with structured logging.
package k8s

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// Client wraps the Kubernetes clientset with additional functionality.
// It provides structured logging and connection management for k8s operations.
type Client struct {
	clientset kubernetes.Interface
	config    *rest.Config
	logger    zerolog.Logger
}

// ClientConfig holds configuration options for creating a Kubernetes client.
type ClientConfig struct {
	// KubeconfigPath specifies the path to the kubeconfig file.
	// If empty, the default locations will be checked.
	KubeconfigPath string

	// Context specifies which context to use from the kubeconfig.
	// If empty, the current context will be used.
	Context string
}

// LoadKubeconfig loads the Kubernetes configuration from various sources.
// It follows the standard precedence: in-cluster config > kubeconfig file > default locations.
// Returns a *rest.Config that can be used to create a Kubernetes client.
func LoadKubeconfig(config ClientConfig, logger zerolog.Logger) (*rest.Config, error) {
	logger.Debug().Msg("Loading Kubernetes configuration")

	// Try in-cluster config first (for pods running inside K8s)
	if inClusterConfig, err := rest.InClusterConfig(); err == nil {
		logger.Info().Msg("Using in-cluster Kubernetes configuration")
		return inClusterConfig, nil
	}

	// Determine kubeconfig path
	kubeconfigPath := config.KubeconfigPath
	if kubeconfigPath == "" {
		kubeconfigPath = getDefaultKubeconfigPath()
	}

	logger.Debug().Str("path", kubeconfigPath).Msg("Loading kubeconfig from file")

	// Check if kubeconfig file exists
	if _, err := os.Stat(kubeconfigPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("kubeconfig file not found at %s", kubeconfigPath)
	}

	// Load config from kubeconfig file
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.ExplicitPath = kubeconfigPath

	configOverrides := &clientcmd.ConfigOverrides{}
	if config.Context != "" {
		configOverrides.CurrentContext = config.Context
		logger.Debug().Str("context", config.Context).Msg("Using specified context")
	}

	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		configOverrides,
	)

	restConfig, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	// Log current context
	if rawConfig, err := kubeConfig.RawConfig(); err == nil {
		logger.Info().
			Str("context", rawConfig.CurrentContext).
			Str("cluster", rawConfig.Contexts[rawConfig.CurrentContext].Cluster).
			Msg("Loaded Kubernetes configuration")
	}

	return restConfig, nil
}

// CreateClient creates a new Kubernetes client with the provided configuration.
// It returns a Client instance that wraps the clientset with additional functionality.
func CreateClient(config ClientConfig, logger zerolog.Logger) (*Client, error) {
	logger.Debug().Msg("Creating Kubernetes client")

	restConfig, err := LoadKubeconfig(config, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	// Create the clientset
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	client := &Client{
		clientset: clientset,
		config:    restConfig,
		logger:    logger.With().Str("component", "k8s-client").Logger(),
	}

	client.logger.Info().Msg("Kubernetes client created successfully")
	return client, nil
}

// TestConnection verifies that the client can connect to the Kubernetes API server.
// It performs a simple API call to list namespaces with a timeout.
func (c *Client) TestConnection(ctx context.Context) error {
	c.logger.Debug().Msg("Testing Kubernetes API connection")

	// Create a context with timeout for the connection test
	testCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Try to list namespaces as a connection test
	namespaces, err := c.clientset.CoreV1().Namespaces().List(testCtx, metav1.ListOptions{
		Limit: 1, // We only need to verify connection, not get all namespaces
	})
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to connect to Kubernetes API")
		return fmt.Errorf("failed to connect to Kubernetes API: %w", err)
	}

	c.logger.Info().
		Int("namespace_count", len(namespaces.Items)).
		Str("server_version", c.config.Host).
		Msg("Successfully connected to Kubernetes API")

	return nil
}

// GetClientset returns the underlying Kubernetes clientset.
// This allows access to all Kubernetes API operations.
func (c *Client) GetClientset() kubernetes.Interface {
	return c.clientset
}

// GetConfig returns the underlying REST config.
// This can be useful for creating other types of clients.
func (c *Client) GetConfig() *rest.Config {
	return c.config
}

// Close performs cleanup operations for the client.
// Currently, there's no cleanup needed for the Kubernetes client,
// but this method is provided for future extensibility.
func (c *Client) Close() error {
	c.logger.Debug().Msg("Closing Kubernetes client")
	return nil
}

// getDefaultKubeconfigPath returns the default kubeconfig file path.
// It follows the standard kubectl conventions.
func getDefaultKubeconfigPath() string {
	// Check KUBECONFIG environment variable first
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		return kubeconfig
	}

	// Use default location in home directory
	if home := homedir.HomeDir(); home != "" {
		return filepath.Join(home, ".kube", "config")
	}

	// Fallback to current directory (unlikely to work, but better than empty)
	return "./kubeconfig"
}
