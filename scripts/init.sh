#!/usr/bin/env bash

# Stop the script if any command fails.
set -e

# Get the absolute path to the directory where the script is located.
SCRIPT_DIR=$(cd -- "$(dirname -- "$0")" &> /dev/null && pwd)
# The project root is one level above the script's directory.
PROJECT_ROOT=$(dirname "$SCRIPT_DIR")
# Change the current working directory to the project root.
cd "$PROJECT_ROOT"

echo "--- Running script from project root: $(pwd) ---"

# --- Script Variables ---
VM_NAME="dev"
APP_DIR="/srv/app"


echo "Creating machine '${VM_NAME}'..."
set -x
podman machine init "${VM_NAME}" \
  --volume ./:${APP_DIR} \
  --timezone local \
  --memory 4096 \
  --cpus 2
set +x

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
  "export ANSIBLE_CONFIG=${APP_DIR}/ansible/ansible.cfg;
  cd ${APP_DIR}/ansible/ &&
  ansible-galaxy collection install -r requirements.yml;
  ansible-playbook init.yml"

echo "--- Playbook execution finished. ---"
echo "--- For a full log, please check /var/tmp/ansible.log ---"

echo
echo "======================================================================================"
echo "IMPORTANT: System packages were installed and a reboot is required to apply them."
echo
# The 'stop' command is blocking and will wait until the machine is fully shut down.
podman machine stop "${VM_NAME}"

echo "--- Restarting machine '${VM_NAME}'... ---"
podman machine start "${VM_NAME}" && sleep 2

echo -n "--- Waiting for machine to become available after reboot..."
# It's crucial to wait for SSH to be ready again after the restart.
while ! podman machine ssh "${VM_NAME}" 'true' &>/dev/null; do
    printf "."
    sleep 2
done

echo -e "\n--- Machine has been restarted successfully and is ready for use! ---"
echo "======================================================================================"
