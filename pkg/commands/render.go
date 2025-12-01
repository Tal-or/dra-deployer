package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"k8s.io/klog/v2"

	"sigs.k8s.io/yaml"

	"github.com/Tal-or/dra-deployer/pkg/helm"
	"github.com/Tal-or/dra-deployer/pkg/params"
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

	// Load Helm chart
	chartLoader, err := helm.NewChartLoader("")
	if err != nil {
		return fmt.Errorf("failed to load Helm chart: %w", err)
	}

	objects, err := chartLoader.Render(params.EnvConfig{
		Namespace: namespace,
	})
	if err != nil {
		return fmt.Errorf("failed to render Helm chart: %w", err)
	}

	// Serialize each manifest as YAML
	for i, obj := range objects {
		if i > 0 {
			fmt.Println("---")
		}

		yamlData, err := yaml.Marshal(obj)
		if err != nil {
			return fmt.Errorf("failed to marshal manifest to YAML: %w", err)
		}

		fmt.Print(string(yamlData))
	}

	klog.InfoS("Successfully rendered manifests")
	return nil
}
