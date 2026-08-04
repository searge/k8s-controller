# k8s runtime smoke tests

Two checks that catch the class of breakage a `k8s_version` bump causes: flags and
config keys that Kubernetes removed upstream, and a runtime stack that no longer
starts pods. Both read their versions from `ansible/group_vars/all.yml`, so they
always test what the playbooks will actually install.

## check_flags.sh

Static, no containers, no privileges, ~30s. Downloads the target-version control
plane binaries and reports every flag the playbooks pass that the binary no longer
accepts. Also validates the rendered `KubeletConfiguration` files and the
containerd config against the real parsers.

```bash
scripts/k8s/check_flags.sh
```

Run this after every `k8s_version` change. Going 1.30.0 -> 1.36.2 removed
`kube-apiserver --cloud-provider` and `kubelet --pod-infra-container-image`; both
would have surfaced only as a health-wait timeout in `task devcontainer-run`.

## smoke_test.sh

Brings up containerd plus kubelet in standalone mode — no etcd, no apiserver, no
PKI — inside a privileged container, then checks whether two static pods reach
Running. One uses `hostNetwork: true` and exercises containerd and runc; the other
goes through the CNI bridge and must get an IP from `pod_network_cidr`.

```bash
sudo bash scripts/k8s/smoke_test.sh
```

**Rootful podman is required.** Rootless cannot create cgroups under
`/sys/fs/cgroup` and kubelet additionally dies on `open /dev/kmsg: operation not
permitted` inside a user namespace.

Results are written to `scripts/k8s/tmp/out/` (`run.log`, `exit-code`,
`logs/containerd.log`, `logs/kubelet.log`), so a failing run can be read after the
container is gone. Exit code is 0 on pass, 2 if only one pod came up, 1 on failure.

## What these do not cover

- The full control plane together (etcd + apiserver + scheduler +
  controller-manager). `check_flags.sh` validates that path's flags statically;
  only `task devcontainer-run` or `task provision` exercises it for real.
- The devcontainer base image. The smoke test uses `debian:trixie-slim` plus
  `iptables`, not `mcr.microsoft.com/devcontainers/python:3.12`.
- Two kubelet config keys are forced off because there is no apiserver:
  `clusterDNS` is dropped and `authentication.webhook` is disabled. Neither
  touches the runtime stack under test. The rendered config is printed in
  `run.log` so the deviations stay visible.
