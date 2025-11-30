package helm

import (
	"path/filepath"
	"testing"

	"github.com/Tal-or/dra-deployer/pkg/params"
)

func TestNewChartLoader(t *testing.T) {
	// Test loading the Helm chart from filesystem
	// Use relative path from project root
	chartPath := filepath.Join("..", "..", "assets", "deployment", "helm", "dra-driver-memory")
	loader, err := NewChartLoader(chartPath)
	if err != nil {
		t.Fatalf("Failed to create chart loader: %v", err)
	}

	if loader == nil {
		t.Fatal("ChartLoader is nil")
	}

	chart := loader.GetChart()
	if chart == nil {
		t.Fatal("Chart is nil")
	}

	if chart.Name() != "dra-driver-memory" {
		t.Errorf("Expected chart name 'dra-driver-memory', got '%s'", chart.Name())
	}

	if chart.Metadata.Version != "0.1.0" {
		t.Errorf("Expected chart version '0.1.0', got '%s'", chart.Metadata.Version)
	}
}

func TestRenderChart(t *testing.T) {
	// Test rendering the Helm chart
	chartPath := filepath.Join("..", "..", "assets", "deployment", "helm", "dra-driver-memory")
	loader, err := NewChartLoader(chartPath)
	if err != nil {
		t.Fatalf("Failed to create chart loader: %v", err)
	}

	// Print chart files for debugging
	t.Logf("Chart has %d files", len(loader.chart.Files))
	for _, file := range loader.chart.Files {
		t.Logf("File: %s", file.Name)
	}
	t.Logf("Chart has %d templates", len(loader.chart.Templates))
	for _, tmpl := range loader.chart.Templates {
		t.Logf("Template: %s (%d bytes)", tmpl.Name, len(tmpl.Data))
	}
	t.Logf("Chart has %d raw files", len(loader.chart.Raw))
	for _, raw := range loader.chart.Raw {
		t.Logf("Raw file: %s", raw.Name)
	}

	opts := params.EnvConfig{
		Namespace: "test-namespace",
		Values:    nil, // Use default values
	}

	objects, err := loader.Render(opts)
	if err != nil {
		t.Fatalf("Failed to render chart: %v", err)
	}

	if len(objects) == 0 {
		t.Error("Expected at least one object to be rendered")
	}

	t.Logf("Rendered %d Kubernetes objects", len(objects))
}

func TestRenderChartWithCustomValues(t *testing.T) {
	// Test rendering with custom values
	chartPath := filepath.Join("..", "..", "assets", "deployment", "helm", "dra-driver-memory")
	loader, err := NewChartLoader(chartPath)
	if err != nil {
		t.Fatalf("Failed to create chart loader: %v", err)
	}

	customValues := map[string]interface{}{
		"image": map[string]interface{}{
			"tag": "v0.2.0",
		},
		"openshift": map[string]interface{}{
			"enabled": true,
		},
	}

	opts := params.EnvConfig{
		Namespace: "test-namespace",
		Values:    customValues,
	}

	objects, err := loader.Render(opts)
	if err != nil {
		t.Fatalf("Failed to render chart with custom values: %v", err)
	}

	if len(objects) == 0 {
		t.Error("Expected at least one object to be rendered")
	}

	t.Logf("Rendered %d Kubernetes objects with custom values", len(objects))
}
