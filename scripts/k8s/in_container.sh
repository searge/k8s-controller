#!/usr/bin/env bash
# Runs INSIDE the privileged container started by smoke_test.sh. Brings up
# containerd and kubelet in standalone mode (no apiserver) and checks whether the
# static pods reach Running.
set -u

say() { printf '\n=== %s ===\n' "$*"; }
LOG=/out/logs   # bind-mounted from the host, survives the --rm container
mkdir -p "$LOG" /etc/containerd /etc/cni/net.d /etc/kubernetes/manifests /var/lib/kubelet

say "versions"
containerd --version
runc --version | head -1
kubelet --version
crictl --version
/opt/cni/bin/bridge 2>&1 | head -1 || true

say "prerequisites"
apt-get update -qq || { echo "FAIL: apt-get update"; exit 1; }
# ca-certificates: registry TLS. openssl: throwaway CA for the kubelet config.
apt-get install -y -qq --no-install-recommends iptables ca-certificates openssl \
  || { echo "FAIL: apt-get install"; exit 1; }
iptables --version || { echo "FAIL: no iptables"; exit 1; }

say "throwaway CA for the kubelet x509 clientCAFile"
mkdir -p /etc/kubernetes/pki
openssl req -x509 -newkey rsa:2048 -nodes -days 1 -subj "/CN=smoke-ca" \
  -keyout /etc/kubernetes/pki/ca.key -out /etc/kubernetes/pki/ca.crt 2>/dev/null \
  || { echo "FAIL: could not generate CA"; exit 1; }

say "config"
cp /test/etc/containerd-config.toml /etc/containerd/config.toml
cp /test/etc/10-bridge.conf /etc/cni/net.d/10-bridge.conf
cp /test/etc/kubelet-config.yaml /var/lib/kubelet/config.yaml
# Two forced deviations, both because there is no apiserver in this test. Neither
# touches the runtime stack under test.
#   1. clusterDNS -- nothing to resolve against
#   2. authentication.webhook -- standalone kubelet has no client to call SAR with
sed -i '/clusterDNS/,+1d' /var/lib/kubelet/config.yaml
sed -i '/^  webhook:/{n;s/enabled: true/enabled: false/}' /var/lib/kubelet/config.yaml
echo "--- kubelet config as tested:"
cat /var/lib/kubelet/config.yaml

say "start containerd"
nohup containerd -c /etc/containerd/config.toml > "$LOG/containerd.log" 2>&1 &
for _ in $(seq 1 30); do
  [ -S /run/containerd/containerd.sock ] && break
  sleep 1
done
[ -S /run/containerd/containerd.sock ] || {
  echo "FAIL: containerd socket never appeared"; tail -20 "$LOG/containerd.log"; exit 1; }
ctr version | tail -2

say "pre-pull images"
ctr -n k8s.io images pull "--platform=linux/$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')" \
  registry.k8s.io/pause:3.10 2>&1 | tail -1
ctr -n k8s.io images pull "--platform=linux/$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')" \
  docker.io/library/busybox:latest 2>&1 | tail -1

say "CRI reachable"
export CONTAINER_RUNTIME_ENDPOINT=unix:///run/containerd/containerd.sock
printf 'runtime-endpoint: unix:///run/containerd/containerd.sock\ntimeout: 10\n' > /etc/crictl.yaml
crictl version || { echo "FAIL: CRI handshake failed"; exit 1; }

say "start kubelet (standalone, no apiserver)"
cp /test/manifests/*.yaml /etc/kubernetes/manifests/
nohup kubelet --config=/var/lib/kubelet/config.yaml --max-pods=4 --v=4 \
  > "$LOG/kubelet.log" 2>&1 &
KUBELET_PID=$!
sleep 5
kill -0 "$KUBELET_PID" 2>/dev/null || {
  echo "FAIL: kubelet exited immediately"; tail -30 "$LOG/kubelet.log"; exit 1; }

say "wait for static pods (up to 90s)"
# Count with crictl's own state filters. Grepping `-o json` for state strings is
# fragile enough to report 0 on a healthy node.
pods_ready() { crictl pods --state ready -q 2>/dev/null | grep -c . ; }
ctrs_running() { crictl ps --state running -q 2>/dev/null | grep -c . ; }
for i in $(seq 1 30); do
  ready=$(ctrs_running)
  printf 't=%3ds sandboxes_ready=%s containers_running=%s | %s\n' \
    "$((i * 3))" "$(pods_ready)" "$ready" \
    "$(grep -oE 'E[0-9]{4} .*' "$LOG/kubelet.log" 2>/dev/null | tail -1 | cut -c1-100)"
  [ "$ready" -ge 2 ] && break
  sleep 3
done

say "crictl pods"; crictl pods 2>&1
say "crictl ps -a"; crictl ps -a 2>&1

say "CNI: pod IP from the configured pod network"
POD_CIDR_PREFIX=$(sed -nE 's/.*"subnet": "([0-9]+\.[0-9]+)\..*/\1/p' /etc/cni/net.d/10-bridge.conf)
POD_IP=$(grep -oE "$POD_CIDR_PREFIX\.[0-9]+\.[0-9]+" "$LOG/kubelet.log" | sort -u | head -1)
if [ -n "$POD_IP" ]; then
  echo "ok: CNI bridge allocated $POD_IP"
else
  echo "WARNING: no address from ${POD_CIDR_PREFIX}.0.0/16 seen -- CNI may not have run"
fi

say "verdict"
ready=$(ctrs_running)
if [ "$ready" -ge 2 ]; then
  echo "PASS: both static pods running (hostNetwork + CNI)"; RC=0
elif [ "$ready" -ge 1 ]; then
  echo "PARTIAL: $ready/2 containers running"; RC=2
else
  echo "FAIL: no containers reached Running"; RC=1
fi

say "kubelet errors"
grep -oE 'E[0-9]{4} .*' "$LOG/kubelet.log" | tail -20
say "containerd errors"
grep -E 'level=(error|fatal)' "$LOG/containerd.log" | tail -15

exit "$RC"
