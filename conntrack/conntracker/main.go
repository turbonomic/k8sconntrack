package main

// Example usage

import (
	"fmt"
	"runtime"
	"time"

	"k8s.io/kubernetes/pkg/util"

	"github.com/spf13/pflag"

	"github.com/dongyiyang/k8sconnection/conntrack"
	"github.com/dongyiyang/k8sconnection/conntrack/k8sconnector"
	"github.com/dongyiyang/k8sconnection/conntrack/transactioncounter"

	"github.com/golang/glog"
)

func main() {
	connCounterMap := make(map[string]int64)

	k8sconnector, err := createK8sConnector()
	if err != nil {
		glog.Fatalf("Cannot create require connection monitor: %s", err)
	}
	conntrack.SetK8sConnector(k8sconnector)

	transactionCounter := transactioncounter.NewTransactionCounter()
	go transactioncounter.ListenAndServeProxyServer(transactionCounter)

	// TODO: do we really need the get it the established when start?
	cs, err := conntrack.Established()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Established on start:\n")
	for _, cn := range cs {
		fmt.Printf(" - %s\n", cn)
	}
	fmt.Println("")

	c, err := conntrack.New()
	if err != nil {
		panic(err)
	}
	for range time.Tick(1 * time.Second) {
		fmt.Printf("Connections:\n")
		for _, cn := range c.Connections() {
			// fmt.Printf(" - %s\n", cn)
			address := cn.Local
			connCounterMap = count(connCounterMap, address)
			svcName, err := k8sconnector.GetServiceNameWithEndpointAddress(address)
			if err != nil {
				fmt.Printf(" - %s,\t count: %d,\tError getting svc name\n", cn, connCounterMap[address])
			}
			fmt.Printf(" - %s,\t count: %d,\t%s\n", cn, connCounterMap[address], svcName)
			transactionCounter.Count(svcName, address)
		}
	}
}

func createK8sConnector() (*k8sconnector.K8sConnector, error) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	s := k8sconnector.NewK8sConnectorBuilder().AddFlags(pflag.CommandLine)

	util.InitFlags()
	util.InitLogs()
	defer util.FlushLogs()

	// monitor, err := s.Build(pflag.CommandLine.Args())
	monitor, err := s.Build()
	if err != nil {
		return nil, err
	}

	return monitor, nil
}

func count(connCounterMap map[string]int64, src string) map[string]int64 {
	count, exist := connCounterMap[src]
	if !exist {
		count = 0
	}
	connCounterMap[src] = count + 1
	return connCounterMap
}
