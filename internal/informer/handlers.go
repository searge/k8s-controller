package informer

import (
	"fmt"
	"sync/atomic"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/tools/cache"
)

// eventCounters records what the handlers actually observed.
//
// This is not yet a metrics exporter, and it is not decoration: without it the event
// handlers only write log lines, so the code that unwraps a tombstone in onDelete
// cannot be verified by any test. A counter makes that path observable now and is the
// value a real metric will read later.
type eventCounters struct {
	added   atomic.Int64
	updated atomic.Int64
	deleted atomic.Int64
	// malformed counts events whose payload was not a Deployment. Non-zero means
	// either a bug here or an informer delivering something unexpected; both are
	// worth seeing rather than silently ignoring.
	malformed atomic.Int64
}

// EventCounts is a snapshot of the handler counters.
type EventCounts struct {
	Added     int64
	Updated   int64
	Deleted   int64
	Malformed int64
}

// EventCounts returns a snapshot of how many events the handlers have processed.
// The four fields are read independently, so a snapshot taken while events are
// arriving is approximate rather than a consistent instant.
func (w *DeploymentWatcher) EventCounts() EventCounts {
	return EventCounts{
		Added:     w.counts.added.Load(),
		Updated:   w.counts.updated.Load(),
		Deleted:   w.counts.deleted.Load(),
		Malformed: w.counts.malformed.Load(),
	}
}

// registerHandlers attaches the event handlers to the informer.
//
// AddEventHandler returns an error rather than panicking as it once did, and the
// error is real: registering on an already-stopped informer fails.
func (w *DeploymentWatcher) registerHandlers() error {
	_, err := w.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    w.onAdd,
		UpdateFunc: w.onUpdate,
		DeleteFunc: w.onDelete,
	})
	if err != nil {
		return fmt.Errorf("failed to register deployment event handlers: %w", err)
	}
	return nil
}

// asDeployment asserts obj is a usable Deployment.
//
// The nil check is not redundant with the type assertion: a (*appsv1.Deployment)(nil)
// satisfies the assertion and then panics on first field access. The informer does not
// deliver one, but the handlers are exported to the package's tests and the check costs
// a comparison.
func asDeployment(obj any) (*appsv1.Deployment, bool) {
	deployment, ok := obj.(*appsv1.Deployment)
	if !ok || deployment == nil {
		return nil, false
	}
	return deployment, true
}

// onAdd logs a Deployment entering the cache.
func (w *DeploymentWatcher) onAdd(obj any) {
	deployment, ok := asDeployment(obj)
	if !ok {
		w.counts.malformed.Add(1)
		w.logger.Warn().Type("got", obj).Msg("Add event carried an unexpected type")
		return
	}
	w.counts.added.Add(1)
	w.logger.Debug().
		Str("deployment", deployment.Name).
		Str("namespace", deployment.Namespace).
		Msg("Deployment added")
}

// onUpdate logs a Deployment changing.
//
// The informer delivers an update on every resync even when nothing changed, so the
// resource version is logged to tell a real change from a replay.
func (w *DeploymentWatcher) onUpdate(oldObj, newObj any) {
	oldDeployment, okOld := asDeployment(oldObj)
	newDeployment, okNew := asDeployment(newObj)
	if !okOld || !okNew {
		w.counts.malformed.Add(1)
		w.logger.Warn().Msg("Update event carried an unexpected type")
		return
	}
	w.counts.updated.Add(1)
	w.logger.Debug().
		Str("deployment", newDeployment.Name).
		Str("namespace", newDeployment.Namespace).
		Str("old_version", oldDeployment.ResourceVersion).
		Str("new_version", newDeployment.ResourceVersion).
		Msg("Deployment updated")
}

// onDelete logs a Deployment leaving the cache.
//
// A delete may arrive wrapped in DeletedFinalStateUnknown when the watch missed the
// event and the object was gone by the time the informer re-listed. Unwrapping it is
// not optional: the type assertion fails otherwise and the deletion goes unnoticed.
func (w *DeploymentWatcher) onDelete(obj any) {
	if tombstone, ok := obj.(cache.DeletedFinalStateUnknown); ok {
		obj = tombstone.Obj
	}

	deployment, ok := asDeployment(obj)
	if !ok {
		w.counts.malformed.Add(1)
		w.logger.Warn().Type("got", obj).Msg("Delete event carried an unexpected type")
		return
	}
	w.counts.deleted.Add(1)
	w.logger.Debug().
		Str("deployment", deployment.Name).
		Str("namespace", deployment.Namespace).
		Msg("Deployment deleted")
}
