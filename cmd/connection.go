// Package cmd contains the CLI commands for the k8s-controller application.
// This file implements the 'connection' command which verifies Kubernetes API connectivity.
package cmd

import (
	"context"
	"os"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/Searge/k8s-controller/pkg/k8s"
)

// Connection test configuration variables, set via CLI flags.
var (
	kubeconfigPath string
	contextName    string
	timeoutSeconds int
)

// connectionCmd represents the connection command.
// It creates a Kubernetes client and verifies connectivity to the API server.
var connectionCmd = &cobra.Command{
	Use:   "connection",
	Short: "Test Kubernetes API connectivity",
	Long: `Test the connection to the Kubernetes API server using the configured kubeconfig.

This command will:
  - Load the kubeconfig from the specified path or default location
  - Create a Kubernetes client
  - Perform a connection test by listing namespaces
  - Report the connection status and basic cluster information

Examples:
  k8s-controller connection
  k8s-controller connection --kubeconfig=/path/to/config
  k8s-controller connection --context=my-context --timeout=30`,
	Run: func(_ *cobra.Command, _ []string) {
		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
		defer cancel()

		// Configure client
		config := k8s.ClientConfig{
			KubeconfigPath: kubeconfigPath,
			Context:        contextName,
		}

		log.Info().Msg("Testing Kubernetes API connection...")

		// Create client
		client, err := k8s.CreateClient(config, log.Logger)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create Kubernetes client")
			os.Exit(1)
		}
		defer func() {
			if closeErr := client.Close(); closeErr != nil {
				log.Warn().Err(closeErr).Msg("Failed to close client")
			}
		}()

		// Test connection
		if err := client.TestConnection(ctx); err != nil {
			log.Error().Err(err).Msg("Connection test failed")
			os.Exit(1)
		}

		log.Info().Msg("✅ Connection test successful! Kubernetes API is reachable.")
	},
}

// init registers the connection command with the root command and configures its flags.
func init() {
	rootCmd.AddCommand(connectionCmd)

	// Kubeconfig path flag
	connectionCmd.Flags().StringVar(&kubeconfigPath, "kubeconfig", "",
		"Path to kubeconfig file (default: $KUBECONFIG or $HOME/.kube/config)")

	// Context name flag
	connectionCmd.Flags().StringVar(&contextName, "context", "",
		"Kubernetes context to use (default: current context from kubeconfig)")

	// Timeout flag
	connectionCmd.Flags().IntVar(&timeoutSeconds, "timeout", 10,
		"Connection timeout in seconds")
}
