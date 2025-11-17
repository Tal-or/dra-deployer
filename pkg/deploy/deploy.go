package deploy

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/Tal-or/dra-deployer/pkg/manifests"
)

func Deploy(ctx context.Context, cli client.Client, namespace string) error {
	klog.InfoS("deploying manifests to cluster", "namespace", namespace)

	// Check and create namespace if needed
	err := createNamespaceIfNeeded(ctx, cli, namespace)
	if err != nil {
		return fmt.Errorf("failed to create namespace: %w", err)
	}

	// Get all manifests
	m, err := manifests.GetAll(namespace)
	if err != nil {
		return fmt.Errorf("failed to get manifests: %w", err)
	}

	// Deploy all manifests
	for _, obj := range m.GetObjects() {
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

	// Get all manifests
	m, err := manifests.GetAll(namespace)
	if err != nil {
		return fmt.Errorf("failed to get manifests: %w", err)
	}

	// Delete cluster-scoped resources in reverse order
	clusterScopedObjects := []client.Object{
		m.ValidatingAdmissionPolicyBinding,
		m.ValidatingAdmissionPolicy,
		m.SecurityContextConstraints,
		m.ClusterRoleBinding,
		m.ClusterRole,
	}

	for _, obj := range clusterScopedObjects {
		if obj == nil {
			continue
		}
		key := fmt.Sprintf("%s/%s", obj.GetObjectKind().GroupVersionKind().Kind, obj.GetName())
		klog.V(2).InfoS("Deleting cluster-scoped resource", "key", key)

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

	klog.InfoS("Successfully deleted all manifests from cluster")
	return nil
}
