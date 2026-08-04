#!/usr/bin/env bash
# Static-pod smoke test for the containerd / runc / cni / kubelet versions pinned
# in ansible/group_vars/all.yml, using this repo's real containerd, CNI and kubelet
# configs.
#
#   sudo bash scripts/k8s/smoke_test.sh
#
# Rootful podman is required: rootless cannot create cgroups under /sys/fs/cgroup,
# and kubelet additionally dies on `open /dev/kmsg` inside a user namespace.
set -uo pipefail

# shellcheck source-path=SCRIPTDIR
# shellcheck source=lib.sh
. "$(dirname -- "${BASH_SOURCE[0]}")/lib.sh"

BASE_IMAGE=${BASE_IMAGE:-docker.io/library/debian:trixie-slim}
HERE=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
OUT="$WORK/out"

[ "$(id -u)" -eq 0 ] || die "must run as root: sudo bash $0"

load_versions
say "versions under test"
print_versions

say "preparing $WORK"
mkdir -p "$WORK"
download_k8s_binaries
download_runtime
render_configs
cp "$HERE/in_container.sh" "$WORK/in_container.sh"
mkdir -p "$WORK/manifests"
cp "$HERE"/manifests/*.yaml "$WORK/manifests/"
chmod +x "$BIN"/* "$WORK/in_container.sh"

rm -rf "$OUT"; mkdir -p "$OUT"; chmod 0777 "$OUT"

# A containers.conf with `userns = "keep-id"` would deny the container real root,
# and kubelet then fails on /dev/kmsg. Neutralise whatever root would inherit.
printf '[containers]\n' > "$WORK/containers.conf"
export CONTAINERS_CONF="$WORK/containers.conf"

say "running (no -t: a TTY would inject escape codes into run.log)"
podman run --rm \
  --userns=host \
  --privileged \
  --cgroupns=host \
  --security-opt seccomp=unconfined \
  --security-opt label=disable \
  --cap-add=SYS_ADMIN --cap-add=NET_ADMIN \
  --tmpfs /run \
  --tmpfs /var/lib/kubelet \
  -v "$WORK":/test:ro \
  -v "$OUT":/out \
  -v "$BIN":/usr/local/bin-test:ro \
  -v "$CNI_DIR":/opt/cni/bin:ro \
  -e PATH=/usr/local/bin-test:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin \
  "$BASE_IMAGE" \
  bash /test/in_container.sh 2>&1 | tee "$OUT/run.log"

RC=${PIPESTATUS[0]}
echo "$RC" > "$OUT/exit-code"
chmod -R a+rX "$OUT" 2>/dev/null

echo
echo "exit: $RC"
echo "results: $OUT (run.log, exit-code, logs/)"
exit "$RC"
