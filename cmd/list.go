// Package cmd contains the CLI commands for the k8s-controller application.
// This file implements the 'list' command which provides subcommands for listing Kubernetes resources.
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/Searge/k8s-controller/pkg/k8s"
)

// listCmd represents the list command.
// It serves as a parent command for various resource listing operations.
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List Kubernetes resources",
	Long: `List various Kubernetes resources in your cluster.

This command provides subcommands for listing different types of resources
such as deployments, pods, services, etc.

Examples:
  kc list deployments
  kc list deployments --namespace=default
  kc list deployments --output=json`,
	Run: func(cmd *cobra.Command, _ []string) {
		// If no subcommand is specified, show help
		_ = cmd.Help()
	},
}

// Shared flags for list operations
var (
	// namespace specifies the Kubernetes namespace to list resources from.
	// If empty, resources from all namespaces will be listed.
	namespace string

	// outputFormat specifies the output format for the listed resources.
	// Supported formats: table, json
	outputFormat string

	// labelSelector allows filtering resources by labels.
	labelSelector string
)

// listDeploymentsCmd represents the list deployments command.
// It lists Kubernetes deployments with optional namespace filtering and output formatting.
var listDeploymentsCmd = &cobra.Command{
	Use:   "deployments",
	Short: "List deployments",
	Long: `List Kubernetes deployments in the specified namespace or all namespaces.

This command connects to the Kubernetes API and retrieves deployment information.
You can filter by namespace and choose different output formats.

Examples:
  kc list deployments                           # List all deployments
  kc list deployments -n default               # List deployments in default namespace
  kc list deployments -o json                  # Output in JSON format
  kc list deployments -n kube-system -o table  # Specific namespace, table format
  kc list deployments -l app=nginx             # Filter by label selector
  kc list deployments --kubeconfig=/path/to/config  # Use specific kubeconfig`,
	Run: func(_ *cobra.Command, _ []string) {
		log.Info().
			Str("namespace", namespace).
			Str("output", outputFormat).
			Str("labelSelector", labelSelector).
			Msg("Listing deployments")

		if err := runListDeployments(); err != nil {
			log.Error().Err(err).Msg("Failed to list deployments")
			os.Exit(1)
		}
	},
}

// runListDeployments executes the deployment listing logic.
// It creates a Kubernetes client, fetches deployments, and formats the output.
func runListDeployments() error {
	// Validate output format first
	if err := validateOutputFormat(outputFormat); err != nil {
		return fmt.Errorf("invalid output format: %w", err)
	}

	// Validate namespace
	if err := validateNamespace(namespace); err != nil {
		return fmt.Errorf("invalid namespace: %w", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	// Configure Kubernetes client
	clientConfig := k8s.ClientConfig{
		KubeconfigPath: kubeconfigPath,
		Context:        contextName,
	}

	// Create Kubernetes client
	client, err := k8s.CreateClient(clientConfig, log.Logger)
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			log.Warn().Err(closeErr).Msg("Failed to close Kubernetes client")
		}
	}()

	// Prepare list options
	listOptions := k8s.ListDeploymentsOptions{
		Namespace:     namespace,
		LabelSelector: labelSelector,
	}

	// List deployments
	deployments, err := client.ListDeployments(ctx, listOptions)
	if err != nil {
		// Provide helpful error context
		if strings.Contains(err.Error(), "connection refused") {
			return fmt.Errorf("failed to connect to Kubernetes API server - "+
				"is the cluster running and accessible? %w", err)
		}
		if strings.Contains(err.Error(), "forbidden") {
			return fmt.Errorf("insufficient permissions to list deployments - check your RBAC configuration: %w", err)
		}
		if strings.Contains(err.Error(), "not found") && namespace != "" {
			return fmt.Errorf("namespace '%s' not found: %w", namespace, err)
		}
		return fmt.Errorf("failed to list deployments: %w", err)
	}

	// Format and display output
	return formatDeploymentOutput(deployments, outputFormat)
}

