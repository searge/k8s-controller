// Package cmd contains the CLI commands for the k8s-controller application.
// This file implements the 'version' command which displays the current version.
package cmd

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// appVersion holds the current version of the application.
// This value can be overridden at build time using ldflags.
var appVersion = "dev"

// versionCmd represents the version command.
// It displays the current version of the k8s-controller application
// and logs the version request for audit purposes.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Long:  `Print the version number of k8s-controller`,
	Run: func(_ *cobra.Command, _ []string) {
		log.Info().Str("version", appVersion).Msg("Version requested")
		fmt.Printf("k8s-controller version %s\n", appVersion)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
