package deploy

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/Tal-or/dra-deployer/pkg/helm"
	"github.com/Tal-or/dra-deployer/pkg/params"
)

type Options struct {
	Namespace string
	Image     string
	Command   string
}

func Deploy(ctx context.Context, cli client.Client, envConfig params.EnvConfig) error {
	klog.InfoS("deploying manifests to cluster", "namespace", envConfig.Namespace)

	// Check and create namespace if needed
	err := createNamespaceIfNeeded(ctx, cli, envConfig.Namespace)
	if err != nil {
		return fmt.Errorf("failed to create namespace: %w", err)
	}

	// Load Helm chart
	chartLoader, err := helm.NewChartLoader("")
	if err != nil {
		return fmt.Errorf("failed to load Helm chart: %w", err)
	}

	objects, err := chartLoader.Render(envConfig)

	if err != nil {
		return fmt.Errorf("failed to render Helm chart: %w", err)
	}

	// Deploy all objects
	for _, obj := range objects {
		key := fmt.Sprintf("%s/%s/%s", obj.GetObjectKind().GroupVersionKind().Kind, obj.GetNamespace(), obj.GetName())
		klog.V(4).InfoS("creating/updating", "key", key)
		result, err := controllerutil.CreateOrUpdate(ctx, cli, obj, nil)
		if err != nil {
			return fmt.Errorf("failed to create/update object %s: %w", key, err)
		}
		klog.V(4).InfoS("created/updated", "key", key, "result", result)
	}

	klog.InfoS("Successfully deployed core manifests to cluster")
	return nil
}

func createNamespaceIfNeeded(ctx context.Context, cli client.Client, namespace string) error {
	// Check and create namespace if needed
	ns := &corev1.Namespace{}
	err := cli.Get(ctx, client.ObjectKey{Name: namespace}, ns)
	if err != nil {
		if errors.IsNotFound(err) {
			// Create namespace
			ns.Name = namespace
			ns.Labels = map[string]string{
				"app.kubernetes.io/managed-by": "dra-deployer",
			}
			err = cli.Create(ctx, ns)
			if err != nil {
				return fmt.Errorf("failed to create namespace: %w", err)
			}
			klog.InfoS("Created namespace", "namespace", namespace)
		} else {
			return fmt.Errorf("failed to get namespace: %w", err)
		}
	} else {
		klog.V(4).InfoS("Namespace already exists", "namespace", namespace)
	}
	return nil
}

// Delete removes all DRA plugin manifests from the cluster
func Delete(ctx context.Context, cli client.Client, namespace string) error {
	klog.InfoS("Deleting manifests from cluster", "namespace", namespace)

	// Load Helm chart
	chartLoader, err := helm.NewChartLoader("")
	if err != nil {
		return fmt.Errorf("failed to load Helm chart: %w", err)
	}

	objects, err := chartLoader.Render(params.EnvConfig{
		Namespace: namespace,
		Values:    nil,
	})
	if err != nil {
		return fmt.Errorf("failed to render Helm chart: %w", err)
	}

	// Delete namespace (this will cascade delete namespaced resources like ServiceAccount and DaemonSet)
	ns := &corev1.Namespace{}
	ns.Name = namespace

	klog.V(2).InfoS("Deleting namespace", "namespace", namespace)
	err = cli.Delete(ctx, ns)
	if err != nil {
		if errors.IsNotFound(err) {
			klog.V(4).InfoS("Namespace already deleted", "namespace", namespace)
		} else {
			return fmt.Errorf("failed to delete namespace: %w", err)
		}
	} else {
		klog.InfoS("Deleted namespace", "namespace", namespace)
	}

	// Delete objects
	for _, obj := range objects {
		// Skip namespaced objects
		if obj.GetNamespace() != "" {
			continue
		}

		// Delete cluster-scoped objects
		key := fmt.Sprintf("%s/%s", obj.GetObjectKind().GroupVersionKind().Kind, obj.GetName())
		klog.V(2).InfoS("Deleting object", "key", key)

		err := cli.Delete(ctx, obj)
		if err != nil {
			if errors.IsNotFound(err) {
				klog.V(4).InfoS("Resource already deleted", "key", key)
			} else {
				return fmt.Errorf("failed to delete object %s: %w", key, err)
			}
		} else {
			klog.InfoS("Deleted resource", "key", key)
		}
	}

	klog.InfoS("Successfully deleted all manifests from cluster")
	return nil
}
