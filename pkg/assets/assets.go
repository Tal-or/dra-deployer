package assets

import (
	"embed"
)

var (
	// Yamls contains all YAML files placed under the yamls directory
	//go:embed yamls
	Yamls embed.FS
)
