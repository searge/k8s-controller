package cmd

import (
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/Searge/k8s-controller/pkg/server"
)

var serverPort int

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start HTTP server",
	Run: func(cmd *cobra.Command, args []string) {
		if err := server.Start(serverPort); err != nil {
			log.Error().Err(err).Msg("Failed to start server")
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().IntVar(&serverPort, "port", 8080, "Port to run the server on")
}
