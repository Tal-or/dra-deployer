package image

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantImage string
		wantTag   string
		wantErr   bool
	}{
		{
			name:      "full image with tag",
			input:     "quay.io/organization/image:v1.0.0",
			wantImage: "quay.io/organization/image",
			wantTag:   "v1.0.0",
			wantErr:   false,
		},
		{
			name:      "image with latest tag",
			input:     "docker.io/library/nginx:latest",
			wantImage: "docker.io/library/nginx",
			wantTag:   "latest",
			wantErr:   false,
		},
		{
			name:      "localhost registry with tag",
			input:     "localhost:5000/myimage:dev",
			wantImage: "localhost:5000/myimage",
			wantTag:   "dev",
			wantErr:   false,
		},
		{
			name:      "image with numeric tag",
			input:     "registry.example.com/app:123",
			wantImage: "registry.example.com/app",
			wantTag:   "123",
			wantErr:   false,
		},
		{
			name:      "image with semantic version tag",
			input:     "ghcr.io/owner/repo:v1.2.3-alpha.1",
			wantImage: "ghcr.io/owner/repo",
			wantTag:   "v1.2.3-alpha.1",
			wantErr:   false,
		},
		{
			name:      "image without tag",
			input:     "quay.io/organization/image",
			wantImage: "quay.io/organization/image",
			wantTag:   "latest",
			wantErr:   false,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid image - just tag",
			input:   ":v1.0",
			wantErr: true,
		},
		{
			name:    "invalid image - uppercase",
			input:   "UPPERCASE:tag",
			wantErr: true,
		},
		{
			name:      "docker hub with org and tag",
			input:     "docker.io/myuser/myapp:stable",
			wantImage: "docker.io/myuser/myapp",
			wantTag:   "stable",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if got.Image != tt.wantImage {
				t.Errorf("Parse(%q).Image = %q, want %q", tt.input, got.Image, tt.wantImage)
			}

			if got.Tag != tt.wantTag {
				t.Errorf("Parse(%q).Tag = %q, want %q", tt.input, got.Tag, tt.wantTag)
			}
		})
	}
}

func TestReferenceString(t *testing.T) {
	tests := []struct {
		name string
		ref  Reference
		want string
	}{
		{
			name: "standard image",
			ref: Reference{
				Image: "quay.io/org/image",
				Tag:   "v1.0.0",
			},
			want: "quay.io/org/image:v1.0.0",
		},
		{
			name: "image with latest tag",
			ref: Reference{
				Image: "nginx",
				Tag:   "latest",
			},
			want: "nginx:latest",
		},
		{
			name: "localhost registry",
			ref: Reference{
				Image: "localhost:5000/myapp",
				Tag:   "dev",
			},
			want: "localhost:5000/myapp:dev",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ref.String()
			if got != tt.want {
				t.Errorf("Reference.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseRoundTrip(t *testing.T) {
	// Test that Parse -> String produces the expected output
	tests := []struct {
		input string
		want  string
	}{
		{
			input: "quay.io/org/image:v1.0.0",
			want:  "quay.io/org/image:v1.0.0",
		},
		{
			input: "docker.io/library/nginx:latest",
			want:  "docker.io/library/nginx:latest",
		},
		{
			input: "localhost:5000/app:dev",
			want:  "localhost:5000/app:dev",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			ref, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse(%q) failed: %v", tt.input, err)
			}

			got := ref.String()
			if got != tt.want {
				t.Errorf("Parse(%q).String() = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
