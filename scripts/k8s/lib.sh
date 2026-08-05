#!/usr/bin/env bash
# Shared helpers: version discovery, downloads, template rendering.
# Sourced by check_flags.sh and smoke_test.sh.

# Path arithmetic rather than `git rev-parse --show-toplevel`, unlike the other
# scripts here: smoke_test.sh runs under sudo, and git refuses to operate on a
# repository owned by another user ("dubious ownership") without safe.directory.
# The trade-off is that this breaks if the file moves to a different depth.
REPO_ROOT=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd)
GROUP_VARS="$REPO_ROOT/ansible/group_vars/all.yml"
TEMPLATES="$REPO_ROOT/ansible/templates"
WORK="$REPO_ROOT/scripts/k8s/tmp"
BIN="$WORK/bin"
CNI_DIR="$WORK/cni"

say() {
  printf '\n=== %s ===\n' "$*"
  return 0
}

# Never returns -- callers rely on `... || die "msg"` aborting the script. Sonar's
# S7682 does not treat `exit` as terminal and asks for a return here; adding one
# after `exit 1` would be unreachable, and replacing the exit would let callers
# carry on past a fatal error.
die() { # NOSONAR
  printf 'ERROR: %s\n' "$*" >&2
  exit 1
}

# Reads a scalar from group_vars/all.yml. No YAML parser needed -- every value we
# want is a quoted scalar on its own line.
group_var() {
  local key="$1"
  sed -nE "s/^$key:[[:space:]]*\"?([^\"#]+)\"?.*/\1/p" "$GROUP_VARS" |
    head -1 | tr -d '[:space:]'
  return 0
}

load_versions() {
  [[ -f "$GROUP_VARS" ]] || die "not found: $GROUP_VARS"
  K8S_VERSION=$(group_var k8s_version)
  CONTAINERD_VERSION=$(group_var containerd_version)
  RUNC_VERSION=$(group_var runc_version)
  CNI_VERSION=$(group_var cni_version)
  POD_CIDR=$(group_var pod_network_cidr)
  if [[ "$(uname -m)" == "aarch64" ]]; then ARCH=arm64; else ARCH=amd64; fi
  local v
  for v in K8S_VERSION CONTAINERD_VERSION RUNC_VERSION CNI_VERSION POD_CIDR; do
    [[ -n "${!v}" ]] || die "could not read $v from $GROUP_VARS"
  done
  return 0
}

print_versions() {
  printf 'k8s        %s\ncontainerd %s\nrunc       %s\ncni        %s\npod cidr   %s\narch       %s\n' \
    "$K8S_VERSION" "$CONTAINERD_VERSION" "$RUNC_VERSION" "$CNI_VERSION" "$POD_CIDR" "$ARCH"
  return 0
}

# fetch <url> <dest> -- skips if dest already exists, so reruns are cheap.
fetch() {
  local url="$1" dest="$2"
  [[ -s "$dest" ]] && return 0
  printf 'downloading %s\n' "$(basename "$dest")"
  curl -sfL -o "$dest.part" "$url" || die "download failed: $url"
  mv "$dest.part" "$dest"
  return 0
}

download_k8s_binaries() {
  mkdir -p "$BIN"
  local b
  for b in kube-apiserver kube-controller-manager kube-scheduler kubelet; do
    fetch "https://dl.k8s.io/v$K8S_VERSION/bin/linux/$ARCH/$b" "$BIN/$b"
  done
  chmod +x "$BIN"/*
  return 0
}

download_runtime() {
  mkdir -p "$BIN" "$CNI_DIR"
  local ctr_tgz="$WORK/containerd.tgz" cni_tgz="$WORK/cni.tgz" crictl_tgz="$WORK/crictl.tgz"
  local crictl_version="${K8S_VERSION%.*}.0"
  fetch "https://github.com/containerd/containerd/releases/download/v$CONTAINERD_VERSION/containerd-$CONTAINERD_VERSION-linux-$ARCH.tar.gz" "$ctr_tgz"
  fetch "https://github.com/opencontainers/runc/releases/download/v$RUNC_VERSION/runc.$ARCH" "$BIN/runc"
  fetch "https://github.com/containernetworking/plugins/releases/download/v$CNI_VERSION/cni-plugins-linux-$ARCH-v$CNI_VERSION.tgz" "$cni_tgz"
  fetch "https://github.com/kubernetes-sigs/cri-tools/releases/download/v$crictl_version/crictl-v$crictl_version-linux-$ARCH.tar.gz" "$crictl_tgz"
  tar xzf "$ctr_tgz" --strip-components=1 -C "$BIN"
  tar xzf "$cni_tgz" -C "$CNI_DIR"
  tar xzf "$crictl_tgz" -C "$BIN"
  chmod +x "$BIN"/*
  return 0
}

# The playbook writes its kubelet config inline rather than from a template, so
# pull that heredoc out of devcontainer-run.yml to test what actually ships.
render_inline_kubelet_config() {
  local dest="$1"
  python3 - "$REPO_ROOT/ansible/devcontainer-run.yml" "$dest" <<'PY'
import re, sys
src = open(sys.argv[1]).read()
m = re.search(r'dest: "\{\{ kubelet_data_path \}\}/config\.yaml".*?content: \|\n(.*?)\n\n', src, re.S)
if not m:
    sys.exit("could not find the inline kubelet config in devcontainer-run.yml")
body = [l[14:] if len(l) > 14 else l.strip() for l in m.group(1).split('\n')]
out = '\n'.join(body)
out = out.replace('{{ pki_path }}', '/etc/kubernetes/pki')
out = re.sub(r'\{\{ clusterDNS ?\}\}', '10.216.224.53', out)
open(sys.argv[2], 'w').write(out + '\n')
PY
  return 0
}

render_configs() {
  mkdir -p "$WORK/etc"
  cp "$TEMPLATES/containerd-config.toml.j2" "$WORK/etc/containerd-config.toml"
  sed "s|{{ pod_network_cidr }}|$POD_CIDR|" "$TEMPLATES/10-bridge.conf.j2" > "$WORK/etc/10-bridge.conf"
  render_inline_kubelet_config "$WORK/etc/kubelet-config.yaml"
  return 0
}
