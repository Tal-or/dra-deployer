package commands

import (
	"context"

	"github.com/spf13/cobra"

	cli "github.com/Tal-or/dra-deployer/pkg/client"
	"github.com/Tal-or/dra-deployer/pkg/deploy"
)

func NewApplyCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "apply",
		Short: "Apply DRA plugin manifests to a Kubernetes cluster",
		Long: `Apply all DRA plugin manifests directly to a Kubernetes cluster. This will 
		create or update the necessary resources including ServiceAccount, ClusterRole, 
		ClusterRoleBinding, DaemonSet, DeviceClasses, and ValidatingAdmissionPolicy.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return apply()
		},
	}
}

func NewDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete",
		Short: "Delete DRA plugin manifests from a Kubernetes cluster",
		Long: `Delete all DRA plugin manifests from a Kubernetes cluster. If a namespace 
		is specified, deleting the namespace will automatically remove all namespaced resources 
		(ServiceAccount, DaemonSet). Cluster-scoped resources will be deleted explicitly.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return delete()
		},
	}
}

// deploy applies all manifests to the Kubernetes cluster
func apply() error {
	// Create client with all necessary types registered
	c, err := cli.New()
	if err != nil {
		return err
	}

	return deploy.Deploy(context.Background(), c, namespace, image)
}

// delete removes all manifests from the Kubernetes cluster
func delete() error {
	// Create client with all necessary types registered
	c, err := cli.New()
	if err != nil {
		return err
	}

	return deploy.Delete(context.Background(), c, namespace)
}
