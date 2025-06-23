// Package cmd contains the CLI commands for the k8s-controller application.
// This file implements the 'serve' command which starts the HTTP server.
package cmd

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/Searge/k8s-controller/pkg/server"
)

// serverPort holds the port number for the HTTP server, configured via CLI flag.
var serverPort int

// serveCmd represents the serve command which starts the HTTP server.
// It accepts a --port flag to specify which port to bind to (default: 8080).
// The command will block until the server encounters an error or is terminated.
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start HTTP server",
	Long: `Start the HTTP server with health check and debug endpoints.

The server provides the following endpoints:
  - GET /health: Health check endpoint returning JSON status
  - GET /*: Default greeting message for all other paths

Examples:
  k8s-controller serve
  k8s-controller serve --port=9090
  k8s-controller serve --port=8080 --log-level=debug`,
	Run: func(_ *cobra.Command, _ []string) {
		// Validate port range
		if err := validatePort(serverPort); err != nil {
			log.Error().Err(err).Msg("Invalid port number")
			os.Exit(1)
		}

		// Log server startup information
		log.Info().Int("port", serverPort).Msg("Starting HTTP server")

		// Start the server - this blocks until error or termination
		if err := server.Start(serverPort, log.Logger); err != nil {
			log.Error().Err(err).Msg("Failed to start server")
			os.Exit(1)
		}
	},
}

// validatePort checks if the provided port number is within the valid range.
// Valid TCP port numbers are 1-65535 (0 is reserved and typically not usable for binding).
func validatePort(port int) error {
	if port <= 0 || port > 65535 {
		return fmt.Errorf("invalid port number: %d, must be between 1 and 65535", port)
	}
	return nil
}

// init registers the serve command with the root command and configures its flags.
func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().IntVar(&serverPort, "port", 8080, "Port to run the server on (1-65535)")
}
