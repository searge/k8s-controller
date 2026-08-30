// Package server contains tests for the HTTP server functionality.
// This file tests the endpoint handlers against a fake deployment source.
package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/valyala/fasthttp"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// fakeSource implements DeploymentSource for handler tests. It records the
// namespaces List was asked for, so a test can assert the query parameter
// actually reaches the cache lookup.
type fakeSource struct {
	synced      bool
	deployments []*appsv1.Deployment
	listErr     error
	listedNS    []string
}

func (f *fakeSource) List(namespace string, _ labels.Selector) ([]*appsv1.Deployment, error) {
	f.listedNS = append(f.listedNS, namespace)
	if f.listErr != nil {
		return nil, f.listErr
	}
	if namespace == "" {
		return f.deployments, nil
	}
	var out []*appsv1.Deployment
	for _, d := range f.deployments {
		if d.Namespace == namespace {
			out = append(out, d)
		}
	}
	return out, nil
}

func (f *fakeSource) HasSynced() bool { return f.synced }

// fakeDeployment builds a minimal Deployment for projection through the handler.
func fakeDeployment(namespace, name string, replicas int32) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name},
		Spec:       appsv1.DeploymentSpec{Replicas: &replicas},
	}
}

// do runs one request through the full handler chain, logging into logBuf.
func do(source DeploymentSource, method, uri string, logBuf *bytes.Buffer) *fasthttp.RequestCtx {
	handler := newHandler(source, zerolog.New(logBuf))
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI(uri)
	ctx.Request.Header.SetMethod(method)
	handler(ctx)
	return ctx
}

func TestEndpoints(t *testing.T) {
	synced := &fakeSource{synced: true, deployments: []*appsv1.Deployment{
		fakeDeployment("default", "api", 3),
		fakeDeployment("other", "worker", 1),
	}}
	unsynced := &fakeSource{synced: false}

	tests := []struct {
		name       string
		source     DeploymentSource
		uri        string
		wantStatus int
		wantInBody string
	}{
		{"healthz is alive", synced, "/healthz", 200, `"ok"`},
		{"healthz ignores sync state", unsynced, "/healthz", 200, `"ok"`},
		{"readyz after sync", synced, "/readyz", 200, `"ok"`},
		{"readyz before sync", unsynced, "/readyz", 503, "cache not synced"},
		{"deployments before sync refuse", unsynced, "/deployments", 503, "cache not synced yet"},
		{"deployments after sync", synced, "/deployments", 200, `"count":2`},
		{"deployments filtered by namespace", synced, "/deployments?namespace=other", 200, `"count":1`},
		{"unknown path", synced, "/nope", 404, "not found"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logBuf bytes.Buffer
			ctx := do(tt.source, "GET", tt.uri, &logBuf)

			if got := ctx.Response.StatusCode(); got != tt.wantStatus {
				t.Errorf("GET %s status = %d, want %d", tt.uri, got, tt.wantStatus)
			}
			if body := string(ctx.Response.Body()); !strings.Contains(body, tt.wantInBody) {
				t.Errorf("GET %s body = %q, want it to contain %q", tt.uri, body, tt.wantInBody)
			}
		})
	}
}

func TestDeploymentsBody(t *testing.T) {
	source := &fakeSource{synced: true, deployments: []*appsv1.Deployment{
		fakeDeployment("default", "api", 3),
	}}

	var logBuf bytes.Buffer
	ctx := do(source, "GET", "/deployments", &logBuf)

	var got deploymentList
	if err := json.Unmarshal(ctx.Response.Body(), &got); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if got.Count != 1 || len(got.Items) != 1 {
		t.Fatalf("got count=%d items=%d, want 1 and 1", got.Count, len(got.Items))
	}

	item := got.Items[0]
	if item.Name != "api" || item.Namespace != "default" || item.Replicas.Desired != 3 {
		t.Errorf("projected item = %+v, want api/default with 3 desired replicas", item)
	}

	// The namespace parameter must reach the cache lookup as given.
	if len(source.listedNS) != 1 || source.listedNS[0] != "" {
		t.Errorf("List was called with namespaces %v, want one call with %q", source.listedNS, "")
	}
}

func TestDeploymentsListError(t *testing.T) {
	source := &fakeSource{synced: true, listErr: errors.New("index corrupted")}

	var logBuf bytes.Buffer
	ctx := do(source, "GET", "/deployments", &logBuf)

	if got := ctx.Response.StatusCode(); got != 500 {
		t.Errorf("status = %d, want 500", got)
	}
	// The real error goes to the log, not to the client.
	if body := string(ctx.Response.Body()); strings.Contains(body, "index corrupted") {
		t.Errorf("internal error leaked to the client: %q", body)
	}
	if !strings.Contains(logBuf.String(), "index corrupted") {
		t.Error("internal error missing from the log")
	}
}

func TestRequestLogging(t *testing.T) {
	source := &fakeSource{synced: true}

	var logBuf bytes.Buffer
	ctx := do(source, "GET", "/readyz", &logBuf)

	requestID := string(ctx.Response.Header.Peek("X-Request-ID"))
	if len(requestID) != 16 {
		t.Fatalf("X-Request-ID = %q, want 16 hex chars", requestID)
	}

	var entry struct {
		RequestID string `json:"request_id"`
		Method    string `json:"method"`
		Path      string `json:"path"`
		Status    int    `json:"status"`
	}
	line := logBuf.String()
	if err := json.Unmarshal([]byte(line), &entry); err != nil {
		t.Fatalf("log line is not structured JSON: %v (%q)", err, line)
	}

	if entry.RequestID != requestID {
		t.Errorf("log request_id = %q, header = %q: cannot correlate them", entry.RequestID, requestID)
	}
	if entry.Method != "GET" || entry.Path != "/readyz" || entry.Status != 200 {
		t.Errorf("log entry = %+v, want GET /readyz 200", entry)
	}
}

func TestRequestIDsAreUnique(t *testing.T) {
	source := &fakeSource{synced: true}
	seen := map[string]bool{}

	for i := range 100 {
		var logBuf bytes.Buffer
		ctx := do(source, "GET", "/healthz", &logBuf)
		id := string(ctx.Response.Header.Peek("X-Request-ID"))
		if seen[id] {
			t.Fatalf("request id %q repeated within %d requests", id, i+1)
		}
		seen[id] = true
	}
}

// TestOldHealthEndpointGone pins the removal: /health was replaced by /healthz
// and /readyz, and a monitor still probing it must get an unambiguous 404, not
// a greeting page with a 200.
func TestOldHealthEndpointGone(t *testing.T) {
	source := &fakeSource{synced: true}

	var logBuf bytes.Buffer
	ctx := do(source, "GET", "/health", &logBuf)

	if got := ctx.Response.StatusCode(); got != 404 {
		t.Errorf("GET /health = %d, want 404 after the healthz/readyz split", got)
	}
}
