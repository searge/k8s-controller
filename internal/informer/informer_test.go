package informer

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/rs/zerolog"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

// syncTimeout bounds every wait in these tests. Long enough that a slow machine does
// not produce a flake, short enough that a genuine hang fails the run rather than
// stalling it.
const syncTimeout = 5 * time.Second

// testLogger discards output. The event handlers log at debug level, and asserting on
// log lines would test the logging rather than the cache.
func testLogger() zerolog.Logger {
	return zerolog.New(io.Discard)
}

// deployment builds a Deployment with the fields the cache keys on.
func deployment(namespace, name, resourceVersion string, replicas int32) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       namespace,
			Name:            name,
			ResourceVersion: resourceVersion,
			Labels:          map[string]string{"app": name},
		},
		Spec: appsv1.DeploymentSpec{Replicas: &replicas},
	}
}

// newWatchedFakeClient returns a fake clientset whose Deployment watch is driven by
// the returned FakeWatcher, so a test can deliver add, update and delete events by
// hand. The clientset starts with no objects: everything the cache holds arrives
// through the watch, which is what makes the assertions about event handling real
// rather than assertions about the initial list.
func newWatchedFakeClient() (*fake.Clientset, *watch.FakeWatcher) {
	client := fake.NewClientset()
	fakeWatch := watch.NewFake()
	client.PrependWatchReactor("deployments", k8stesting.DefaultWatchReactor(fakeWatch, nil))

	return client, fakeWatch
}

