// Package cmd contains the CLI commands for the k8s-controller application.
// This file implements the 'list' command which provides subcommands for listing Kubernetes resources.
package cmd

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
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
  kc list deployments                      # List all deployments
  kc list deployments -n default          # List deployments in default namespace
  kc list deployments -o json             # Output in JSON format
  kc list deployments -n kube-system -o table  # Specific namespace, table format`,
	Run: func(_ *cobra.Command, _ []string) {
		log.Info().
			Str("namespace", namespace).
			Str("output", outputFormat).
			Msg("Listing deployments")

		if err := runListDeployments(); err != nil {
			log.Error().Err(err).Msg("Failed to list deployments")
			os.Exit(1)
		}
	},
}

// runListDeployments executes the deployment listing logic.
// This function will be expanded in the next steps to include actual Kubernetes operations.
func runListDeployments() error {
	// TODO: Implement actual deployment listing using pkg/k8s
	// For now, just validate the parameters and show what we would do

	// Validate output format
	if err := validateOutputFormat(outputFormat); err != nil {
		return fmt.Errorf("invalid output format: %w", err)
	}

	// Validate namespace (if specified)
	if err := validateNamespace(namespace); err != nil {
		return fmt.Errorf("invalid namespace: %w", err)
	}

	// Placeholder implementation
	if namespace == "" {
		log.Info().Str("format", outputFormat).Msg("Would list deployments from all namespaces")
	} else {
		log.Info().
			Str("namespace", namespace).
			Str("format", outputFormat).
			Msg("Would list deployments from specified namespace")
	}

	return nil
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

	// More detailed validation could be added here if needed
	// For now, we trust that invalid names will be caught by the K8s API

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
}
