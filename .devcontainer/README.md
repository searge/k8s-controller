# Kubernetes Devcontainer Setup

This devcontainer provisions a local, single-node Kubernetes control plane for controller development and API testing.

## What works

- Control plane components: etcd, kube-apiserver, kube-controller-manager, kube-scheduler
- Kubelet with containerd on the same node
- kubectl access to the API server
- RBAC, service accounts, and CRDs

## Getting started

Run the setup once to install components and generate certs:

```bash
task collections
cd ansible/
ansible-playbook devcontainer.yml
```

Start the cluster:

```bash
ansible-playbook devcontainer-run.yml
```

Verify:

```bash
kubectl get nodes
kubectl get all -A
kubectl cluster-info
```

Check or stop services:

```bash
~/k8s-status.sh
~/k8s-stop.sh
```

## Requirements for pods to start

- `iptables` must be installed and usable.
- The pause image must be pulled with a platform hint:
  `sudo ctr -n k8s.io images pull --platform linux/amd64 registry.k8s.io/pause:3.10` (use `linux/arm64` on ARM)
- Codespaces/devcontainer should run with `--cgroupns=host` so cgroup v2 can be delegated.
- Containerd should use local image pull to avoid platform unpack errors.

## Common paths

- Logs: `/var/log/kubernetes/`
- Configs: `/etc/kubernetes/`
- Data: `/var/lib/etcd/`, `/var/lib/kubelet/`
- Binaries: `/usr/local/bin/`

## Troubleshooting

Check running processes:

```bash
ps aux | grep -E "(etcd|kube-|containerd)"
~/k8s-status.sh
```

Restart control plane and kubelet:

```bash
~/k8s-stop.sh
ansible-playbook devcontainer-run.yml -t control-plane,worker
```

View logs:

```bash
tail -f /var/log/kubernetes/kubelet.log
tail -f /var/log/kubernetes/kube-apiserver.log
```
