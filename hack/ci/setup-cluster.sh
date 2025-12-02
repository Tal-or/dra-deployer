#!/usr/bin/env bash

set -euo pipefail

echo "Setting up Minikube cluster with CRI-O and NRI enabled..."

# Create CRI-O NRI configuration file
echo "Creating CRI-O NRI configuration..."
cat > /tmp/crio-nri.conf <<EOF
[crio.nri]
enable_nri = true
EOF

echo "CRI-O NRI configuration created at /tmp/crio-nri.conf"

# Check if minikube is already installed
if ! command -v minikube &> /dev/null; then
    echo "Installing Minikube..."
    curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
    sudo install minikube-linux-amd64 /usr/local/bin/minikube
    rm minikube-linux-amd64
    echo "Minikube installed successfully"
else
    echo "Minikube is already installed"
fi

# Start Minikube with CRI-O runtime and NRI enabled
echo "Starting Minikube cluster..."
minikube start \
    --container-runtime=crio \
    --kubernetes-version=v1.34.0 \
    --feature-gates=DynamicResourceAllocation=true \
    --mount-string="/tmp/crio-nri.conf:/etc/crio/crio.conf.d/99-nri.conf" \
    --mount

echo "Waiting for cluster to be ready..."
kubectl wait --for=condition=Ready nodes --all --timeout=5m

echo "Verifying NRI configuration..."
if minikube ssh "test -f /etc/crio/crio.conf.d/99-nri.conf"; then
    echo "NRI configuration file exists in the cluster"
    minikube ssh "cat /etc/crio/crio.conf.d/99-nri.conf"
else
    echo "Warning: NRI configuration file not found"
fi

echo -e "\nCluster setup completed successfully!"
kubectl get nodes

