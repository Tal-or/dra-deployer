package commands

import (
	"context"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	cli "github.com/Tal-or/dra-deployer/pkg/client"
	"github.com/Tal-or/dra-deployer/pkg/deploy"
	"github.com/Tal-or/dra-deployer/pkg/params"
)

type applyArgs struct {
	command string
}

func NewApplyCommand(applyArgs *applyArgs) *cobra.Command {
	applyCmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply DRA plugin manifests to a Kubernetes cluster",
		Long: `Apply all DRA plugin manifests directly to a Kubernetes cluster. This will 
		create or update the necessary resources including ServiceAccount, ClusterRole, 
		ClusterRoleBinding, DaemonSet, DeviceClasses, and ValidatingAdmissionPolicy.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := cli.New()
			if err != nil {
				return err
			}

			return deploy.Deploy(context.Background(), c, params.EnvConfig{
				Namespace:   namespace,
				Image:       image,
				Command:     applyArgs.command,
				IsOpenshift: true,
			})
		},
	}
	parseApplyCmdFlags(applyCmd.PersistentFlags(), applyArgs)
	return applyCmd
}

func NewDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete",
		Short: "Delete DRA plugin manifests from a Kubernetes cluster",
		Long: `Delete all DRA plugin manifests from a Kubernetes cluster. If a namespace 
		is specified, deleting the namespace will automatically remove all namespaced resources 
		(ServiceAccount, DaemonSet). Cluster-scoped resources will be deleted explicitly.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := cli.New()
			if err != nil {
				return err
			}

			return deploy.Delete(context.Background(), c, namespace)
		},
	}
}

func parseApplyCmdFlags(flags *flag.FlagSet, args *applyArgs) {
	flags.StringVar(&args.command, "command", "", "Command pass for running the container")
}
