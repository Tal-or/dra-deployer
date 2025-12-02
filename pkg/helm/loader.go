package helm

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/klog/v2"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"

	"github.com/Tal-or/dra-deployer/pkg/image"
	"github.com/Tal-or/dra-deployer/pkg/params"

	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform"
)

const (
	// DefaultChartPath is the default path to the Helm chart directory
	DefaultChartPath = "assets/deployment/helm/dra-driver-memory"
)

// ChartLoader loads and renders Helm charts
type ChartLoader struct {
	chart *chart.Chart
}

// NewChartLoader creates a new ChartLoader by loading the Helm chart from the filesystem
// If chartPath is empty, it uses DefaultChartPath
func NewChartLoader(chartPath string) (*ChartLoader, error) {
	if chartPath == "" {
		chartPath = DefaultChartPath
	}

	klog.V(4).InfoS("Loading Helm chart from filesystem", "path", chartPath)

	// Load the chart using Helm's loader
	chart, err := loader.Load(chartPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load Helm chart from %s: %w", chartPath, err)
	}

	klog.V(4).InfoS("Loaded Helm chart", "name", chart.Name(), "version", chart.Metadata.Version)

	return &ChartLoader{
		chart: chart,
	}, nil
}

// Render renders the Helm chart with the given options and returns Kubernetes objects
func (l *ChartLoader) Render(envConfig params.EnvConfig) ([]*unstructured.Unstructured, error) {
	releaseName := l.chart.Metadata.AppVersion
	klog.V(4).InfoS("Rendering Helm chart", "release", releaseName, "namespace", envConfig.Namespace)

	// Start with default values from values.yaml
	values := l.chart.Values

	// Build runtime values from envConfig
	runtimeValues, err := buildValuesFromEnvConfig(envConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build values from envConfig: %w", err)
	}

	// Merge runtime values with defaults (runtime values take precedence)
	if len(runtimeValues) > 0 {
		values = chartutil.CoalesceTables(runtimeValues, values)
	}

	// Merge any additional custom values provided
	if envConfig.Values != nil {
		values = chartutil.CoalesceTables(envConfig.Values, values)
	}

	// Set up release options
	releaseOptions := chartutil.ReleaseOptions{
		Name:      releaseName,
		Namespace: envConfig.Namespace,
		IsInstall: true,
	}

	// Convert values to renderable format
	valuesToRender, err := chartutil.ToRenderValues(l.chart, values, releaseOptions, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare values for rendering: %w", err)
	}

	// Render templates
	rendered, err := engine.Render(l.chart, valuesToRender)
	if err != nil {
		return nil, fmt.Errorf("failed to render templates: %w", err)
	}

	// Convert rendered YAML to Kubernetes objects
	objects, err := parseRenderedTemplates(rendered)
	if err != nil {
		return nil, fmt.Errorf("failed to parse rendered templates: %w", err)
	}

	klog.V(4).InfoS("Successfully rendered Helm chart", "objectCount", len(objects))
	return objects, nil
}

// GetChart returns the loaded Helm chart
func (l *ChartLoader) GetChart() *chart.Chart {
	return l.chart
}

// buildValuesFromEnvConfig converts EnvConfig fields into Helm values
func buildValuesFromEnvConfig(envConfig params.EnvConfig) (map[string]any, error) {
	values := make(map[string]any)

	// Set image if provided
	if envConfig.Image != "" {
		// Parse image into repository and tag
		ref, err := image.Parse(envConfig.Image)
		if err != nil {
			return nil, fmt.Errorf("failed to parse image: %w", err)
		}
		values["image"] = map[string]any{
			"repository": ref.Image,
			"tag":        ref.Tag,
		}
		klog.V(5).InfoS("Set image reference from envConfig", "envConfig.Image", envConfig.Image, "image", ref.Image, "tag", ref.Tag)
	}

	// Set OpenShift flag
	values["openshift"] = map[string]any{
		"enabled": envConfig.Platform == platform.OpenShift,
	}
	klog.V(5).InfoS("Platform detected", "platform", envConfig.Platform)

	// Build daemonset configuration
	daemonsetValues := make(map[string]any)

	// Set command if provided
	if envConfig.Command != "" {
		daemonsetValues["command"] = []string{envConfig.Command}
		klog.V(5).InfoS("Set command from envConfig", "command", envConfig.Command)
	}

	// Set node selector if provided
	if len(envConfig.NodeSelector) > 0 {
		daemonsetValues["nodeSelector"] = envConfig.NodeSelector
		klog.V(5).InfoS("Set nodeSelector from envConfig", "nodeSelector", envConfig.NodeSelector)
	}

	// Only set daemonset values if we have any
	if len(daemonsetValues) > 0 {
		values["daemonset"] = daemonsetValues
	}

	return values, nil
}

// parseRenderedTemplates converts rendered YAML templates to Kubernetes objects
func parseRenderedTemplates(rendered map[string]string) ([]*unstructured.Unstructured, error) {
	var objects []*unstructured.Unstructured

	for name, content := range rendered {
		// Skip non-YAML files (like NOTES.txt, helpers.tpl, etc.)
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			klog.V(6).InfoS("Skipping non-YAML file", "name", name)
			continue
		}

		// Skip empty content
		content = strings.TrimSpace(content)
		if content == "" {
			klog.V(6).InfoS("Skipping empty template", "name", name)
			continue
		}

		// Parse YAML content (may contain multiple documents separated by ---)
		objs, err := parseYAMLDocuments(content)
		if err != nil {
			return nil, fmt.Errorf("failed to parse template %s: %w", name, err)
		}

		objects = append(objects, objs...)
		klog.V(5).InfoS("Parsed template", "name", name, "objects", len(objs))
	}

	return objects, nil
}

// parseYAMLDocuments parses YAML content that may contain multiple documents
func parseYAMLDocuments(content string) ([]*unstructured.Unstructured, error) {
	var objects []*unstructured.Unstructured

	// Split by document separator
	decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader([]byte(content)), 4096)

	for {
		obj := &unstructured.Unstructured{}
		err := decoder.Decode(obj)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to decode YAML: %w", err)
		}

		if len(obj.Object) == 0 {
			continue
		}

		objects = append(objects, obj)
	}

	return objects, nil
}
