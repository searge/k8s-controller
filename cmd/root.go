package cmd

import (
	"os"

	"github.com/Searge/k8s-controller/pkg/logger"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var logLevel string

// rootCmd represents the base command when called without any subcommands
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
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Initialize logger with the specified log level
		logger.Init(logLevel)
		log.Info().Str("version", "dev").Msg("Starting k8s-controller")
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to execute command")
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info",
		"Log level (debug, info, warn, error, fatal, panic)")

	// Local flags
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
