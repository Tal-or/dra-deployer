# dra-deployer

Command line tool for deploying Dynamic Resource Allocation (DRA) plugins to Kubernetes clusters.

## Installation

```shell
make build
```

The binary will be available at `./bin/dra-deployer`.

## Commands

### `render`

Render all DRA plugin manifests as YAML to stdout. This is useful for reviewing the manifests before applying them or for piping to kubectl.

```shell
./bin/dra-deployer render
```

### `apply`

Apply all DRA plugin manifests directly to a Kubernetes cluster. This will create or update the necessary resources including ServiceAccount, ClusterRole, ClusterRoleBinding, DaemonSet, DeviceClasses, and ValidatingAdmissionPolicy.

```shell
./bin/dra-deployer apply
```

### `delete`

Delete all DRA plugin manifests from a Kubernetes cluster. Deleting the namespace will automatically remove all namespaced resources (ServiceAccount, DaemonSet). Cluster-scoped resources will be deleted explicitly.

```shell
./bin/dra-deployer delete
```

## Global Flags

All commands support the following flags:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--namespace` | `-n` | string | `dra-deployer` | Namespace for namespaced resources |
| `--image` | `-i` | string | `quay.io/titzhak/dra-cpu-driver:latest` | Container image for the DRA plugin |
| `--verbose` | `-v` | int | `2` | Log level verbosity (0-10) |

## Usage Examples

```shell
# Apply manifests with default settings
./bin/dra-deployer apply

# Apply manifests with custom namespace and image
./bin/dra-deployer apply --namespace my-dra-namespace --image quay.io/myorg/dra-driver:v1.0.0

# Render manifests with custom settings
./bin/dra-deployer render -n my-namespace -i custom-image:latest > manifests.yaml

# Delete manifests from custom namespace
./bin/dra-deployer delete --namespace my-dra-namespace

# Increase verbosity for debugging
./bin/dra-deployer apply -v 4
```

## Development

```shell
# Build the project
make build

# Run tests
make test-unit

# Run linters
make lint
```

## Notes

For DRA drivers deployment DynamicResourceAllocation Feature must be enable under FeatureGate.
On OpenShift, add `TechPreviewNoUpgrade` to FeatureGate CR:

```yaml
apiVersion: config.openshift.io/v1
kind: FeatureGate
metadata:
  name: cluster 
....

spec:
  featureSet: TechPreviewNoUpgrade 
```
