package informer

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
)

// newUnstartedWatcher builds a watcher without running it, for testing the event
// handlers directly.
//
// Calling the handlers by hand rather than through a fake watch is deliberate. A
// FakeWatcher can deliver an add, an update and a delete, but it cannot deliver a
// DeletedFinalStateUnknown: the informer synthesises those internally when a re-list
// finds an object gone that the watch never reported. Driving onDelete directly is the
// only way to cover that branch.
func newUnstartedWatcher(t *testing.T) *DeploymentWatcher {
	t.Helper()

	w, err := NewDeploymentWatcher(fake.NewClientset(), Config{}, testLogger())
	if err != nil {
		t.Fatalf("NewDeploymentWatcher() error = %v", err)
	}
	return w
}

func TestOnDeleteUnwrapsTombstone(t *testing.T) {
	tests := []struct {
		name      string
		obj       any
		wantCount EventCounts
	}{
		{
			name:      "plain deployment",
			obj:       deployment("default", "api", "1", 1),
			wantCount: EventCounts{Deleted: 1},
		},
		{
			name: "tombstone wrapping a deployment",
			obj: cache.DeletedFinalStateUnknown{
				Key: "default/api",
				Obj: deployment("default", "api", "1", 1),
			},
			wantCount: EventCounts{Deleted: 1},
		},
		{
			name:      "tombstone wrapping something else",
			obj:       cache.DeletedFinalStateUnknown{Key: "default/api", Obj: "not a deployment"},
			wantCount: EventCounts{Malformed: 1},
		},
		{
			name:      "bare unexpected type",
			obj:       "not a deployment",
			wantCount: EventCounts{Malformed: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := newUnstartedWatcher(t)
			w.onDelete(tt.obj)

			if got := w.EventCounts(); got != tt.wantCount {
				t.Errorf("after onDelete, EventCounts() = %+v, want %+v", got, tt.wantCount)
			}
		})
	}
}

func TestOnAddCountsOnlyDeployments(t *testing.T) {
	w := newUnstartedWatcher(t)

	w.onAdd(deployment("default", "api", "1", 1))
	w.onAdd("not a deployment")

	// A typed nil pointer satisfies the type assertion and then panics on the first
	// field access. This line crashed the handler until asDeployment grew a nil check,
	// so it stays as the regression test for that.
	w.onAdd((*appsv1.Deployment)(nil))

	want := EventCounts{Added: 1, Malformed: 2}
	if got := w.EventCounts(); got != want {
		t.Errorf("EventCounts() = %+v, want %+v", got, want)
	}
}

func TestOnUpdateRequiresBothObjects(t *testing.T) {
	tests := []struct {
		name      string
		oldObj    any
		newObj    any
		wantCount EventCounts
	}{
		{
			name:      "both deployments",
			oldObj:    deployment("default", "api", "1", 1),
			newObj:    deployment("default", "api", "2", 3),
			wantCount: EventCounts{Updated: 1},
		},
		{
			name:      "old is wrong type",
			oldObj:    "nope",
			newObj:    deployment("default", "api", "2", 3),
			wantCount: EventCounts{Malformed: 1},
		},
		{
			name:      "new is wrong type",
			oldObj:    deployment("default", "api", "1", 1),
			newObj:    "nope",
			wantCount: EventCounts{Malformed: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := newUnstartedWatcher(t)
			w.onUpdate(tt.oldObj, tt.newObj)

			if got := w.EventCounts(); got != tt.wantCount {
				t.Errorf("EventCounts() = %+v, want %+v", got, tt.wantCount)
			}
		})
	}
}

// TestEventCountsTrackWatchEvents ties the counters to real informer delivery, so the
// handlers are known to be wired in rather than merely correct when called by hand.
func TestEventCountsTrackWatchEvents(t *testing.T) {
	client, fakeWatch := newWatchedFakeClient()
	w, stop := startWatcher(t, client, Config{})
	defer stop()

	fakeWatch.Add(deployment("default", "api", "1", 1))
	waitForNames(t, w, "", []string{"default/api"})

	fakeWatch.Modify(deployment("default", "api", "2", 5))
	waitForReplicas(t, w, "default", "api", 5)

	fakeWatch.Delete(deployment("default", "api", "2", 5))
	waitForNames(t, w, "", nil)

	got := w.EventCounts()
	if got.Added != 1 || got.Updated != 1 || got.Deleted != 1 {
		t.Errorf("EventCounts() = %+v, want one of each add, update and delete", got)
	}
	if got.Malformed != 0 {
		t.Errorf("EventCounts().Malformed = %d, want 0", got.Malformed)
	}
}
