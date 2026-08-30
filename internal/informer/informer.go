// Package informer provides a cached, watch-backed view of cluster resources.
//
// Everything that answers a question about the cluster reads from here rather than
// calling the API server, so a handler serving many requests costs the API server
// nothing. Two consequences shape the API below: the cache is useless until it has
// synced, so callers must gate on [DeploymentWatcher.HasSynced]; and a watch runs
// for as long as the process does, so the lifetime is a context rather than a
// package-level goroutine.
//
// Dependencies arrive as arguments. Nothing here reads a command-line flag or a
// package-level variable, which is what lets the tests below drive it with a fake
// clientset and no cluster.
package informer

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"
)

// DefaultResyncPeriod is how often the informer re-lists and replays its cache
// through the event handlers. This is not a poll interval: changes arrive over the
// watch as they happen. Resync exists to recover from a handler that dropped work,
// so it is deliberately long.
const DefaultResyncPeriod = 10 * time.Minute

// Config holds the options for a watcher.
type Config struct {
	// Namespace restricts the watch to one namespace. Empty means all namespaces,
	// which needs a cluster-scoped list and watch permission.
	Namespace string

	// ResyncPeriod overrides DefaultResyncPeriod. Zero means use the default.
	ResyncPeriod time.Duration
}

// resyncPeriod returns the effective resync period.
func (c Config) resyncPeriod() time.Duration {
	if c.ResyncPeriod <= 0 {
		return DefaultResyncPeriod
	}
	return c.ResyncPeriod
}

// DeploymentWatcher keeps a synced local cache of Deployments.
//
// Deployment is not the resource this tool ultimately cares about. It is here
// because the shared informer factory is generic over resource types, so the
// mechanism can be built and tested before the question it will answer is chosen.
type DeploymentWatcher struct {
	factory  informers.SharedInformerFactory
	informer cache.SharedIndexInformer
	lister   appslisters.DeploymentLister
	logger   zerolog.Logger
	resync   time.Duration
	counts   eventCounters
}

// NewDeploymentWatcher builds a watcher over the given clientset.
//
// It registers event handlers but starts nothing: call [DeploymentWatcher.Run] for
// that. Taking kubernetes.Interface rather than *kubernetes.Clientset is what makes
// the fake clientset usable in tests.
func NewDeploymentWatcher(
	client kubernetes.Interface,
	cfg Config,
	logger zerolog.Logger,
) (*DeploymentWatcher, error) {
	opts := []informers.SharedInformerOption{}
	if cfg.Namespace != "" {
		opts = append(opts, informers.WithNamespace(cfg.Namespace))
	}

	resync := cfg.resyncPeriod()
	factory := informers.NewSharedInformerFactoryWithOptions(client, resync, opts...)
	deployments := factory.Apps().V1().Deployments()

	w := &DeploymentWatcher{
		factory:  factory,
		informer: deployments.Informer(),
		lister:   deployments.Lister(),
		resync:   resync,
		logger: logger.With().
			Str("component", "deployment-informer").
			Str("namespace", namespaceLabel(cfg.Namespace)).
			Logger(),
	}

	if err := w.registerHandlers(); err != nil {
		return nil, err
	}

	return w, nil
}

// namespaceLabel renders an empty namespace as something readable in a log line.
func namespaceLabel(ns string) string {
	if ns == "" {
		return "<all>"
	}
	return ns
}

// Run starts the watch and blocks until ctx is cancelled.
//
// On cancellation it waits for the factory's goroutines to finish, so a caller that
// returns from Run knows nothing is still running. That is what makes a leak
// detectable in a test rather than merely unlikely.
func (w *DeploymentWatcher) Run(ctx context.Context) error {
	w.logger.Info().Dur("resync", w.resync).Msg("Starting deployment informer")

	w.factory.Start(ctx.Done())

	// Cancellation during the initial sync is a shutdown, not a failure: a
	// SIGTERM that lands while the watch is still establishing must end the
	// process cleanly, the same as one that lands an hour later. Checked
	// before the sync error so the two cases stay distinct branches.
	err := w.WaitForSync(ctx)
	if ctx.Err() != nil {
		w.factory.Shutdown()
		w.logger.Info().Msg("Cancelled during initial sync, deployment informer stopped")
		return nil
	}
	if err != nil {
		return err
	}

	<-ctx.Done()
	w.logger.Info().Msg("Context cancelled, shutting down deployment informer")
	w.factory.Shutdown()
	w.logger.Info().Msg("Deployment informer stopped")

	return nil
}

// WaitForSync blocks until the cache holds the current state of the cluster, or
// until ctx is cancelled.
//
// Reading the cache before this returns gives an answer that is not wrong so much as
// arbitrary: it reflects however much of the initial list happened to arrive.
func (w *DeploymentWatcher) WaitForSync(ctx context.Context) error {
	if !cache.WaitForNamedCacheSync("deployments", ctx.Done(), w.informer.HasSynced) {
		return fmt.Errorf("deployment cache sync did not complete: %w", ctx.Err())
	}
	w.logger.Info().Msg("Deployment cache synced")
	return nil
}

// HasSynced reports whether the cache is populated. This is what a readiness probe
// gates on; liveness must not, since a process with an unsynced cache is healthy but
// not yet useful.
func (w *DeploymentWatcher) HasSynced() bool {
	return w.informer.HasSynced()
}

// List returns Deployments from the cache, never from the API server.
//
// A nil selector matches everything. An empty namespace lists across all namespaces
// the watcher covers.
func (w *DeploymentWatcher) List(namespace string, selector labels.Selector) ([]*appsv1.Deployment, error) {
	if selector == nil {
		selector = labels.Everything()
	}

	var (
		result []*appsv1.Deployment
		err    error
	)
	if namespace == "" {
		result, err = w.lister.List(selector)
	} else {
		result, err = w.lister.Deployments(namespace).List(selector)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments from cache: %w", err)
	}

	return result, nil
}
