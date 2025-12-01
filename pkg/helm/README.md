# Helm Chart Loader Package

This package provides functionality to load and render the DRA Memory Driver Helm chart from the filesystem.

## Overview

The Helm chart loader allows you to:

- Load the Helm chart from the filesystem
- Render templates with default or custom values
- Convert rendered templates into Kubernetes runtime objects

## Usage

### Basic Usage with Default Values

```go
package main

import (
    "github.com/Tal-or/dra-deployer/pkg/helm"
)

func main() {
    // Create a new chart loader (empty string uses default path)
    loader, err := helm.NewChartLoader("")
    if err != nil {
        panic(err)
    }

    // Or specify a custom chart path
    // loader, err := helm.NewChartLoader("/path/to/chart")

    // Render with default values from values.yaml
    opts := helm.Options{
        ReleaseName: "dra-driver-memory",
        Namespace:   "dra-system",
        Values:      nil, // Use defaults
    }

    objects, err := loader.Render(opts)
    if err != nil {
        panic(err)
    }

    // objects is []*unstructured.Unstructured ready to apply
    for _, obj := range objects {
        // Apply to cluster using client.Create(ctx, obj)
    }
}
```

### Advanced Usage with Custom Values

```go
// Override specific values at runtime
opts := helm.Options{
    ReleaseName: "dra-driver-memory",
    Namespace:   "my-namespace",
    Values: map[string]interface{}{
        "image": map[string]interface{}{
            "repository": "quay.io/myorg/dra-memory-driver",
            "tag":        "v0.2.0",
            "pullPolicy": "IfNotPresent",
        },
        "openshift": map[string]interface{}{
            "enabled": true, // Enable OpenShift SCC
        },
        "daemonset": map[string]interface{}{
            "env": map[string]interface{}{
                "numDevices": "16",
            },
        },
    },
}

objects, err := loader.Render(opts)
```

## How It Works

1. **Chart Loading**: The Helm chart is loaded from the filesystem at `assets/deployment/helm/dra-driver-memory/` by default.

2. **Chart Rendering**: The loader uses the official Helm SDK (`helm.sh/helm/v3`) to:
   - Load chart files from the filesystem
   - Parse the `values.yaml` file for defaults
   - Merge any provided custom values with defaults (using `chartutil.CoalesceTables`)
   - Render all templates using the Helm template engine

3. **Object Conversion**: Rendered YAML templates are parsed into `*unstructured.Unstructured` objects, making them ready to apply to a cluster.

## API Reference

### `NewChartLoader(chartPath string) (*ChartLoader, error)`

Creates a new ChartLoader by loading the Helm chart from the filesystem.

**Parameters:**

- `chartPath` (string): Path to the Helm chart directory. If empty, uses `assets/deployment/helm/dra-driver-memory`

### `ChartLoader.Render(opts Options) ([]*unstructured.Unstructured, error)`

Renders the Helm chart with the given options and returns Kubernetes objects.

**Options:**

- `ReleaseName` (string): The Helm release name
- `Namespace` (string): The target Kubernetes namespace
- `Values` (map[string]interface{}): Optional custom values to override defaults

**Returns:**

- A slice of `[]*unstructured.Unstructured` containing all rendered Kubernetes resources
- An error if rendering fails

### `ChartLoader.GetChart() *chart.Chart`

Returns the loaded Helm chart metadata.

## Default Values

All default values are defined in `assets/deployment/helm/dra-driver-memory/values.yaml`.

Key configurable values:

- `image.repository`: Container image repository
- `image.tag`: Container image tag
- `image.pullPolicy`: Image pull policy
- `openshift.enabled`: Enable OpenShift-specific resources (SCC)
- `daemonset.env.numDevices`: Number of memory devices per node
- `rbac.create`: Create RBAC resources
- `validatingAdmissionPolicy.create`: Create ValidatingAdmissionPolicy

## Testing

Run the tests:

```bash
cd pkg/helm
go test -v
```

## Notes

- The chart is loaded from the filesystem, so changes to the Helm chart take effect immediately without rebuilding
- Values in `values.yaml` serve as defaults; they can be overridden programmatically via the `Values` field in `Options`
- The loader automatically handles multi-document YAML files (templates with `---` separators)
- Empty or non-YAML files (like `_helpers.tpl`) are automatically skipped during parsing
- The default chart path is relative to the project root: `assets/deployment/helm/dra-driver-memory`
