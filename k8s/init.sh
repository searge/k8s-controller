#!/usr/bin/env bash

# Inialize Podman machine
podman machine init dev \
  -v ./:/srv/k8s \
  --playbook init.yml \
  --rootful \
  --timezone local
