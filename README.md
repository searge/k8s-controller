# Kubernetes Controller

My implementation of the Golang Kubernetes Controller course from FWDays.

![Visitor](https://visitor-badge.laobi.icu/badge?page_id=Searge.k8s-controller)
[![Go Reference](https://pkg.go.dev/badge/github.com/Searge/k8s-controller.svg?style=flat-square)](https://pkg.go.dev/github.com/Searge/k8s-controller)
[![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/Searge/k8s-controller/go.yml?branch=main&style=flat-square&logo=githubactions&logoColor=white&label=test-n-build)](https://github.com/Searge/k8s-controller/actions/workflows/go.yml)
![Repo size](https://img.shields.io/github/repo-size/Searge/k8s-controller?style=flat-square)
[![Updates](https://img.shields.io/github/last-commit/Searge/k8s-controller.svg?style=flat-square&logo=git&logoColor=white&color=blue)](https://github.com/Searge/k8s-controller/commits/main/)

## About

A read-only tool for asking a Kubernetes cluster questions that are awkward to answer with
`kubectl`. Every question is a plugin; the plugins that need an API the cluster does not have stay
silently disabled, so the tool is worth the same in any environment.

It is also a learning project, and that is the part driving the order of work. It is built
alongside the [FWDays crash course on Kubernetes
controllers](https://fwdays.com/event/kubernetes-controllers-course), taught by @den-vasyliev and
@Alex0M, whose [reference
implementation](https://github.com/den-vasyliev/k8s-controller-tutorial-ref) runs to fourteen step
branches. The course's mechanisms are the architecture here; what they are pointed at is different.

**Step 8 of 14.** The CLI, structured logging, `list deployments` through client-go, a cached
Deployment informer, and an HTTP server that answers from that cache. Plugins, reconciliation and
controller-runtime are not started yet.

## Documentation

Start at [docs/README.md](docs/README.md). It carries the reading order and explains what is
deliberately kept out of this repository.

| Document | What is in it |
| --- | --- |
| [docs/COURSE.md](docs/COURSE.md) | Which course steps are done, and where this diverges from the reference implementation |
| [docs/ROADMAP.md](docs/ROADMAP.md) | What comes next, effort estimates, known defects left unfixed on purpose |
| [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) | Current and target layout, package boundaries, testing strategy |
| [docs/decisions/](docs/decisions/) | Numbered decisions with reasoning and rejected alternatives |
| [docs/backup-check.md](docs/backup-check.md) | First domain: Longhorn object graph, the join, what the assertion does not claim |
| [docs/api.md](docs/api.md) | HTTP endpoints and CLI commands as they exist today |
| [docs/CODING_GUIDELINES.md](docs/CODING_GUIDELINES.md) | Rules, quick reference |
| [docs/BEST_PRACTICES.md](docs/BEST_PRACTICES.md) | The same rules with worked examples |

## Quick Start

You need [Go 1.26+](https://golang.org/doc/install) and
[Taskfile](https://taskfile.dev/installation/). For the cluster, either
[Podman](https://podman.io/getting-started/installation) or Docker will do. If you would rather
not install any of it, skip to [Dev Container / Codespaces](#dev-container--codespaces).

```bash
git clone https://github.com/Searge/k8s-controller.git
cd k8s-controller

# Create the Podman machine and provision Kubernetes inside it
task init && task ssh -- 'cd /srv/app && go-task provision'

# Use the cluster
task ssh
kubectl get nodes
```

That gives you a Podman machine running Fedora CoreOS with a single-node Kubernetes v1.36.2
cluster inside it: etcd, API server, scheduler and controller-manager, kubelet on containerd, CNI
bridge networking, and a generated PKI. The Ansible that does all of this, including the tags for
running only part of it and where to look when it breaks, is documented in
[ansible/README.md](ansible/README.md).

### Dev Container / Codespaces

The same cluster runs inside a Dev Container, so no Podman machine is needed. Open the repository
in VS Code (*Reopen in Container*) or in
[GitHub Codespaces](https://github.com/features/codespaces), then:

```bash
task devcontainer      # install the binaries, generate the PKI, write the configs
task devcontainer-run  # start containerd, etcd, the control plane and kubelet
kubectl get nodes
```

Two helpers land in the container home directory: `~/k8s-status.sh` reports the state of every
component and `~/k8s-stop.sh` stops the control plane. Both paths read the same
`ansible/group_vars/all.yml`, so versions, CIDRs and paths match the Podman setup.

## Tasks

`task` on its own lists everything. The ones worth knowing:

```bash
task dev          # format, lint, test, build
task test-watch   # tests in watch mode
task docker-build # build the container image

task init         # create and set up the Podman machine
task ssh          # shell into it
task provision    # run the Ansible provisioning
task reboot       # restart the machine
task rm           # delete the machine
```

## Layout

```text
ansible/        cluster provisioning (see ansible/README.md)
cmd/            CLI commands
docs/           project documentation (start at docs/README.md)
internal/       application internals: domain logic, informers, controllers
notebooks/      Go learning notebooks
pkg/            client-go access, logging, HTTP server
scripts/        release tooling, smoke tests, Podman machine setup
.devcontainer/  Dev Container and Codespaces definition
```

Package boundaries and where each of those is going: [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).

## License

GNU General Public License v3.0. See [LICENSE](LICENSE).
