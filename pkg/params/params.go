package params

type EnvConfig struct {
	Namespace   string
	Image       string
	Command     string
	IsOpenshift bool // If true, the deployment is running on OpenShift
	Values      map[string]any
}
