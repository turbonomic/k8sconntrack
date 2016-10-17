package main

import (
	"fmt"
	"os"

	"github.com/dongyiyang/k8sconnection/cmd/app"
	"github.com/dongyiyang/k8sconnection/cmd/app/options"

	"k8s.io/kubernetes/pkg/util/flag"
	"k8s.io/kubernetes/pkg/util/logs"

	"github.com/spf13/pflag"
)

func main() {
	config := options.NewK8sConntrackConfig()
	config.AddFlags(pflag.CommandLine)

	flag.InitFlags()
	logs.InitLogs()
	defer logs.FlushLogs()

	s, err := app.NewK8sConntrackServer(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	s.Run()
}
