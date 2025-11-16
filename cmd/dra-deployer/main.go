package main

import (
	"os"

	"k8s.io/klog/v2"

	"github.com/Tal-or/dra-deployer/pkg/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		klog.ErrorS(err, "Error executing command")
		os.Exit(1)
	}
}
