#!/usr/bin/env bash

set -euo pipefail

NAMESPACE="${NAMESPACE:-dra-deployer}"

echo "=== Cluster Information ==="
kubectl cluster-info || true

echo -e "\n=== Nodes ==="
kubectl get nodes -o wide || true

echo -e "\n=== All Pods ==="
kubectl get pods -A || true

echo -e "\n=== DRA Deployer Namespace ==="
kubectl get all -n "${NAMESPACE}" || true

echo -e "\n=== DRA Deployer Pods Description ==="
kubectl describe pods -n "${NAMESPACE}" || true

echo -e "\n=== DRA Deployer Pods Logs ==="
kubectl logs -n "${NAMESPACE}" -l app.kubernetes.io/name=dra-driver-memory --tail=100 || true

echo -e "\n=== Cluster Resources ==="
kubectl get clusterrole,clusterrolebinding,validatingadmissionpolicy | grep dra-driver-memory || true

