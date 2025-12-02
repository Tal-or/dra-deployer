#!/usr/bin/env bash

set -euo pipefail

NAMESPACE="${NAMESPACE:-dra-deployer}"
LABEL="${LABEL:-app.kubernetes.io/name=dra-driver-memory}"
TIMEOUT="${TIMEOUT:-5m}"

echo -e "Waiting for DaemonSet pods to be ready...\nNamespace: ${NAMESPACE}\nLabel: ${LABEL}\nTimeout: ${TIMEOUT}"

if kubectl wait --for=condition=ready pod -l "${LABEL}" -n "${NAMESPACE}" --timeout="${TIMEOUT}"; then
    echo -e "\nDaemonSet pods are ready!"
    kubectl get pods -n "${NAMESPACE}" -l "${LABEL}"
    exit 0
else
    echo -e "\nDaemonSet pods failed to become ready within ${TIMEOUT}\n\nPod status:"
    kubectl get pods -n "${NAMESPACE}" -l "${LABEL}" || true
    echo -e "\nPod descriptions:"
    kubectl describe pods -n "${NAMESPACE}" -l "${LABEL}" || true
    exit 1
fi