// startWatcher builds a watcher, runs it, and waits for the initial sync. It returns
// the watcher plus a function that cancels it and asserts Run returned.
func startWatcher(t *testing.T, client *fake.Clientset, cfg Config) (*DeploymentWatcher, func()) {
	t.Helper()

	w, err := NewDeploymentWatcher(client, cfg, testLogger())
	if err != nil {
		t.Fatalf("NewDeploymentWatcher() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	runDone := make(chan error, 1)
	go func() { runDone <- w.Run(ctx) }()

	syncCtx, syncCancel := context.WithTimeout(ctx, syncTimeout)
	defer syncCancel()
	if err := w.WaitForSync(syncCtx); err != nil {
		cancel()
		t.Fatalf("WaitForSync() error = %v", err)
	}

	return w, func() {
		cancel()
		select {
		case err := <-runDone:
			if err != nil {
				t.Errorf("Run() returned error after cancellation: %v", err)
			}
		case <-time.After(syncTimeout):
			// The whole point of Run waiting on factory.Shutdown() is that this
			// cannot happen quietly.
			t.Error("Run() did not return within the timeout after cancellation: informer goroutines leaked")
		}
	}
}

func TestConfigResyncPeriod(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want time.Duration
	}{
		{name: "zero falls back to the default", cfg: Config{}, want: DefaultResyncPeriod},
		{name: "negative falls back to the default", cfg: Config{ResyncPeriod: -time.Second}, want: DefaultResyncPeriod},
		{name: "positive is respected", cfg: Config{ResyncPeriod: 42 * time.Second}, want: 42 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.resyncPeriod(); got != tt.want {
				t.Errorf("resyncPeriod() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestCacheReflectsWatchEvents is the regression test for the cache itself: add,
// update and delete arriving over the watch must each be visible through List.
func TestCacheReflectsWatchEvents(t *testing.T) {
	client, fakeWatch := newWatchedFakeClient()
	w, stop := startWatcher(t, client, Config{})
	defer stop()

	if got := listNames(t, w, ""); len(got) != 0 {
		t.Fatalf("cache should start empty, got %v", got)
	}

	fakeWatch.Add(deployment("default", "api", "1", 1))
	waitForNames(t, w, "", []string{"default/api"})

	fakeWatch.Add(deployment("other", "worker", "1", 1))
	waitForNames(t, w, "", []string{"default/api", "other/worker"})

	// An update must change the cached object, not merely leave the name present.
	fakeWatch.Modify(deployment("default", "api", "2", 5))
	waitForReplicas(t, w, "default", "api", 5)

	fakeWatch.Delete(deployment("other", "worker", "1", 1))
	waitForNames(t, w, "", []string{"default/api"})
}

// TestListServesFromCacheNotAPI is the test decision 015 asks for: a read must not
// reach the API server. The fake clientset records every call, so a List that went to
// the API would show up as a "list" action after the informer's initial one.
func TestListServesFromCacheNotAPI(t *testing.T) {
	client, fakeWatch := newWatchedFakeClient()
	w, stop := startWatcher(t, client, Config{})
	defer stop()

	fakeWatch.Add(deployment("default", "api", "1", 1))
	waitForNames(t, w, "", []string{"default/api"})

	client.ClearActions()

	for range 5 {
		if _, err := w.List("", nil); err != nil {
			t.Fatalf("List() error = %v", err)
		}
	}

	for _, action := range client.Actions() {
		if action.GetVerb() == "list" || action.GetVerb() == "get" {
			t.Errorf("List() reached the API server: verb %q on %q", action.GetVerb(), action.GetResource().Resource)
		}
	}
}

func TestListFiltersByNamespaceAndSelector(t *testing.T) {
	client, fakeWatch := newWatchedFakeClient()
	w, stop := startWatcher(t, client, Config{})
	defer stop()

	fakeWatch.Add(deployment("default", "api", "1", 1))
	fakeWatch.Add(deployment("default", "web", "1", 1))
	fakeWatch.Add(deployment("other", "api", "1", 1))
	waitForNames(t, w, "", []string{"default/api", "default/web", "other/api"})

	if got := listNames(t, w, "default"); len(got) != 2 {
		t.Errorf("List(default) = %v, want 2 entries", got)
	}

	selector := labels.SelectorFromSet(labels.Set{"app": "web"})
	found, err := w.List("", selector)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(found) != 1 || found[0].Name != "web" {
		t.Errorf("List(selector app=web) returned %d entries, want just default/web", len(found))
	}
}

// TestHasSyncedGatesReads covers the readiness contract: HasSynced is false before the
// cache is populated, which is what /readyz must gate on.
func TestHasSyncedGatesReads(t *testing.T) {
	client, _ := newWatchedFakeClient()

	w, err := NewDeploymentWatcher(client, Config{}, testLogger())
	if err != nil {
		t.Fatalf("NewDeploymentWatcher() error = %v", err)
	}

	if w.HasSynced() {
		t.Error("HasSynced() = true before Run(), want false")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = w.Run(ctx) }()

	syncCtx, syncCancel := context.WithTimeout(ctx, syncTimeout)
	defer syncCancel()
	if err := w.WaitForSync(syncCtx); err != nil {
		t.Fatalf("WaitForSync() error = %v", err)
	}

	if !w.HasSynced() {
		t.Error("HasSynced() = false after sync, want true")
	}
}

// TestRunCancelledDuringSyncIsClean pins the shutdown-versus-failure distinction:
// a context that dies before the initial sync completes is a shutdown, so Run
// returns nil rather than a sync error. Found by the serve lifecycle test, where a
// fast cancel raced the initial sync and runServe reported a spurious failure.
func TestRunCancelledDuringSyncIsClean(t *testing.T) {
	client, _ := newWatchedFakeClient()

	w, err := NewDeploymentWatcher(client, Config{}, testLogger())
	if err != nil {
		t.Fatalf("NewDeploymentWatcher() error = %v", err)
	}

	// Cancelled before Run starts: the sync can never complete, which is the
	// deterministic version of "the signal arrived first".
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	done := make(chan error, 1)
	go func() { done <- w.Run(ctx) }()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run() on a pre-cancelled context = %v, want nil (shutdown, not failure)", err)
		}
	case <-time.After(syncTimeout):
		t.Error("Run() did not return on a pre-cancelled context")
	}
}

// TestWaitForSyncRespectsCancellation makes sure a cancelled context ends the wait
// rather than blocking forever. An informer whose watch never establishes must not
// hang the process that is waiting on it.
func TestWaitForSyncRespectsCancellation(t *testing.T) {
	client, _ := newWatchedFakeClient()

	w, err := NewDeploymentWatcher(client, Config{}, testLogger())
	if err != nil {
		t.Fatalf("NewDeploymentWatcher() error = %v", err)
	}

	// Never started, so the cache cannot sync.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := w.WaitForSync(ctx); err == nil {
		t.Error("WaitForSync() on a cancelled context returned nil, want an error")
	}
}

func TestNamespaceLabel(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{in: "", want: "<all>"},
		{in: "default", want: "default"},
	}
	for _, tt := range tests {
		if got := namespaceLabel(tt.in); got != tt.want {
			t.Errorf("namespaceLabel(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
