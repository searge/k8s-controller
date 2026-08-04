#!/usr/bin/env bash
# Reports every flag the playbooks pass that the target-version binary no longer
# accepts, then validates the kubelet and containerd configs against the real
# parsers. No containers, no privileges.
set -uo pipefail

# shellcheck source-path=SCRIPTDIR
# shellcheck source=lib.sh
. "$(dirname -- "${BASH_SOURCE[0]}")/lib.sh"

load_versions
say "versions under test"
print_versions

say "downloading binaries"
mkdir -p "$WORK"
download_k8s_binaries
download_runtime

say "flags: playbooks vs binary --help"
python3 - "$REPO_ROOT" "$BIN" <<'PY'
import re, subprocess, sys, glob, os
repo, bindir = sys.argv[1], sys.argv[2]
BINS = ['kube-apiserver', 'kube-controller-manager', 'kube-scheduler', 'kubelet']
helps = {b: subprocess.run([os.path.join(bindir, b), '--help'],
                           capture_output=True, text=True).stdout for b in BINS}
bad = 0

def report(where, binary, flags):
    global bad
    missing = sorted(f for f in flags if '--' + f not in helps[binary])
    status = 'REMOVED: ' + ', '.join(missing) if missing else 'ok'
    if missing:
        bad += 1
    print(f'{where:42} {binary:24} {len(flags):3} flags  {status}')

# The devcontainer path starts components inline with nohup.
run_yml = os.path.join(repo, 'ansible/devcontainer-run.yml')
src = open(run_yml).read()
for b in BINS:
    m = re.search(r'nohup \{\{ bin_path \}\}/' + re.escape(b) + r'(.*?)(?=\n\s*(?:>|echo|\Z))', src, re.S)
    if m:
        report('devcontainer-run.yml', b, set(re.findall(r'--([a-z0-9-]+)=', m.group(1))))

# The VM path starts them from systemd units.
for path in sorted(glob.glob(os.path.join(repo, 'ansible/templates/*.service.j2'))):
    name = os.path.basename(path)
    binary = name[:-len('.service.j2')]
    if binary in BINS:
        report(name, binary, set(re.findall(r'--([a-z0-9-]+)=', open(path).read())))

sys.exit(1 if bad else 0)
PY
FLAGS_RC=$?

say "kubelet config files vs kubelet $K8S_VERSION"
# Reaching the /var/lib/kubelet mkdir means load + validation both succeeded.
render_configs
CONF_RC=0
for f in "$WORK/etc/kubelet-config.yaml" \
         "$TEMPLATES/kubelet-config.yaml.j2" "$TEMPLATES/kubelet-config-devc.yaml.j2"; do
  rendered="$WORK/etc/probe-$(basename "$f" .j2)"
  sed -e 's|{{ pki_path }}|/etc/kubernetes/pki|g' -e 's|{{ clusterDNS ?}}|10.216.224.53|g' \
      -e 's|{{ clusterDNS}}|10.216.224.53|g' -e 's|{{ clusterDNS }}|10.216.224.53|g' \
      -e '/^{#/d' "$f" > "$rendered"
  err=$("$BIN/kubelet" --config="$rendered" --kubeconfig=/nonexistent 2>&1 | head -1)
  if printf '%s' "$err" | grep -q 'mkdir /var/lib/kubelet'; then
    printf '%-40s ok\n' "$(basename "$f")"
  else
    printf '%-40s REJECTED: %s\n' "$(basename "$f")" "${err#*] }"
    CONF_RC=1
  fi
done

say "containerd config vs containerd $CONTAINERD_VERSION"
# Unknown or renamed keys are reported on stderr as warnings, not as a failure,
# so treat any stderr output as a finding.
CTR_ERR=$("$BIN/containerd" -c "$WORK/etc/containerd-config.toml" config dump 2>&1 >/dev/null)
if [ -z "$CTR_ERR" ]; then
  echo "containerd-config.toml.j2               ok"
  CTR_RC=0
else
  echo "containerd-config.toml.j2               FINDINGS:"
  printf '%s\n' "$CTR_ERR"
  CTR_RC=1
fi

say "result"
RC=0
[ "$FLAGS_RC" -ne 0 ] && { echo "FAIL: removed flags found"; RC=1; }
[ "$CONF_RC" -ne 0 ] && { echo "FAIL: a kubelet config was rejected"; RC=1; }
[ "$CTR_RC" -ne 0 ] && { echo "FAIL: containerd config findings"; RC=1; }
[ "$RC" -eq 0 ] && echo "PASS: flags and configs are clean for k8s $K8S_VERSION"
exit "$RC"
