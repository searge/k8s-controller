# Kubernetes Devcontainer Setup

## ✅ Working Features

- **Full Kubernetes Control Plane** - etcd, kube-apiserver, kube-controller-manager, kube-scheduler
- **Kubelet with containerd** - single-node cluster ready for development
- **kubectl CLI** - all API operations work
- **Pod scheduling** - pods are assigned to the node
- **Service accounts & RBAC** - authentication and authorization work
- **Custom Resources** - CRDs can be created and managed

## ⚠️ Known Limitations

### Pod Networking

Pods cannot start due to missing `iptables` in devcontainer environment:

```bash
failed to locate iptables: exec: "iptables": executable file not found in $PATH
```

**Impact:**

- Pods remain in `ContainerCreating` state
- No network connectivity for pods
- Services cannot route traffic

**Workaround:** Use for development that doesn't require running pods (API testing, controller development, etc.)

## 🚀 Getting Started

1. **Start the cluster:**

   ```bash
   task collections  # Install Ansible collections
   ansible-playbook devcontainer.yml     # Setup components
   ansible-playbook devcontainer-run.yml # Start cluster
   ```

2. **Verify cluster:**

   ```bash
   kubectl get nodes
   kubectl get all -A
   kubectl cluster-info
   ```

3. **Check status:**

   ```bash
   ~/k8s-status.sh  # Check running processes
   ~/k8s-stop.sh    # Stop all components
   ```

## 🧪 What You Can Test

- **API operations:** `kubectl apply`, `kubectl get`, `kubectl delete`
- **Custom controllers:** Deploy and test Kubernetes operators
- **CRDs:** Create custom resource definitions
- **RBAC:** Test roles, bindings, service accounts
- **Admission controllers:** Test webhooks and policies
- **Scheduling:** Test node selectors, taints, tolerations

## 📁 Important Paths

- **Logs:** `/var/log/kubernetes/`
- **Configs:** `/etc/kubernetes/`
- **Data:** `/var/lib/etcd/`, `/var/lib/kubelet/`
- **Binaries:** `/usr/local/bin/`

## 🔧 Troubleshooting

**Cluster not responding?**

```bash
ps aux | grep -E "(etcd|kube-|containerd)"
~/k8s-status.sh
```

**Need to restart?**

```bash
~/k8s-stop.sh
ansible-playbook devcontainer-run.yml -t control-plane,worker
```

**View component logs:**

```bash
tail -f /var/log/kubernetes/kubelet.log
tail -f /var/log/kubernetes/kube-apiserver.log
```
