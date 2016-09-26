package main

import (
	"fmt"
	"runtime"
	"time"

	"k8s.io/kubernetes/pkg/util/flag"
	"k8s.io/kubernetes/pkg/util/logs"

	"github.com/spf13/pflag"

	"github.com/dongyiyang/k8sconnection/pkg/conntrack"
	"github.com/dongyiyang/k8sconnection/pkg/flowcollector"
	"github.com/dongyiyang/k8sconnection/pkg/k8sconnector"
	"github.com/dongyiyang/k8sconnection/pkg/server"
	"github.com/dongyiyang/k8sconnection/pkg/transactioncounter"

	"github.com/golang/glog"
)

func main() {
	k8scon, err := createK8sConnector()
	if err != nil {
		glog.Fatalf("Cannot create require connection monitor: %s", err)
	}
	conntrack.SetK8sConnector(k8scon)

	c, err := conntrack.New()
	if err != nil {
		panic(err)
	}
	transactionCounter := transactioncounter.NewTransactionCounter(k8scon, c)
	flowCollector := flowcollector.NewFlowCollector(k8scon)

	go server.ListenAndServeProxyServer(transactionCounter, flowCollector)

	// Collect transaction and flow information every second.
	for range time.Tick(1 * time.Second) {

		glog.V(3).Infof("~~~~~~~~~~~~~~~~   Transaction Counter	~~~~~~~~~~~~~~~~~~~~")
		transactionCounter.ProcessConntrackConnections()

		glog.V(3).Infof("----------------   Flow Collector	------------------------")
		flowCollector.TrackFlow()
		glog.V(3).Infof("##########################################################")
		fmt.Println()
	}
}

// Get variables and creates Kubernetes connector.
func createK8sConnector() (*k8sconnector.K8sConnector, error) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	s := k8sconnector.NewK8sConnectorBuilder().AddFlags(pflag.CommandLine)

	flag.InitFlags()
	logs.InitLogs()
	defer logs.FlushLogs()

	monitor, err := s.Build()
	if err != nil {
		return nil, err
	}

	return monitor, nil
}
