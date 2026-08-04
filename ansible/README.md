# Ansible Configuration

This directory contains Ansible playbooks and configurations for automating the setup and provisioning of a Kubernetes development environment on Fedora CoreOS.

## Overview

The Ansible setup provides automated installation and configuration of:

- **Initial system setup** with required packages and dotfiles
- **Complete Kubernetes cluster** with all control plane components
- **Container runtime** (containerd) and networking (CNI)
- **PKI infrastructure** with automatic certificate generation
- **Dev Container support** so the same cluster runs inside Codespaces

## Structure

```txt
ansible/
├── ansible.cfg           # Ansible configuration
├── inventory.yml         # Local inventory configuration
├── requirements.yml      # Required Ansible collections
├── group_vars/all.yml    # Shared vars: versions, CIDRs, paths
├── init.yml              # Initial system setup playbook
├── provision.yml         # Podman machine provisioning
├── devcontainer.yml      # Devcontainer installation
├── devcontainer-run.yml  # Devcontainer control plane start
└── templates/            # Jinja2 templates for configs and services
```

## Prerequisites

Podman machine path:

- Podman machine running Fedora CoreOS
- Ansible installed on the host system
- Required Ansible collections (installed automatically)

Dev Container path:

- Repository opened in the Dev Container
  (`.devcontainer/devcontainer.json`)
- `ansible-core` and the collections are installed by `postCreateCommand`

## Quick Start

### 1. Initialize the Environment

Set up a new Podman machine and run initial configuration:

```bash
task init && task ssh -- 'cd /srv/app && go-task provision'
```

This command:

- Creates and starts a Podman machine
- Runs the initial setup playbook
- Provisions the complete Kubernetes cluster

### 2. Run Specific Provisioning Tags

Execute specific parts of the provisioning process:

```bash
# Run only PKI certificate generation
task ssh -- 'cd /srv/app && go-task provision -- "-t certs"'

# Run only control plane setup
task ssh -- 'cd /srv/app && go-task provision -- "-t control-plane-core"'

# Run only verification tasks
task ssh -- 'cd /srv/app && go-task provision -- "-t leftovers"'
```

### 3. Access the Environment

SSH into the machine to interact with the cluster:

```bash
task ssh
kubectl get nodes
kubectl get all -A
```

### 4. Dev Container / Codespaces

Inside a Dev Container there is no Podman machine, so the playbooks run locally:

```bash
# Install etcd, kube-apiserver, kubelet, containerd, runc and CNI plugins
task devcontainer

# Generate PKI, render the unit files and start every component
task devcontainer-run
```

Extra Ansible arguments are passed through, so tags work the same way:

```bash
task devcontainer -- '-t install'
task devcontainer-run -- '-t status'
```

Helper scripts are installed into the container home directory:

```bash
~/k8s-status.sh   # state of every component
~/k8s-stop.sh     # stop the control plane
```

## Available Tags

The main provisioning playbook supports these tags for selective execution:

- `setup` - Download and install Kubernetes binaries
- `certs` - Generate PKI certificates and authentication tokens
- `control-plane-core` - Configure etcd and kube-apiserver
- `control-plane-managers` - Configure scheduler and controller-manager
- `verify` - Verify cluster health and component status
- `kubelet` - Configure and start kubelet with containerd
- `leftovers` - Final verification and cleanup tasks

The devcontainer playbooks have their own tags: `install`, `certs`,
`config`, `verify` and `summary` for `devcontainer.yml`; `preflight`,
`start`, `control-plane`, `health`, `status` and `report` for
`devcontainer-run.yml`.

## Configuration

Shared variables live in `group_vars/all.yml` and are read by every playbook:

```yaml
k8s_version: "1.30.0"
containerd_version: "2.1.2"
runc_version: "1.2.6"
cni_version: "1.6.2"
cluster_cidr: "10.216.0.0/16"
service_cidr: "10.216.224.0/24"
pod_network_cidr: "10.144.0.0/16"
```

`etcd`, `kube-apiserver` and `kubectl` come from the `controller-tools`
`envtest` releases; containerd, runc and the CNI plugins come from their own
upstream releases.

## Logging

All Ansible execution logs are written to `/var/tmp/ansible.log` for troubleshooting and audit purposes.

## Architecture

The playbook creates a single-node Kubernetes cluster with:

- **etcd** - Key-value store for cluster data
- **kube-apiserver** - Kubernetes API server with TLS
- **kube-controller-manager** - Built-in controllers
- **kube-scheduler** - Pod scheduling
- **kubelet** - Node agent with containerd runtime
- **CNI networking** - Bridge-based pod networking

Perfect for development, testing, and learning Kubernetes internals!
