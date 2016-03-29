package k8sconnector

import (
	"fmt"

	"k8s.io/kubernetes/pkg/client"
	"k8s.io/kubernetes/pkg/client/clientcmd"
	clientcmdapi "k8s.io/kubernetes/pkg/client/clientcmd/api"

	"github.com/golang/glog"
	"github.com/spf13/pflag"
)

type K8sConnectorBuilder struct {
	NodeAddress string
	Master      string
	Kubeconfig  string
}

func NewK8sConnectorBuilder() *K8sConnectorBuilder {
	return &K8sConnectorBuilder{}
}

// Add parameters passed from command line.
func (s *K8sConnectorBuilder) AddFlags(fs *pflag.FlagSet) *K8sConnectorBuilder {
	fs.StringVar(&s.NodeAddress, "node-addr", s.NodeAddress, "The address of current node")
	fs.StringVar(&s.Master, "master", s.Master, "The address of the Kubernetes API server (overrides any value in kubeconfig)")
	fs.StringVar(&s.Kubeconfig, "kubeconfig", s.Kubeconfig, "Path to kubeconfig file with authorization and master location information.")

	return s
}

func (s *K8sConnectorBuilder) Build() (*K8sConnector, error) {
	addr, err := s.getNodeAddress()
	if err != nil {
		return nil, err
	}
	kubeClient, err := s.createKubeClient()
	if err != nil {
		return nil, err
	}

	return NewK8sConnector(kubeClient, addr), nil
}

func (s *K8sConnectorBuilder) getNodeAddress() (string, error) {
	if s.NodeAddress == "" {
		return "", fmt.Errorf("NodeAddress is not set")
	}
	return s.NodeAddress, nil
}

// Create Kubernetes API client from
func (s *K8sConnectorBuilder) createKubeClient() (*client.Client, error) {
	if s.Kubeconfig == "" && s.Master == "" {
		return nil, fmt.Errorf("Neither --kubeconfig nor --master was specified.  Using default API client.  This might not work.")
	}

	glog.V(3).Infof("Master is %s", s.Master)

	// This creates a client, first loading any specified kubeconfig
	// file, and then overriding the Master flag, if non-empty.
	kubeconfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: s.Kubeconfig},
		&clientcmd.ConfigOverrides{ClusterInfo: clientcmdapi.Cluster{Server: s.Master}}).ClientConfig()
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

	return kubeClient, nil
}
