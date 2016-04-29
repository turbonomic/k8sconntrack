package main

import (
	"runtime"
	"time"

	"k8s.io/kubernetes/pkg/util"

	"github.com/spf13/pflag"

	"github.com/dongyiyang/k8sconnection/pkg/conntrack"
	"github.com/dongyiyang/k8sconnection/pkg/k8sconnector"
	"github.com/dongyiyang/k8sconnection/pkg/transactioncounter"

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
	glog.V(3).Infof("Established on start:\n")
	for _, cn := range cs {
		glog.V(3).Infof(" - %s\n", cn)
	}

	c, err := conntrack.New()
	if err != nil {
		panic(err)
	}
	for range time.Tick(1 * time.Second) {
		connections := c.Connections()
		if len(connections) > 0 {
			glog.V(3).Infof("Connections:\n")
			for _, cn := range connections {
				// fmt.Printf(" - %s\n", cn)
				address := cn.Local
				connCounterMap = count(connCounterMap, address)
				svcName, err := k8sconnector.GetServiceNameWithEndpointAddress(address)
				if err != nil {
					glog.Errorf(" - %s,\t count: %d,\tError getting svc name\n", cn, connCounterMap[address])
				}
				glog.V(3).Infof(" - %s,\t count: %d,\t%s\n", cn, connCounterMap[address], svcName)
				transactionCounter.Count(svcName, address+":")
			}
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
