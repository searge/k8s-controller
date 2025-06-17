# Kubernetes Control Plane Setup

This guide helps you build a complete Kubernetes development environment from scratch to understand how all components work together.

> 📚 **Based on**: [k8sdiy-kubernetes-control-plane](https://github.com/den-vasyliev/k8sdiy-kubernetes-control-plane) by Denis Vasiliev

- [Kubernetes Control Plane Setup](#kubernetes-control-plane-setup)
  - [Why Build Your Own Control Plane?](#why-build-your-own-control-plane)
  - [Prerequisites](#prerequisites)
  - [Architecture Overview](#architecture-overview)
  - [Components We'll Install](#components-well-install)
  - [Step 1: Environment Setup](#step-1-environment-setup)
    - [Podman Machine Setup Guide for Fedora 41](#podman-machine-setup-guide-for-fedora-41)
      - [Clean Up Old Configuration](#clean-up-old-configuration)
      - [Create and Start Virtual Machine](#create-and-start-virtual-machine)
      - [Verify Installation](#verify-installation)
      - [Alternative: Machine Without Volume Mounting](#alternative-machine-without-volume-mounting)
      - [Usage Examples](#usage-examples)
    - [Install Basic Tools](#install-basic-tools)
  - [Step 2: Download Kubernetes Binaries](#step-2-download-kubernetes-binaries)
    - [Download Kubebuilder Tools](#download-kubebuilder-tools)
    - [Download Additional Components](#download-additional-components)
  - [Step 3: Generate Certificates and Tokens](#step-3-generate-certificates-and-tokens)
  - [Step 4: Configure kubectl](#step-4-configure-kubectl)
  - [Step 5: Start Core Components](#step-5-start-core-components)
    - [Start etcd](#start-etcd)
    - [Start kube-apiserver](#start-kube-apiserver)
  - [Step 6: Install Container Runtime](#step-6-install-container-runtime)
    - [Install containerd and CNI](#install-containerd-and-cni)
    - [Configure CNI Network](#configure-cni-network)
    - [Configure containerd](#configure-containerd)
  - [Step 7: Start Control Plane Components](#step-7-start-control-plane-components)
    - [Start kube-scheduler](#start-kube-scheduler)
    - [Configure kubelet](#configure-kubelet)
    - [Start kubelet](#start-kubelet)
    - [Start kube-controller-manager](#start-kube-controller-manager)
  - [Step 8: Verify Setup](#step-8-verify-setup)
    - [Check Component Status](#check-component-status)
    - [Test with a Pod](#test-with-a-pod)
  - [Built-in Controllers Overview](#built-in-controllers-overview)
    - [Core Controllers](#core-controllers)
    - [System Controllers](#system-controllers)
    - [Storage Controllers](#storage-controllers)
    - [Cleanup Controllers](#cleanup-controllers)
  - [Troubleshooting](#troubleshooting)
    - [Common Issues](#common-issues)
      - [etcd Connection Issues](#etcd-connection-issues)
      - [kubelet Problems](#kubelet-problems)
      - [API Server Issues](#api-server-issues)
    - [Useful Commands](#useful-commands)
  - [Next Steps](#next-steps)
  - [References](#references)

## Why Build Your Own Control Plane?

- **Deep Understanding**: Learn how K8s components interact
- **Development Environment**: Test controllers without minikube/kind
- **Debugging**: Direct access to all components and logs
- **Education**: See the inner workings of kubelet, api-server, etc.

## Prerequisites

- Fedora Linux or any other Linux distribution
- Podman installed
- Basic understanding of Kubernetes concepts
- Terminal with sudo privileges

## Architecture Overview

```txt
┌─────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   kubectl   │--->│  kube-apiserver │--->│      etcd       │
└─────────────┘    └─────────────────┘    └─────────────────┘
                            │
                            ▼
┌───────────────────────────────────────────────────────────┐
│                 kube-controller-manager                   │
│  • ReplicationController  • DeploymentController          │
│  • ServiceController      • NodeController                │
│  • EndpointsController    • NamespaceController           │
└───────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ kube-sched  │    │     kubelet     │--->│   containerd    │
│   uler      │    │                 │    │                 │
└─────────────┘    └─────────────────┘    └─────────────────┘
```

## Components We'll Install

| Component                   | Purpose                          | Port  |
| --------------------------- | -------------------------------- | ----- |
| **etcd**                    | Key-value store for cluster data | 2379  |
| **kube-apiserver**          | Kubernetes API server            | 6443  |
| **kube-controller-manager** | Built-in controllers             | -     |
| **kube-scheduler**          | Pod scheduler                    | -     |
| **kubelet**                 | Node agent                       | 10250 |
| **containerd**              | Container runtime                | -     |

## Step 1: Environment Setup

### Podman Machine Setup Guide for Fedora 41

> [!IMPORTANT]
> Setting up Podman machines on Fedora 41 can be tricky due to missing dependencies.
> If you're getting errors like `"could not find gvproxy"` or `"failed to find virtiofsd"`,
> here's the complete fix.

```bash
# Remove any existing Podman installation and reinstall with all dependencies:
sudo dnf remove podman &&
  sudo dnf install @container-management netavark gvisor-tap-vsock virtiofsd
```

The key issue is that virtiofsd is installed in `/usr/libexec/` but Podman looks for it
in `$PATH`, so create a symlink:

```bash
# Create a symlink so Podman can find virtiofsd
sudo ln -s /usr/libexec/virtiofsd /usr/bin/virtiofsd
```

#### Clean Up Old Configuration

If you had previous Podman installations:

```bash
# Remove old machines
podman machine reset -f

# Clean up old config (optional)
rm -rf ~/.config/containers/
```

#### Create and Start Virtual Machine

```bash
# Create a new virtual machine
podman machine init dev

# Start the virtual machine
podman machine start dev

# Test SSH connection to the VM
podman machine ssh dev
```

#### Verify Installation

```bash
# Check if machine is running
podman machine list

# Check Podman info
podman info
```

#### Alternative: Machine Without Volume Mounting

If you still have issues, try creating a machine without volume mounting:

```bash
podman machine rm dev --force
podman machine init dev --volume-driver none
podman machine start dev
```

This bypasses the virtiofsd requirement but still gives you a working container environment.

#### Usage Examples

```bash
# Run a container
podman run -it --rm fedora:latest /bin/bash

# List running containers
podman ps

# Stop the virtual machine
podman machine stop dev

# Start the virtual machine
podman machine start dev
```

### Install Basic Tools

By default you will get [Fedora CoreOS](https://fedoraproject.org/coreos/).
It is an automatically updating, immutable operating system designed for running containerized workloads securely and at scale.

```bash
➜  ~ hostnamectl
     Static hostname: (unset)
  Transient hostname: localhost
           Icon name: computer-vm
             Chassis: vm 🖴
          Machine ID: f12c5eecdXXX
             Boot ID: 063b2a74bXXX
      Virtualization: kvm
    Operating System: Fedora CoreOS 41.20250215.3.0
         CPE OS Name: cpe:/o:fedoraproject:fedora:41
      OS Support End: Mon 2025-12-15
OS Support Remaining: 5month 4w
              Kernel: Linux 6.12.13-200.fc41.x86_64
        Architecture: x86-64
     Hardware Vendor: QEMU
      Hardware Model: Standard PC _i440FX + PIIX, 1996_
    Firmware Version: 1.16.3-3.fc41
       Firmware Date: Tue 2014-04-01
        Firmware Age: 11y 2month 2w 2d
```

> [!IMPORTANT]
> That is, truly immutable. To make changes to the environment,
> you need to rewrite the system image using rpm-ostree and reboot.

But for the dev environment, we'll do some hacks:

```bash
sudo rpm-ostree install dnf zsh wget vim --allow-inactive

# Install Oh My Zsh
sh -c "$(curl -fsSL https://raw.github.com/ohmyzsh/ohmyzsh/master/tools/install.sh)"

# Install k9s for cluster management
curl -sS https://webi.sh/k9s | sh
```

## Step 2: Download Kubernetes Binaries

### Download Kubebuilder Tools

```bash
mkdir -p ./kubebuilder/bin && \
  curl -L https://storage.googleapis.com/kubebuilder-tools/kubebuilder-tools-1.30.0-linux-amd64.tar.gz \
    -o kubebuilder-tools.tar.gz && \
  tar -C ./kubebuilder --strip-components=1 -zvxf kubebuilder-tools.tar.gz && \
  rm kubebuilder-tools.tar.gz
```

### Download Additional Components

```bash
# For AMD64
curl -L "https://dl.k8s.io/v1.30.0/bin/linux/amd64/kubelet" -o kubebuilder/bin/kubelet
curl -L "https://dl.k8s.io/v1.30.0/bin/linux/amd64/kube-controller-manager" -o kubebuilder/bin/kube-controller-manager
curl -L "https://dl.k8s.io/v1.30.0/bin/linux/amd64/kube-scheduler" -o kubebuilder/bin/kube-scheduler

# Make binaries executable
chmod +x kubebuilder/bin/*
```

## Step 3: Generate Certificates and Tokens

```bash
# Generate service account key pair
openssl genrsa -out /tmp/sa.key 2048
openssl rsa -in /tmp/sa.key -pubout -out /tmp/sa.pub

# Generate token file
TOKEN="1234567890"
echo "${TOKEN},admin,admin,system:masters" > /tmp/token.csv

# Generate CA certificate for kubelet
openssl genrsa -out /tmp/ca.key 2048
openssl req -x509 -new -nodes -key /tmp/ca.key -subj "/CN=kubelet-ca" -days 365 -out /tmp/ca.crt
```

## Step 4: Configure kubectl

```bash
sudo kubebuilder/bin/kubectl config set-credentials test-user --token=1234567890
sudo kubebuilder/bin/kubectl config set-cluster test-env --server=https://127.0.0.1:6443 --insecure-skip-tls-verify
sudo kubebuilder/bin/kubectl config set-context test-context --cluster=test-env --user=test-user --namespace=default
sudo kubebuilder/bin/kubectl config use-context test-context
```

## Step 5: Start Core Components

### Start etcd

```bash
HOST_IP=$(hostname -I | awk '{print $1}')

kubebuilder/bin/etcd \
  --advertise-client-urls http://$HOST_IP:2379 \
  --listen-client-urls http://0.0.0.0:2379 \
  --data-dir ./etcd \
  --listen-peer-urls http://0.0.0.0:2380 \
  --initial-cluster default=http://$HOST_IP:2380 \
  --initial-advertise-peer-urls http://$HOST_IP:2380 \
  --initial-cluster-state new \
  --initial-cluster-token test-token &

# Verify etcd is running
curl http://127.0.0.1:2379/health
```

### Start kube-apiserver

```bash
sudo kubebuilder/bin/kube-apiserver \
  --etcd-servers=http://$HOST_IP:2379 \
  --service-cluster-ip-range=10.0.0.0/24 \
  --bind-address=0.0.0.0 \
  --secure-port=6443 \
  --advertise-address=$HOST_IP \
  --authorization-mode=AlwaysAllow \
  --token-auth-file=/tmp/token.csv \
  --enable-priority-and-fairness=false \
  --allow-privileged=true \
  --profiling=false \
  --storage-backend=etcd3 \
  --storage-media-type=application/json \
  --v=0 \
  --service-account-issuer=https://kubernetes.default.svc.cluster.local \
  --service-account-key-file=/tmp/sa.pub \
  --service-account-signing-key-file=/tmp/sa.key &

# Verify API server is ready
sudo kubebuilder/bin/kubectl get --raw='/readyz'
```

## Step 6: Install Container Runtime

### Install containerd and CNI

```bash
sudo mkdir -p /opt/cni/bin
sudo mkdir -p /etc/cni/net.d

# Download containerd (choose your architecture)
# For AMD64:
wget https://github.com/containerd/containerd/releases/download/v2.1.2/containerd-static-2.1.2-linux-amd64.tar.gz
sudo tar zxf containerd-static-2.1.2-linux-amd64.tar.gz -C /opt/cni/

# Download runc
sudo curl -L "https://github.com/opencontainers/runc/releases/download/v1.2.6/runc.amd64" -o /opt/cni/bin/runc
sudo chmod +x /opt/cni/bin/runc

# Download CNI plugins (choose your architecture)
# For AMD64:
wget https://github.com/containernetworking/plugins/releases/download/v1.6.2/cni-plugins-linux-amd64-v1.6.2.tgz
sudo tar zxf cni-plugins-linux-amd64-v1.6.2.tgz -C /opt/cni/bin/
```

### Configure CNI Network

```bash
cat <<EOF > 10-mynet.conf
{
  "cniVersion": "0.3.1",
  "name": "mynet",
  "type": "bridge",
  "bridge": "cni0",
  "isGateway": true,
  "ipMasq": true,
  "ipam": {
    "type": "host-local",
    "subnet": "10.22.0.0/16",
    "routes": [
      { "dst": "0.0.0.0/0" }
    ]
  }
}
EOF
sudo mv 10-mynet.conf /etc/cni/net.d/
```

### Configure containerd

```bash
sudo mkdir -p /etc/containerd/
cat <<EOF > config.toml
version = 2
[plugins]
  [plugins."io.containerd.grpc.v1.cri"]
    [plugins."io.containerd.grpc.v1.cri".containerd]
      snapshotter = "overlayfs"
      [plugins."io.containerd.grpc.v1.cri".containerd.default_runtime]
        runtime_type = "io.containerd.runc.v2"
        [plugins."io.containerd.grpc.v1.cri".containerd.default_runtime.options]
          SystemdCgroup = true
EOF
sudo mv config.toml /etc/containerd/config.toml

# Start containerd
export PATH=$PATH:/opt/cni/bin:kubebuilder/bin
sudo PATH=$PATH:/opt/cni/bin:/usr/sbin /opt/cni/bin/containerd -c /etc/containerd/config.toml &
```

## Step 7: Start Control Plane Components

### Start kube-scheduler

```bash
sudo kubebuilder/bin/kube-scheduler \
  --kubeconfig=/root/.kube/config \
  --leader-elect=false \
  --v=2 \
  --bind-address=0.0.0.0 &
```

### Configure kubelet

```bash
# Create kubelet directories
sudo mkdir -p /var/lib/kubelet
sudo mkdir -p /etc/kubernetes/manifests
sudo mkdir -p /var/log/kubernetes

# Copy certificates
sudo cp /tmp/ca.crt /var/lib/kubelet/ca.crt

# Create kubelet configuration
cat << EOF | sudo tee /var/lib/kubelet/config.yaml
apiVersion: kubelet.config.k8s.io/v1beta1
kind: KubeletConfiguration
authentication:
  anonymous:
    enabled: true
  webhook:
    enabled: true
  x509:
    clientCAFile: "/var/lib/kubelet/ca.crt"
authorization:
  mode: AlwaysAllow
clusterDomain: "cluster.local"
clusterDNS:
  - "10.0.0.10"
resolvConf: "/etc/resolv.conf"
runtimeRequestTimeout: "15m"
failSwapOn: false
seccompDefault: true
serverTLSBootstrap: true
containerRuntimeEndpoint: "unix:///run/containerd/containerd.sock"
staticPodPath: "/etc/kubernetes/manifests"
EOF

# Copy kubeconfig
sudo cp /root/.kube/config /var/lib/kubelet/kubeconfig
```

### Start kubelet

```bash
sudo PATH=$PATH:/opt/cni/bin:/usr/sbin kubebuilder/bin/kubelet \
  --kubeconfig=/var/lib/kubelet/kubeconfig \
  --config=/var/lib/kubelet/config.yaml \
  --root-dir=/var/lib/kubelet \
  --cert-dir=/var/lib/kubelet/pki \
  --hostname-override=$(hostname) \
  --pod-infra-container-image=registry.k8s.io/pause:3.10 \
  --node-ip=$HOST_IP \
  --cgroup-driver=cgroupfs \
  --max-pods=4 \
  --v=1 &

# Verify node is registered
sudo kubebuilder/bin/kubectl get nodes
```

### Start kube-controller-manager

```bash
# Create required service accounts and configmaps
export KUBECONFIG=~/.kube/config
cp /tmp/sa.pub /tmp/ca.crt
sudo kubebuilder/bin/kubectl create sa default
sudo kubebuilder/bin/kubectl create configmap kube-root-ca.crt --from-file=ca.crt=/tmp/ca.crt -n default

# Start controller manager
sudo PATH=$PATH:/opt/cni/bin:/usr/sbin kubebuilder/bin/kube-controller-manager \
  --kubeconfig=/var/lib/kubelet/kubeconfig \
  --leader-elect=false \
  --allocate-node-cidrs=true \
  --cluster-cidr=10.0.0.0/16 \
  --service-cluster-ip-range=10.0.0.0/24 \
  --cluster-name=kubernetes \
  --root-ca-file=/var/lib/kubelet/ca.crt \
  --service-account-private-key-file=/tmp/sa.key \
  --use-service-account-credentials=true \
  --v=2 &
```

## Step 8: Verify Setup

### Check Component Status

```bash
# Check all components
sudo kubebuilder/bin/kubectl get componentstatuses

# Check API server readiness
sudo kubebuilder/bin/kubectl get --raw='/readyz?verbose'

# Check nodes
sudo kubebuilder/bin/kubectl get nodes

# Check all resources
sudo kubebuilder/bin/kubectl get all -A
```

### Test with a Pod

```bash
# Deploy a test pod
sudo PATH=$PATH:/usr/sbin kubebuilder/bin/kubectl apply -f -<<EOF
apiVersion: v1
kind: Pod
metadata:
  name: test-pod-nginx
spec:
  containers:
  - name: test-container-nginx
    image: nginx:1.21
    securityContext:
      privileged: true
EOF

# Check pod status
sudo kubebuilder/bin/kubectl get pods

# List containers
sudo /opt/cni/bin/ctr -n k8s.io c ls

# Exec into container (replace with actual container ID)
sudo /opt/cni/bin/ctr -n k8s.io tasks exec -t --exec-id m <CONTAINER_ID> sh
```

## Built-in Controllers Overview

Your control plane now includes these built-in controllers:

### Core Controllers

- **ReplicationController**: Ensures specified number of pod replicas
- **Deployment Controller**: Manages Deployments via ReplicaSets
- **ReplicaSet Controller**: Ensures specified replicas for ReplicaSets
- **StatefulSet Controller**: Manages stateful applications
- **DaemonSet Controller**: Ensures pod runs on all/some nodes
- **Job Controller**: Manages batch/finite tasks
- **CronJob Controller**: Manages scheduled tasks

### System Controllers

- **Namespace Controller**: Handles namespace lifecycle
- **ServiceAccount Controller**: Manages ServiceAccount objects
- **Node Controller**: Monitors node health
- **Endpoints Controller**: Populates Endpoints for Services
- **Service Controller**: Manages Service objects
- **ResourceQuota Controller**: Enforces resource quotas
- **HorizontalPodAutoscaler Controller**: Scales pods based on metrics

### Storage Controllers

- **PersistentVolume Controller**: Manages PersistentVolumes
- **PersistentVolumeClaim Controller**: Binds PVCs to PVs

### Cleanup Controllers

- **Garbage Collector Controller**: Cleans up dependent objects
- **TTL Controller**: Cleans up finished Jobs and Pods

## Troubleshooting

### Common Issues

#### etcd Connection Issues

```bash
# Check etcd health
curl http://127.0.0.1:2379/health

# Check etcd logs
journalctl -u etcd
```

#### kubelet Problems

```bash
# Check kubelet status
systemctl status kubelet

# View kubelet logs
journalctl -u kubelet

# Verify containerd
systemctl status containerd
```

#### API Server Issues

```bash
# Check API server health
kubectl get --raw='/readyz'

# Check API server logs
journalctl -u kube-apiserver
```

### Useful Commands

```bash
# Check all pods in system namespaces
sudo kubebuilder/bin/kubectl get pods --all-namespaces

# Describe node for detailed info
sudo kubebuilder/bin/kubectl describe node $(hostname)

# Check events
sudo kubebuilder/bin/kubectl get events

# View controller manager logs
journalctl -u kube-controller-manager
```

## Next Steps

Now that you have a working Kubernetes control plane:

1. **Experiment with Deployments**: Create different resource types
2. **Study Controller Logs**: See how built-in controllers work
3. **Build Custom Controllers**: Use this as a foundation for your own controllers
4. **Debug Resource Issues**: Practice troubleshooting in a controlled environment

## References

- **Controller Names**: [controller_names.go](https://github.com/kubernetes/kubernetes/blob/master/cmd/kube-controller-manager/names/controller_names.go)
- **Controller Source Code**: [pkg/controller](https://github.com/kubernetes/kubernetes/tree/master/pkg/controller)
- **Original Tutorial**: [k8sdiy-kubernetes-control-plane](https://github.com/den-vasyliev/k8sdiy-kubernetes-control-plane)

---

**🎉 Congratulations!** You now have a fully functional Kubernetes control plane running locally. This gives you deep insight into how K8s works internally and provides an excellent foundation for building and testing your own controllers.
