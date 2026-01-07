// Package k8s provides Kubernetes client functionality for the k8s-controller application.
// It handles kubeconfig loading, client creation, and connection testing with structured logging.
package k8s

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
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

// DeploymentInfo represents essential information about a Kubernetes deployment.
// This struct contains only the fields needed for listing operations.
type DeploymentInfo struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Replicas  struct {
		Desired   int32 `json:"desired"`
		Available int32 `json:"available"`
		Ready     int32 `json:"ready"`
	} `json:"replicas"`
	Age       time.Duration `json:"age"`
	Images    []string      `json:"images"`
	CreatedAt time.Time     `json:"created_at"`
}

// ListDeploymentsOptions holds options for listing deployments.
type ListDeploymentsOptions struct {
	// Namespace specifies the namespace to list deployments from.
	// If empty, deployments from all namespaces will be listed.
	Namespace string

	// LabelSelector allows filtering deployments by labels.
	// Uses the standard Kubernetes label selector syntax.
	LabelSelector string

	// FieldSelector allows filtering deployments by fields.
	// Uses the standard Kubernetes field selector syntax.
	FieldSelector string
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

	// Use client-go's built-in loading rules
	// This automatically handles:
	// - KUBECONFIG env variable (with : separator for multiple files)
	// - ~/.kube/config fallback
	// - Merging multiple configs (kubectl behavior)
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()

	// Only override if explicit --kubeconfig flag provided
	if config.KubeconfigPath != "" {
		loadingRules.ExplicitPath = config.KubeconfigPath
		logger.Debug().Str("path", config.KubeconfigPath).Msg("Using explicit kubeconfig path")
	}

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

	// Use the provided context directly, or add a reasonable timeout if none exists
	testCtx := ctx
	if deadline, ok := ctx.Deadline(); !ok || time.Until(deadline) > 10*time.Second {
		var cancel context.CancelFunc
		testCtx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	// Try to list namespaces as a connection test
	namespaces, err := c.clientset.CoreV1().Namespaces().List(testCtx, metav1.ListOptions{
		Limit: 1, // We only need to verify connection, not get all namespaces
	})
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to connect to Kubernetes API")
		return fmt.Errorf("failed to connect to Kubernetes API: %w", err)
	}

	// Get server version (optional, may add latency)
	if serverVersion, err := c.clientset.Discovery().ServerVersion(); err == nil {
		c.logger.Info().
			Int("namespace_count", len(namespaces.Items)).
			Str("server_version", serverVersion.String()).
			Str("server_host", c.config.Host).
			Msg("Successfully connected to Kubernetes API")
	} else {
		c.logger.Info().
			Int("namespace_count", len(namespaces.Items)).
			Str("server_host", c.config.Host).
			Msg("Successfully connected to Kubernetes API")
	}

	return nil
}

// ListDeployments retrieves deployments from the Kubernetes cluster based on the provided options.
// It returns a slice of DeploymentInfo structs containing essential deployment information.
func (c *Client) ListDeployments(ctx context.Context, opts ListDeploymentsOptions) ([]DeploymentInfo, error) {
	c.logger.Debug().
		Str("namespace", opts.Namespace).
		Str("label_selector", opts.LabelSelector).
		Msg("Listing deployments")

	deploymentList, err := c.fetchDeploymentList(ctx, opts)
	if err != nil {
		return nil, err
	}

	deployments := c.convertToDeploymentInfo(deploymentList.Items)

	c.logger.Info().
		Int("count", len(deployments)).
		Str("namespace", opts.Namespace).
		Msg("Successfully listed deployments")

	return deployments, nil
}

// fetchDeploymentList retrieves the raw deployment list from Kubernetes API.
func (c *Client) fetchDeploymentList(ctx context.Context, opts ListDeploymentsOptions) (*appsv1.DeploymentList, error) {
	listOpts := metav1.ListOptions{
		LabelSelector: opts.LabelSelector,
		FieldSelector: opts.FieldSelector,
	}

	var deploymentList *appsv1.DeploymentList
	var err error

	if opts.Namespace == "" {
		c.logger.Debug().Msg("Listing deployments from all namespaces")
		deploymentList, err = c.clientset.AppsV1().Deployments("").List(ctx, listOpts)
	} else {
		c.logger.Debug().Str("namespace", opts.Namespace).Msg("Listing deployments from namespace")
		deploymentList, err = c.clientset.AppsV1().Deployments(opts.Namespace).List(ctx, listOpts)
	}

	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to list deployments")
		return nil, fmt.Errorf("failed to list deployments: %w", err)
	}

	return deploymentList, nil
}

// convertToDeploymentInfo converts Kubernetes deployment objects to DeploymentInfo structs.
func (c *Client) convertToDeploymentInfo(deployments []appsv1.Deployment) []DeploymentInfo {
	result := make([]DeploymentInfo, 0, len(deployments))
	now := time.Now()

	for _, deployment := range deployments {
		info := c.createDeploymentInfo(deployment, now)
		result = append(result, info)
	}

	return result
}

// createDeploymentInfo creates a DeploymentInfo struct from a Kubernetes deployment.
func (c *Client) createDeploymentInfo(deployment appsv1.Deployment, now time.Time) DeploymentInfo {
	info := DeploymentInfo{
		Name:      deployment.Name,
		Namespace: deployment.Namespace,
		CreatedAt: deployment.CreationTimestamp.Time,
		Age:       now.Sub(deployment.CreationTimestamp.Time),
		Images:    extractImages(&deployment),
	}

	// Extract replica information
	if deployment.Spec.Replicas != nil {
		info.Replicas.Desired = *deployment.Spec.Replicas
	}
	info.Replicas.Available = deployment.Status.AvailableReplicas
	info.Replicas.Ready = deployment.Status.ReadyReplicas

	return info
}

// extractImages extracts all unique container images from a deployment.
// It processes both init containers and regular containers.
func extractImages(deployment *appsv1.Deployment) []string {
	imageSet := make(map[string]struct{})

	collectContainerImages(deployment.Spec.Template.Spec.Containers, imageSet)
	collectContainerImages(deployment.Spec.Template.Spec.InitContainers, imageSet)

	return convertImageSetToSlice(imageSet)
}

// collectContainerImages adds container images to the image set.
func collectContainerImages(containers []corev1.Container, imageSet map[string]struct{}) {
	for _, container := range containers {
		if container.Image != "" {
			imageSet[container.Image] = struct{}{}
		}
	}
}

// convertImageSetToSlice converts a map of images to a slice.
func convertImageSetToSlice(imageSet map[string]struct{}) []string {
	images := make([]string, 0, len(imageSet))
	for image := range imageSet {
		images = append(images, image)
	}
	return images
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
