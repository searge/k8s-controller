# Kubernetes Devcontainer Setup

This devcontainer provisions a local, single-node Kubernetes control plane for controller development and API testing.

## What works

- Control plane components: etcd, kube-apiserver, kube-controller-manager, kube-scheduler
- Kubelet with containerd on the same node
- kubectl access to the API server
- RBAC, service accounts, and CRDs

## Getting started

Everything goes through [Task](https://taskfile.dev), which is the single
source of truth for this repository. Run the setup once to install the
components and generate the certificates:

```bash
task collections
task devcontainer
```

Start the cluster:

```bash
task devcontainer-run
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

## Partial runs

Both playbooks are tagged, so you can re-run only a part of them by
passing extra arguments through Task:

```bash
task devcontainer-run -- -t preflight
task devcontainer-run -- -t control-plane
task devcontainer-run -- -t verify
```

Available tags:

- `devcontainer.yml`: `install`, `certs`, `config`, `verify`, `summary`
- `devcontainer-run.yml`: `preflight`, `control-plane` (alias `start`),
  `verify` (alias `health`), `status` (alias `report`), and `debug`,
  which is also tagged `never` and has to be requested explicitly.

## What makes pods runnable

These used to be manual steps. They are handled automatically now, so no
extra commands are needed:

- `iptables` is installed by `task devcontainer`
  ("Install | Ensure iptables is available").
- The pause image is pre-pulled with a platform hint by
  `task devcontainer-run` ("Start | Pre-pull pause image for kubelet"),
  using the architecture detected in `ansible/group_vars/all.yml`.
- `--cgroupns=host` is already set in `runArgs` in
  `.devcontainer/devcontainer.json`, so cgroup v2 can be delegated.
- `use_local_image_pull = true` is set in the generated containerd config
  (`ansible/templates/containerd-config.toml.j2`) to avoid platform
  unpack errors.

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
task devcontainer-run -- -t control-plane
```

View logs:

```bash
tail -f /var/log/kubernetes/kubelet.log
tail -f /var/log/kubernetes/kube-apiserver.log
```
