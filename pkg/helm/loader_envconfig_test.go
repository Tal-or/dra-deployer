package helm

import (
	"path/filepath"
	"testing"

	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform"

	"github.com/Tal-or/dra-deployer/pkg/image"
	"github.com/Tal-or/dra-deployer/pkg/params"
)

func TestRenderWithEnvConfig(t *testing.T) {
	chartPath := filepath.Join("..", "..", "assets", "deployment", "helm", "dra-driver-memory")
	loader, err := NewChartLoader(chartPath)
	if err != nil {
		t.Fatalf("Failed to create chart loader: %v", err)
	}

	// Test with all envConfig fields set
	envConfig := params.EnvConfig{
		Namespace: "test-namespace",
		Image:     "quay.io/myorg/dra-memory-driver:v0.5.0",
		Command:   "my-custom-command",
		Platform:  platform.OpenShift,
	}

	objects, err := loader.Render(envConfig)
	if err != nil {
		t.Fatalf("Failed to render chart: %v", err)
	}

	if len(objects) == 0 {
		t.Fatal("Expected at least one object to be rendered")
	}

	t.Logf("Rendered %d objects", len(objects))

	// Verify the objects contain our expected values
	foundDaemonSet := false
	foundSCC := false

	for _, obj := range objects {
		kind := obj.GetKind()
		name := obj.GetName()
		t.Logf("Object: %s/%s", kind, name)

		switch kind {
		case "DaemonSet":
			foundDaemonSet = true
			// Check if image is set correctly
			containers, found, err := getNestedSlice(obj.Object, "spec", "template", "spec", "containers")
			if err != nil || !found {
				t.Errorf("Failed to get containers from DaemonSet: %v", err)
				continue
			}
			if len(containers) == 0 {
				t.Error("No containers found in DaemonSet")
				continue
			}
			container := containers[0].(map[string]interface{})
			image, ok := container["image"].(string)
			if !ok {
				t.Error("Image not found in container")
				continue
			}
			expectedImage := "quay.io/myorg/dra-memory-driver:v0.5.0"
			if image != expectedImage {
				t.Errorf("Expected image %q, got %q", expectedImage, image)
			} else {
				t.Logf("Image correctly set to %q", image)
			}

			// Check if command is set
			command, found, err := getNestedSlice(container, "command")
			if err != nil || !found {
				t.Error("Command not found in container")
				continue
			}
			if len(command) == 0 {
				t.Error("Command is empty")
				continue
			}
			cmdStr := command[0].(string)
			if cmdStr != "my-custom-command" {
				t.Errorf("Expected command %q, got %q", "my-custom-command", cmdStr)
			} else {
				t.Logf("Command correctly set to %q", cmdStr)
			}

		case "SecurityContextConstraints":
			foundSCC = true
			t.Logf("OpenShift SCC was created (IsOpenshift=true)")
		}
	}

	if !foundDaemonSet {
		t.Error("DaemonSet not found in rendered objects")
	}

	if !foundSCC {
		t.Error("SecurityContextConstraints not found (expected when IsOpenshift=true)")
	}
}

func TestRenderWithoutOpenShift(t *testing.T) {
	chartPath := filepath.Join("..", "..", "assets", "deployment", "helm", "dra-driver-memory")
	loader, err := NewChartLoader(chartPath)
	if err != nil {
		t.Fatalf("Failed to create chart loader: %v", err)
	}

	// Test with different platform than OpenShift
	envConfig := params.EnvConfig{
		Namespace: "test-namespace",
		Platform:  platform.Kubernetes,
	}

	objects, err := loader.Render(envConfig)
	if err != nil {
		t.Fatalf("Failed to render chart: %v", err)
	}

	// Verify SCC is NOT created
	for _, obj := range objects {
		if obj.GetKind() == "SecurityContextConstraints" {
			t.Error("SecurityContextConstraints should not be created when platform is not OpenShift")
		}
	}

	t.Log("SCC correctly not created when platform is not OpenShift")
}

func TestParseImage(t *testing.T) {
	tests := []struct {
		input          string
		wantRepository string
		wantTag        string
	}{
		{
			input:          "quay.io/org/image:v1.0",
			wantRepository: "quay.io/org/image",
			wantTag:        "v1.0",
		},
		{
			input:          "quay.io/org/image",
			wantRepository: "quay.io/org/image",
			wantTag:        "latest",
		},
		{
			input:          "localhost:5000/image:tag",
			wantRepository: "localhost:5000/image",
			wantTag:        "tag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			ref, err := image.Parse(tt.input)
			if err != nil {
				t.Fatalf("Failed to parse image: %v", err)
			}
			if ref.Image != tt.wantRepository {
				t.Errorf("parseImage(%q) repository = %q, want %q", tt.input, ref.Image, tt.wantRepository)
			}
			if ref.Tag != tt.wantTag {
				t.Errorf("parseImage(%q) tag = %q, want %q", tt.input, ref.Tag, tt.wantTag)
			}
		})
	}
}

// getNestedSlice is a helper to get nested slices from unstructured data
func getNestedSlice(obj map[string]interface{}, fields ...string) ([]interface{}, bool, error) {
	current := obj
	for i, field := range fields[:len(fields)-1] {
		val, ok := current[field]
		if !ok {
			return nil, false, nil
		}
		currentMap, ok := val.(map[string]interface{})
		if !ok {
			return nil, false, nil
		}
		current = currentMap
		_ = i
	}

	lastField := fields[len(fields)-1]
	val, ok := current[lastField]
	if !ok {
		return nil, false, nil
	}

	slice, ok := val.([]interface{})
	if !ok {
		return nil, false, nil
	}

	return slice, true, nil
}
