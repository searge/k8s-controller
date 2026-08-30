// Package cmd contains the CLI commands for the k8s-controller application.
// This file implements the 'serve' command which starts the HTTP server and the
// deployment informer under one signal-bound context.
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"k8s.io/client-go/kubernetes"

	"github.com/Searge/k8s-controller/internal/informer"
	"github.com/Searge/k8s-controller/pkg/server"
)

// serverPort holds the port number for the HTTP server, configured via CLI flag.
var serverPort int

// serveCmd represents the serve command which starts the HTTP server alongside a
// deployment informer. Both run until SIGINT or SIGTERM, then shut down together.
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP server and the deployment informer",
	Long: `Start the HTTP server and a deployment informer under one lifecycle.

The informer watches Deployments and keeps a local cache; the HTTP server runs
beside it. Both stop together on SIGINT or SIGTERM. Cluster access is required:
the command loads a kubeconfig the same way 'list' does.

Endpoints:
  - GET /healthz: liveness -- the process is up
  - GET /readyz: readiness -- 200 once the informer cache has synced, 503 before
  - GET /deployments[?namespace=...]: deployments served from the cache

Examples:
  k8s-controller serve
  k8s-controller serve --port=9090
  k8s-controller serve --port=8080 --log-level=debug
  k8s-controller serve --kubeconfig=/path/to/config --context=my-cluster`,
	Run: func(_ *cobra.Command, _ []string) {
		if err := validatePort(serverPort); err != nil {
			log.Error().Err(err).Msg("Invalid port number")
			os.Exit(1)
		}

		// The root context of the process: everything long-running hangs off
		// this, so one signal stops the informer and the server together.
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		client, err := createK8sClient()
		if err != nil {
			log.Error().Err(err).Msg("Failed to create Kubernetes client")
			os.Exit(1)
		}
		defer closeClient(client)

		log.Info().Int("port", serverPort).Msg("Starting serve")
		if err := runServe(ctx, client.GetClientset(), serverPort); err != nil {
			log.Error().Err(err).Msg("Serve exited with error")
			os.Exit(1)
		}
	},
}

// runServe runs the deployment informer and the HTTP server until ctx is
// cancelled or either of them fails.
//
// errgroup ties their fates together: if the server cannot listen, the group's
// context cancels and the informer stops rather than running on without an API
// to serve from; the same applies in reverse. It returns only when both have
// stopped, so a caller that saw it return knows nothing leaked.
//
// It takes kubernetes.Interface rather than the wrapped client so a test can
// drive it with a fake clientset and no cluster.
func runServe(ctx context.Context, clientset kubernetes.Interface, port int) error {
	watcher, err := informer.NewDeploymentWatcher(clientset, informer.Config{}, log.Logger)
	if err != nil {
		return fmt.Errorf("failed to build deployment informer: %w", err)
	}

	// The watcher is the server's only data source: handlers read the cache and
	// have no clientset to fall back on, and /readyz gates on its sync state.
	srv := server.New(port, watcher, log.Logger)

	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error { return watcher.Run(gctx) })
	g.Go(func() error { return srv.Start(gctx) })

	return g.Wait()
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

	// The same kubeconfig flags 'list deployments' has, bound to the same
	// shared variables from flags.go. The informer needs cluster access.
	serveCmd.Flags().StringVar(&kubeconfigPath, "kubeconfig", "",
		"Path to kubeconfig file (default: $KUBECONFIG or $HOME/.kube/config)")
	serveCmd.Flags().StringVar(&contextName, "context", "",
		"Kubernetes context to use (default: current context from kubeconfig)")
}
