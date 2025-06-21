package cmd

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var appVersion = "dev"

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Long:  `Print the version number of k8s-controller`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info().Str("version", appVersion).Msg("Version requested")
		fmt.Printf("k8s-controller version %s\n", appVersion)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
