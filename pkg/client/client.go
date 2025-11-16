package client

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	securityv1 "github.com/openshift/api/security/v1"
)

var (
	// scheme contains all the types needed for the controller-runtime client
	scheme = runtime.NewScheme()
)

func init() {
	// Register standard Kubernetes types
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	// Register OpenShift security types
	utilruntime.Must(securityv1.Install(scheme))
}

// New creates a new controller-runtime client with all necessary types registered
func New() (client.Client, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}

	// Create client with the custom scheme
	cli, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return cli, nil
}
