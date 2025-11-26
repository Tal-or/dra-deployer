package commands

import (
	"flag"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	namespace string
	verbosity int
	image     string
)

const (
	defaultImage     = "quay.io/titzhak/dra-example-driver:v0.1.0"
	defaultNamespace = "dra-deployer"
	defaultVerbosity = 2
)

func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "dra-deployer",
		Short: "Command line tool for deploying DRA plugins",
		Long:  `dra-deployer is a CLI tool that helps you deploy Dynamic Resource Allocation (DRA) plugins to Kubernetes clusters. It can render manifests to stdout or apply them directly to a cluster.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Set klog verbosity level based on the verbose flag
			flag.Set("v", fmt.Sprintf("%d", verbosity))
		},
	}

	parseFlags(rootCmd.PersistentFlags())

	rootCmd.AddCommand(NewRenderCommand())
	rootCmd.AddCommand(NewApplyCommand(&applyArgs{}))
	rootCmd.AddCommand(NewDeleteCommand())
	return rootCmd
}

// Execute runs the root command
func Execute() error {
	return NewRootCommand().Execute()
}

func parseFlags(flags *pflag.FlagSet) {
	flags.IntVarP(&verbosity, "verbose", "v", defaultVerbosity, "Log level verbosity")
	flags.StringVarP(&namespace, "namespace", "n", defaultNamespace, "Namespace for namespaced resources")
	flags.StringVarP(&image, "image", "i", defaultImage, "Container image for the DRA plugin")
}
