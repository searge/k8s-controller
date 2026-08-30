package informer

import (
	"fmt"
	"sort"
	"testing"
	"time"
)

// pollInterval is how often the wait helpers re-read the cache. The informer delivers
// watch events asynchronously, so a test cannot assert on the cache immediately after
// sending one; it has to wait for the state it expects, with a deadline.
const pollInterval = 5 * time.Millisecond

// listNames returns the cached Deployments as sorted "namespace/name" strings.
func listNames(t *testing.T, w *DeploymentWatcher, namespace string) []string {
	t.Helper()

	deployments, err := w.List(namespace, nil)
	if err != nil {
		t.Fatalf("List(%q) error = %v", namespace, err)
	}

	names := make([]string, 0, len(deployments))
	for _, d := range deployments {
		names = append(names, d.Namespace+"/"+d.Name)
	}
	sort.Strings(names)

	return names
}

// waitForNames blocks until the cache holds exactly want, or fails after syncTimeout.
//
// Reporting the last state seen matters: "timed out" alone does not distinguish an
// event that never arrived from one that arrived with the wrong content.
func waitForNames(t *testing.T, w *DeploymentWatcher, namespace string, want []string) {
	t.Helper()

	wantSorted := make([]string, len(want))
	copy(wantSorted, want)
	sort.Strings(wantSorted)

	var last []string
	waitFor(t, fmt.Sprintf("cache to hold %v", wantSorted), func() bool {
		last = listNames(t, w, namespace)
		return equalStrings(last, wantSorted)
	}, func() string {
		return fmt.Sprintf("cache holds %v, want %v", last, wantSorted)
	})
}

// waitForReplicas blocks until the cached Deployment reports want replicas.
//
// This is the assertion that separates a real update from a name that merely stayed
// present: an update handler that dropped the new object would still list the name.
func waitForReplicas(t *testing.T, w *DeploymentWatcher, namespace, name string, want int32) {
	t.Helper()

	var got string
	waitFor(t, fmt.Sprintf("%s/%s to report %d replicas", namespace, name, want), func() bool {
		deployments, err := w.List(namespace, nil)
		if err != nil {
			t.Fatalf("List(%q) error = %v", namespace, err)
		}
		for _, d := range deployments {
			if d.Name != name {
				continue
			}
			if d.Spec.Replicas == nil {
				got = "<nil>"
				return false
			}
			got = fmt.Sprint(*d.Spec.Replicas)
			return *d.Spec.Replicas == want
		}
		got = "<absent>"
		return false
	}, func() string {
		return fmt.Sprintf("%s/%s has %s replicas, want %d", namespace, name, got, want)
	})
}

// waitFor polls until condition holds or syncTimeout elapses, failing with detail().
func waitFor(t *testing.T, what string, condition func() bool, detail func() string) {
	t.Helper()

	deadline := time.Now().Add(syncTimeout)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(pollInterval)
	}

	t.Fatalf("timed out after %v waiting for %s: %s", syncTimeout, what, detail())
}

// equalStrings reports whether two sorted slices hold the same elements.
func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
