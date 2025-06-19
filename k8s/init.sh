#!/usr/bin/env bash

set -e

VM_NAME="dev"

echo "Creating machine '${VM_NAME}'..."
podman machine init "${VM_NAME}" \
  --volume ./:/srv/k8s \
  --timezone local \
  --memory 4096 \
  --cpus 2

# Start the machine
podman machine start "${VM_NAME}"

echo -n "--- Waiting for SSH on '${VM_NAME}'..."
# This loop tries to connect until it succeeds.
while ! podman machine ssh "${VM_NAME}" 'true' &>/dev/null; do
    printf "."
    sleep 2
done
echo -e "\n--- Machine is ready. ---"

echo "--- Running Ansible playbook 'init.yml' inside the machine... ---"
podman machine ssh "${VM_NAME}" \
  'export ANSIBLE_CONFIG=/srv/k8s/ansible.cfg;
  ansible-galaxy collection install -r /srv/k8s/requirements.yml;
  cd /srv/k8s/ &&
  ansible-playbook /srv/k8s/init.yml'

echo "--- Playbook execution finished. ---"
echo "--- For a full log, you can run the playbook command manually with -vvv. ---"
echo "--- Sending reboot command to apply rpm-ostree changes... ---"
podman machine ssh "${VM_NAME}" 'sudo reboot'
