# Ansible Configuration

This directory contains Ansible playbooks and configurations for automating the setup and provisioning of a Kubernetes development environment on Fedora CoreOS.

## Overview

The Ansible setup provides automated installation and configuration of:

- **Initial system setup** with required packages and dotfiles
- **Complete Kubernetes cluster** with all control plane components
- **Container runtime** (containerd) and networking (CNI)
- **PKI infrastructure** with automatic certificate generation

## Structure

```txt
ansible/
├── ansible.cfg          # Ansible configuration
├── inventory.yml        # Local inventory configuration
├── requirements.yml     # Required Ansible collections
├── init.yml            # Initial system setup playbook
├── provision.yml       # Main Kubernetes provisioning playbook
└── templates/          # Jinja2 templates for configs and services
```

## Prerequisites

- Podman machine running Fedora CoreOS
- Ansible installed on the host system
- Required Ansible collections (installed automatically)

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

## Available Tags

The main provisioning playbook supports these tags for selective execution:

- `setup` - Download and install Kubernetes binaries
- `certs` - Generate PKI certificates and authentication tokens
- `control-plane-core` - Configure etcd and kube-apiserver
- `control-plane-managers` - Configure scheduler and controller-manager
- `verify` - Verify cluster health and component status
- `kubelet` - Configure and start kubelet with containerd
- `leftovers` - Final verification and cleanup tasks

## Configuration

Key variables are defined in `provision.yml`:

```yaml
k8s_version: "1.30.0"
containerd_version: "2.1.2"
pod_network_cidr: "10.244.0.0/16"
service_cidr: "10.96.0.0/12"
```

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
