package cmd

import (
	"context"
	"flag"
	"fmt"

	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"

	"github.com/spf13/cobra"

	cli "github.com/Tal-or/dra-deployer/pkg/client"
	"github.com/Tal-or/dra-deployer/pkg/deploy"
	"github.com/Tal-or/dra-deployer/pkg/manifests"
)

var (
	namespace string
	verbosity int
)

var rootCmd = &cobra.Command{
	Use:   "dra-deployer",
	Short: "Command line tool for deploying DRA plugins",
	Long: `dra-deployer is a CLI tool that helps you deploy Dynamic Resource Allocation (DRA) 
plugins to Kubernetes clusters. It can render manifests to stdout or apply them directly 
to a cluster.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Set klog verbosity level based on the verbose flag
		flag.Set("v", fmt.Sprintf("%d", verbosity))
	},
}

var renderCmd = &cobra.Command{
	Use:   "render",
	Short: "Render DRA plugin manifests to stdout",
	Long: `Render all DRA plugin manifests as YAML to stdout. This is useful for 
reviewing the manifests before applying them or for piping to kubectl apply.`,
	Example: `  # Render manifests with default namespace
  dra-deployer render

  # Render manifests with custom namespace
  dra-deployer render --namespace my-namespace`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return render()
	},
}

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply DRA plugin manifests to a Kubernetes cluster",
	Long: `Apply all DRA plugin manifests directly to a Kubernetes cluster. This will 
create or update the necessary resources including ServiceAccount, ClusterRole, 
ClusterRoleBinding, DaemonSet, DeviceClasses, and ValidatingAdmissionPolicy.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return apply()
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete DRA plugin manifests from a Kubernetes cluster",
	Long: `Delete all DRA plugin manifests from a Kubernetes cluster. If a namespace 
is specified, deleting the namespace will automatically remove all namespaced resources 
(ServiceAccount, DaemonSet). Cluster-scoped resources will be deleted explicitly.`,
	Example: `  # Delete with default namespace
  dra-deployer delete

  # Delete with custom namespace
  dra-deployer delete --namespace my-namespace`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return delete()
	},
}

func init() {
	// Initialize klog flags (for internal use)
	klog.InitFlags(nil)

	// Add verbose flag to root command
	rootCmd.PersistentFlags().IntVarP(&verbosity, "verbose", "v", 2, "Log level verbosity")

	// Add namespace flag to render command
	renderCmd.Flags().StringVarP(&namespace, "namespace", "n", "dra-deployer",
		"Namespace for namespaced resources")

	// Add namespace flag to apply command
	applyCmd.Flags().StringVarP(&namespace, "namespace", "n", "dra-deployer",
		"Namespace for namespaced resources")

	// Add namespace flag to delete command
	deleteCmd.Flags().StringVarP(&namespace, "namespace", "n", "dra-deployer",
		"Namespace for namespaced resources")

	// Add subcommands
	rootCmd.AddCommand(renderCmd)
	rootCmd.AddCommand(applyCmd)
	rootCmd.AddCommand(deleteCmd)
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

// Render renders all manifests to stdout as YAML
func render() error {
	klog.InfoS("Rendering manifests", "namespace", namespace)

	m, err := manifests.GetAll(namespace)
	if err != nil {
		return fmt.Errorf("failed to get manifests: %w", err)
	}

	// Get DeviceClasses separately since they're not included in GetAll
	deviceClasses, err := manifests.GetDeviceClasses()
	if err != nil {
		return fmt.Errorf("failed to get DeviceClasses: %w", err)
	}

	// Render all manifests
	manifestsList := []interface{}{
		m.ServiceAccount,
		m.SecurityContextConstraints,
		m.ClusterRole,
		m.ClusterRoleBinding,
		m.DaemonSet,
		m.ValidatingAdmissionPolicy,
		m.ValidatingAdmissionPolicyBinding,
	}

	// Add DeviceClasses
	for _, dc := range deviceClasses {
		manifestsList = append(manifestsList, dc)
	}

	// Serialize each manifest as YAML
	for i, manifest := range manifestsList {
		if i > 0 {
			fmt.Println("---")
		}

		yamlData, err := yaml.Marshal(manifest)
		if err != nil {
			return fmt.Errorf("failed to marshal manifest to YAML: %w", err)
		}

		fmt.Print(string(yamlData))
	}

	klog.InfoS("Successfully rendered manifests")
	return nil
}

// deploy applies all manifests to the Kubernetes cluster
func apply() error {
	// Create client with all necessary types registered
	c, err := cli.New()
	if err != nil {
		return err
	}

	return deploy.Deploy(context.Background(), c, namespace)
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
