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
