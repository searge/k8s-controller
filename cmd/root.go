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
	Short: "Read-only Kubernetes cluster interrogation",
	Long: `Ask a Kubernetes cluster questions that are awkward to answer with kubectl,
and serve a cached view of cluster state over HTTP.

Read-only: the only cluster access is list and watch. Built step by step
alongside the FWDays course on Kubernetes controllers; the plan and the
decisions behind it live in docs/ in the repository.`,
	// Logger configuration only. The startup banner that used to live here
	// printed a log line before the output of help, version and every other
	// command that produces no logs of its own; commands that do real work
	// announce themselves instead (serve logs its port and informer).
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		logger.Init(logLevel)
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
