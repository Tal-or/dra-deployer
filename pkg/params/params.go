package params

type EnvConfig struct {
	Namespace    string
	NodeSelector map[string]string // NodeSelector to be applied to the daemonset pods
	Image        string
	Command      string
	IsOpenshift  bool // If true, the deployment is running on OpenShift
	Values       map[string]any
}
