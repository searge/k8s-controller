// Package cmd implements the command-line interface for the k8s-controller application.
// It uses the Cobra library to provide a structured CLI with subcommands and flags.
package cmd

import (
	"github.com/Searge/k8s-controller/pkg/logger"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var logLevel string

// rootCmd represents the base command when called without any subcommands.
// It serves as the entry point for the CLI application and handles global configuration
// such as logging setup that applies to all subcommands.
var rootCmd = &cobra.Command{
	Use:   "k8s-controller",
	Short: "A production-grade Golang Kubernetes controller",
	Long: `This project is a step-by-step tutorial for DevOps and SRE engineers
to learn about building Golang applications and Kubernetes controllers.
Each step is implemented as a feature branch and includes
a README section with explanations and command history

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	PersistentPreRun: func(cmd *cobra.Command, _ []string) {
		// Skip logging for version command - it should be clean output
		if cmd.Use == "version" {
			return
		}

		// Initialize logger with the specified log level
		logger.Init(logLevel)
		log.Info().Str("version", Version).Msg("Starting k8s-controller")
	},
	Run: func(cmd *cobra.Command, _ []string) {
		// If no subcommand is specified, show help
		_ = cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
// If the command execution fails, the application will exit with status code 1.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to execute command")
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info",
		"Log level (debug, info, warn, error, fatal, panic)")

	// Version flags - using SetVersionTemplate for proper Cobra integration
	rootCmd.Version = Version
	rootCmd.SetVersionTemplate("k8s-controller version {{.Version}}\n")

	// Silence automatic help/usage output on errors since we already log them
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true
}