// formatDeploymentOutput formats and displays deployments in the specified format.
func formatDeploymentOutput(deployments []k8s.DeploymentInfo, format string) error {
	switch format {
	case "json":
		return formatDeploymentJSON(deployments)
	case "table":
		return formatDeploymentTable(deployments)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

// formatDeploymentJSON outputs deployments in JSON format.
func formatDeploymentJSON(deployments []k8s.DeploymentInfo) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")

	output := struct {
		Kind       string               `json:"kind"`
		APIVersion string               `json:"apiVersion"`
		Items      []k8s.DeploymentInfo `json:"items"`
		Count      int                  `json:"count"`
	}{
		Kind:       "DeploymentList",
		APIVersion: "apps/v1",
		Items:      deployments,
		Count:      len(deployments),
	}

	return encoder.Encode(output)
}

// formatDeploymentTable outputs deployments in table format.
func formatDeploymentTable(deployments []k8s.DeploymentInfo) error {
	if len(deployments) == 0 {
		fmt.Println("No deployments found.")
		return nil
	}

	// Create tabwriter for aligned output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer func() {
		if err := w.Flush(); err != nil {
			log.Warn().Err(err).Msg("Failed to flush table writer")
		}
	}()

	// Print header
	var err error
	if namespace == "" {
		// Show namespace column when listing from all namespaces
		_, err = fmt.Fprintln(w, "NAMESPACE\tNAME\tREADY\tUP-TO-DATE\tAVAILABLE\tAGE\tIMAGES")
	} else {
		// Hide namespace column when listing from specific namespace
		_, err = fmt.Fprintln(w, "NAME\tREADY\tUP-TO-DATE\tAVAILABLE\tAGE\tIMAGES")
	}
	if err != nil {
		return fmt.Errorf("failed to write table header: %w", err)
	}

	// Print deployments
	for _, deployment := range deployments {
		readyStatus := fmt.Sprintf("%d/%d", deployment.Replicas.Ready, deployment.Replicas.Desired)
		ageString := formatAge(deployment.Age)
		imagesString := formatImages(deployment.Images)

		if namespace == "" {
			_, err = fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d\t%s\t%s\n",
				deployment.Namespace,
				deployment.Name,
				readyStatus,
				deployment.Replicas.Ready, // UP-TO-DATE approximation
				deployment.Replicas.Available,
				ageString,
				imagesString,
			)
		} else {
			_, err = fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%s\t%s\n",
				deployment.Name,
				readyStatus,
				deployment.Replicas.Ready, // UP-TO-DATE approximation
				deployment.Replicas.Available,
				ageString,
				imagesString,
			)
		}
		if err != nil {
			return fmt.Errorf("failed to write deployment row: %w", err)
		}
	}

	return nil
}

// formatAge formats a duration as a human-readable age string.
// It follows kubectl's age formatting conventions.
func formatAge(duration time.Duration) string {
	if duration < time.Minute {
		return fmt.Sprintf("%ds", int(duration.Seconds()))
	}
	if duration < time.Hour {
		return fmt.Sprintf("%dm", int(duration.Minutes()))
	}
	if duration < 24*time.Hour {
		return fmt.Sprintf("%dh", int(duration.Hours()))
	}
	days := int(duration.Hours() / 24)
	if days == 1 {
		return "1d"
	}
	return fmt.Sprintf("%dd", days)
}

// formatImages formats a slice of image names for display.
// It truncates long lists and shows a summary.
func formatImages(images []string) string {
	if len(images) == 0 {
		return "<none>"
	}

	if len(images) == 1 {
		return truncateString(images[0], 40)
	}

	if len(images) <= 3 {
		result := make([]string, len(images))
		for i, image := range images {
			result[i] = truncateString(image, 30)
		}
		return strings.Join(result, ",")
	}

	// Show first 2 images and count
	first := truncateString(images[0], 25)
	second := truncateString(images[1], 25)
	return fmt.Sprintf("%s,%s +%d more", first, second, len(images)-2)
}

// truncateString truncates a string to the specified length with ellipsis.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// validateOutputFormat ensures the output format is supported.
func validateOutputFormat(format string) error {
	switch format {
	case "table", "json":
		return nil
	default:
		return fmt.Errorf("unsupported format '%s', must be one of: table, json", format)
	}
}

// validateNamespace performs basic validation on the namespace parameter.
// Kubernetes namespace names must follow DNS label standards.
func validateNamespace(ns string) error {
	if ns == "" {
		return nil // Empty namespace means "all namespaces"
	}

	// Basic namespace name validation
	// Kubernetes names must be lowercase alphanumeric with hyphens
	if len(ns) > 63 {
		return fmt.Errorf("namespace name too long (max 63 characters)")
	}

	// Basic DNS label validation
	for i, r := range ns {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '-' {
			return fmt.Errorf("namespace name contains invalid character '%c' "+
				"(must be lowercase alphanumeric with hyphens)", r)
		}
		if (i == 0 || i == len(ns)-1) && r == '-' {
			return fmt.Errorf("namespace name cannot start or end with hyphen")
		}
	}

	return nil
}

func init() {
	// Register the list command with root
	rootCmd.AddCommand(listCmd)

	// Register the deployments subcommand with list
	listCmd.AddCommand(listDeploymentsCmd)

	// Add flags to the deployments command
	listDeploymentsCmd.Flags().StringVarP(&namespace, "namespace", "n", "",
		"Kubernetes namespace (default: all namespaces)")

	listDeploymentsCmd.Flags().StringVarP(&outputFormat, "output", "o", "table",
		"Output format (table|json)")

	listDeploymentsCmd.Flags().StringVarP(&labelSelector, "selector", "l", "",
		"Label selector to filter deployments")

	listDeploymentsCmd.Flags().StringVar(&kubeconfigPath, "kubeconfig", "",
		"Path to kubeconfig file (default: $KUBECONFIG or $HOME/.kube/config)")

	listDeploymentsCmd.Flags().StringVar(&contextName, "context", "",
		"Kubernetes context to use (default: current context from kubeconfig)")

	listDeploymentsCmd.Flags().IntVar(&timeoutSeconds, "timeout", 30,
		"Timeout for Kubernetes operations in seconds")
}
