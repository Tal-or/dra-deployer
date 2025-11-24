package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"k8s.io/klog/v2"

	"sigs.k8s.io/yaml"

	"github.com/Tal-or/dra-deployer/pkg/manifests"
)

func NewRenderCommand() *cobra.Command {
	return &cobra.Command{
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
}

// Render renders all manifests to stdout as YAML
func render() error {
	klog.InfoS("Rendering manifests", "namespace", namespace, "image", image)

	m, err := manifests.GetAll(namespace, image)
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
