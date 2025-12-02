package params

import "github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform"

type EnvConfig struct {
	Namespace    string
	NodeSelector map[string]string // NodeSelector to be applied to the daemonset pods
	Image        string
	Command      string
	Platform     platform.Platform // Platform of the cluster
	Values       map[string]any
}
