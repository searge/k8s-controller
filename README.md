# Kubernetes Controller

My implementation of the Golang Kubernetes Controller course from FWDays.

## About

This project follows [the step-by-step tutorial](https://github.com/den-vasyliev/k8s-controller-tutorial-ref) for building production-grade Kubernetes controllers in Go. Each step is implemented as a separate commit/branch with detailed explanations.

**Course**: [Crash Course: Kubernetes controllers](https://fwdays.com/event/kubernetes-controllers-course)
**Instructors**: @den-vasyliev (Principal SRE), @Alex0M (Senior Platform Engineer)

## Progress

- [x] Golang CLI Application using Cobra
- [ ] Zerolog for structured logging
- [ ] pflag for CLI log level flags
- [ ] FastHTTP server command
- [ ] Makefile, Dockerfile, GitHub Workflow
- [ ] List Kubernetes Deployments with client-go
- [ ] Deployment Informer with client-go
- [ ] JSON API Endpoint for deployments
- [ ] controller-runtime Deployment Controller
- [ ] Leader Election and Metrics
- [ ] Custom Resource (FrontendPage CRD)
- [ ] Platform API (CRUD + Swagger)
- [ ] MCP Integration
- [ ] JWT Authentication
- [ ] OpenTelemetry Instrumentation

## Architecture

```mermaid
C4Container
    title Kubernetes Controller Architecture

    Person(user, "DevOps Engineer", "Uses CLI")

    System_Boundary(app, "Controller Application") {
        Container(cli, "CLI Client", "Go, Cobra", "Command line interface")
        Container(server, "HTTP Server", "Go, FastHTTP", "REST API and UI")
        Container(controller, "Controller", "Go, controller-runtime", "Reconciliation logic")
        Container(informers, "Informers", "Go, client-go", "Watch and cache")
    }

    System_Ext(k8s, "Kubernetes API", "Manages cluster resources")

    Rel(user, cli, "Uses")
    Rel(cli, server, "Commands")
    Rel(server, k8s, "API calls")
    Rel(k8s, informers, "Events")
    Rel(informers, controller, "Cached data")
    Rel(controller, k8s, "Reconcile")
```

## Quick Start

```bash
# Clone and setup
git clone https://github.com/Searge/k8s-controller.git
cd k8s-controller
go mod download

# Build and run
go build -o bin/controller main.go
./bin/controller --help
```

## Dependencies

- **CLI**: cobra, pflag, zerolog
- **HTTP**: fasthttp
- **Kubernetes**: client-go, controller-runtime
- **Observability**: OpenTelemetry, Prometheus metrics
- **Auth**: JWT tokens
- **Build**: Docker, GitHub Actions
