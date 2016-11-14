package app

import (
	"fmt"
	"time"

	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
	proxyconfig "k8s.io/kubernetes/pkg/proxy/config"

	"github.com/dongyiyang/k8sconnection/cmd/app/options"
	"github.com/dongyiyang/k8sconnection/pkg/conntrack"
	"github.com/dongyiyang/k8sconnection/pkg/flowcollector"
	"github.com/dongyiyang/k8sconnection/pkg/server"
	"github.com/dongyiyang/k8sconnection/pkg/transactioncounter"

	"github.com/golang/glog"
)

type K8sConntrackServer struct {
	config             *options.K8sConntrackConfig
	transactionCounter *transactioncounter.TransactionCounter
	flowCollector      *flowcollector.FlowCollector
}

func NewK8sConntrackServer(config *options.K8sConntrackConfig) (*K8sConntrackServer, error) {
	// Set up sysctl variables
	conntracker := realConntracker{}
	if err := conntracker.EnableAcct(); err != nil {
		return nil, fmt.Errorf("Error setting netfilter_conntrack_acct: %++v", err)
	}
	if err := conntracker.EnableTimestamp(); err != nil {
		return nil, fmt.Errorf("Error setting netfilter_conntrack_timestamp: %++v", err)
	}
	switch config.SocketBufferSize {
	case "medium":
		conntracker.SetSocketReadBuffer(33554432)
	}

	if config.Kubeconfig == "" && config.Master == "" {
		return nil, fmt.Errorf("Neither --kubeconfig nor --master was specified.  Using default API client.  This might not work.")
	}

	glog.V(3).Infof("Master is %s", config.Master)

	kubeconfig, err := clientcmd.BuildConfigFromFlags(config.Master, config.Kubeconfig)
	if err != nil {
		glog.Errorf("Error getting kubeconfig:  %s", err)
		return nil, err
	}
	// This specifies the number and the max number of query per second to the api server.
	kubeconfig.QPS = 20.0
	kubeconfig.Burst = 30

	kubeClient, err := client.New(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("Invalid API configuration: %v", err)
	}

	c, err := conntrack.New(conntrack.DefaultFilter)
	if err != nil {
		panic(err)
	}

	endpointsConfig := proxyconfig.NewEndpointsConfig()

	var transactionCounter *transactioncounter.TransactionCounter
	if config.EnableConnectionCounter {
		glog.V(3).Infof("Connection Counter Enabled.")
		transactionCounter = transactioncounter.NewTransactionCounter(c)
		endpointsConfig.RegisterHandler(transactionCounter)
	}
	var flowCollector *flowcollector.FlowCollector
	if config.EnableFlowCollector {
		glog.V(3).Infof("Flow Collector Enabled.")
		flowCollector = flowcollector.NewFlowCollector(c)
		endpointsConfig.RegisterHandler(flowCollector)
	}

	proxyconfig.NewSourceAPI(
		kubeClient,
		time.Second*10,
		nil,
		endpointsConfig.Channel("api"))

	return &K8sConntrackServer{
		config,
		transactionCounter,
		flowCollector,
	}, nil
}

func (this *K8sConntrackServer) Run() {
	go server.ListenAndServeProxyServer(this.config.ConntrackBindAddress, this.config.ConntrackPort, this.transactionCounter, this.flowCollector)

	// Collect transaction and flow information every second.
	for range time.Tick(1 * time.Second) {

		if this.transactionCounter != nil {
			glog.V(3).Infof("~~~~~~~~~~~~~~~~   Transaction Counter	~~~~~~~~~~~~~~~~~~~~")
			this.transactionCounter.ProcessConntrackConnections()
		}

		if this.flowCollector != nil {
			glog.V(3).Infof("----------------   Flow Collector	------------------------")
			this.flowCollector.TrackFlow()
		}
		glog.V(3).Infof("##########################################################")
		fmt.Println()
	}
}
