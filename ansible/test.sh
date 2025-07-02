#!/bin/bash

# Exit on error
set -e

echo "Setting up Kubernetes control plane for devcontainer..."

# Function to check if a process is running
is_running() {
    pgrep -f "$1" >/dev/null
}

# Function to check if all components are running
check_running() {
    is_running "etcd" && \
    is_running "kube-apiserver" && \
    is_running "kube-controller-manager" && \
    is_running "kube-scheduler" && \
    is_running "kubelet" && \
    is_running "containerd"
}

# Function to kill process if running
stop_process() {
    if is_running "$1"; then
        echo "Stopping $1..."
        sudo pkill -f "$1" || true
        while is_running "$1"; do
            sleep 1
        done
    fi
}

start() {
    if check_running; then
        echo "Kubernetes components are already running"
        return 0
    fi

    HOST_IP=$(hostname -I | awk '{print $1}')

    # Start components if not running
    if ! is_running "etcd"; then
        echo "Starting etcd..."
        sudo /usr/local/bin/etcd \
            --advertise-client-urls http://$HOST_IP:2379 \
            --listen-client-urls http://0.0.0.0:2379 \
            --data-dir /var/lib/etcd \
            --listen-peer-urls http://0.0.0.0:2380 \
            --initial-cluster default=http://$HOST_IP:2380 \
            --initial-advertise-peer-urls http://$HOST_IP:2380 \
            --initial-cluster-state new \
            --initial-cluster-token test-token &
    fi

    if ! is_running "kube-apiserver"; then
        echo "Starting kube-apiserver..."
        echo "use application/vnd.kubernetes.protobuf for better performance"
        sudo /usr/local/bin/kube-apiserver \
            --etcd-servers=http://$HOST_IP:2379 \
            --service-cluster-ip-range=10.96.0.0/12 \
            --bind-address=0.0.0.0 \
            --secure-port=6443 \
            --advertise-address=$HOST_IP \
            --authorization-mode=AlwaysAllow \
            --token-auth-file=/etc/kubernetes/token.csv \
            --enable-priority-and-fairness=false \
            --allow-privileged=true \
            --profiling=false \
            --storage-backend=etcd3 \
            --storage-media-type=application/json \
            --v=0 \
            --service-account-issuer=https://kubernetes.default.svc.cluster.local \
            --service-account-key-file=/etc/kubernetes/pki/sa.pub \
            --service-account-signing-key-file=/etc/kubernetes/pki/sa.key \
            --client-ca-file=/etc/kubernetes/pki/ca.crt \
            --tls-cert-file=/etc/kubernetes/pki/apiserver.crt \
            --tls-private-key-file=/etc/kubernetes/pki/apiserver.key \
            --kubelet-client-certificate=/etc/kubernetes/pki/apiserver.crt \
            --kubelet-client-key=/etc/kubernetes/pki/apiserver.key &
    fi

    if ! is_running "containerd"; then
        echo "Starting containerd..."
        export PATH=$PATH:/opt/cni/bin:/usr/local/bin
        sudo PATH=$PATH:/opt/cni/bin:/usr/sbin /usr/local/bin/containerd -c /etc/containerd/config.toml &
    fi

    if ! is_running "kube-scheduler"; then
        echo "Starting kube-scheduler..."
        sudo /usr/local/bin/kube-scheduler \
            --kubeconfig=/etc/kubernetes/kube-scheduler.kubeconfig \
            --leader-elect=false \
            --v=2 \
            --bind-address=0.0.0.0 &
    fi

    # Set up kubelet kubeconfig
    sudo cp /root/.kube/config /var/lib/kubelet/kubeconfig
    export KUBECONFIG=~/.kube/config
    sudo cp /etc/kubernetes/pki/sa.pub /etc/kubernetes/pki/ca.crt /var/lib/kubelet/

    # Create service account and configmap if they don't exist
    sudo /usr/local/bin/kubectl create sa default 2>/dev/null || true
    sudo /usr/local/bin/kubectl create configmap kube-root-ca.crt --from-file=ca.crt=/etc/kubernetes/pki/ca.crt -n default 2>/dev/null || true

    if ! is_running "kubelet"; then
        echo "Starting kubelet..."
        sudo PATH=$PATH:/opt/cni/bin:/usr/sbin /usr/local/bin/kubelet \
            --kubeconfig=/var/lib/kubelet/kubeconfig \
            --root-dir=/var/lib/kubelet \
            --cert-dir=/var/lib/kubelet/pki \
            --tls-cert-file=/etc/kubernetes/pki/kubelet.crt \
            --tls-private-key-file=/etc/kubernetes/pki/kubelet.key \
            --hostname-override=$(hostname) \
            --pod-infra-container-image=registry.k8s.io/pause:3.9 \
            --node-ip=$HOST_IP \
            --container-runtime-endpoint=unix:///var/run/containerd/containerd.sock \
            --cgroup-driver=systemd \
            --max-pods=20 \
            --fail-swap-on=false \
            --protect-kernel-defaults=false \
            --make-iptables-util-chains=false \
            --feature-gates=KubeletInUserNamespace=true \
            --v=2 &
    fi

    # Label the node so static pods with nodeSelector can be scheduled
    NODE_NAME=$(hostname)
    sudo /usr/local/bin/kubectl label node "$NODE_NAME" node-role.kubernetes.io/master="" --overwrite || true

    if ! is_running "kube-controller-manager"; then
        echo "Starting kube-controller-manager..."
        sudo PATH=$PATH:/opt/cni/bin:/usr/sbin /usr/local/bin/kube-controller-manager \
            --kubeconfig=/etc/kubernetes/kube-controller-manager.kubeconfig \
            --leader-elect=false \
            --service-cluster-ip-range=10.96.0.0/12 \
            --cluster-name=kubernetes \
            --cluster-cidr=10.244.0.0/16 \
            --root-ca-file=/etc/kubernetes/pki/ca.crt \
            --service-account-private-key-file=/etc/kubernetes/pki/sa.key \
            --use-service-account-credentials=true \
            --v=2 &
    fi

    echo "Waiting for components to be ready..."
    sleep 10

    echo "Verifying setup..."
    sudo /usr/local/bin/kubectl get nodes
    sudo /usr/local/bin/kubectl get all -A
    sudo /usr/local/bin/kubectl get componentstatuses || true
    sudo /usr/local/bin/kubectl get --raw='/readyz?verbose'
}

stop() {
    echo "Stopping Kubernetes components..."
    stop_process "kube-controller-manager"
    stop_process "kubelet"
    stop_process "kube-scheduler"
    stop_process "kube-apiserver"
    stop_process "containerd"
    stop_process "etcd"
    echo "All components stopped"
}

cleanup() {
    stop
    echo "Cleaning up..."
    sudo rm -rf /var/lib/etcd/*
    sudo rm -rf /var/lib/kubelet/pods/*
    sudo rm -rf /run/containerd/*
    echo "Cleanup complete"
}

case "${1:-}" in
    start)
        start
        ;;
    stop)
        stop
        ;;
    cleanup)
        cleanup
        ;;
    *)
        echo "Usage: $0 {start|stop|cleanup}"
        exit 1
        ;;
esac
