package k8sconnector

import (
	"fmt"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"

	"github.com/dongyiyang/k8sconnection/conntrack/k8sconnector/getter"

	"github.com/golang/glog"
)

// A map translating endpoint address to corresponding service.
// We build this so that we do not need to iterate over all the endpoint set every time.
var endpointServiceMap map[string]string = make(map[string]string)

// A K8sConnector is used to connect to the Kubernetes cluster and get info from cluster.
type K8sConnector struct {
	kubeClient  *client.Client
	nodeAddress string
}

func NewK8sConnector(c *client.Client, addr string) *K8sConnector {
	glog.V(3).Infof("Building Kubernetes client")
	initEndpointMap(c)
	return &K8sConnector{
		kubeClient:  c,
		nodeAddress: addr,
	}
}

func initEndpointMap(c *client.Client) error {
	glog.V(3).Infof("Initialize endpoint map.")

	epGetter := getter.NewK8sEndpointGetter(c)
	namespace := api.NamespaceAll
	label := labels.Everything()
	endpoints, err := epGetter.GetEndpoints(namespace, label)
	if err != nil {
		return err
	}
	for _, endpoint := range endpoints {
		for _, endpointSubset := range endpoint.Subsets {
			addresses := endpointSubset.Addresses
			for _, address := range addresses {
				target := address.TargetRef
				if target == nil {
					continue
				}
				nameWithNamespace := endpoint.Namespace + "/" + endpoint.Name

				endpointServiceMap[address.IP] = nameWithNamespace
			}
		}
	}

	return nil

}

// Get the ip of pods those are running on the current node.
func (this *K8sConnector) GetPodsIPsOnNode() ([]string, error) {
	glog.V(3).Infof("Now get pod running on current node.")

	podGetter := getter.NewK8sPodGetter(this.kubeClient)
	namespace := api.NamespaceAll
	label := labels.Everything()
	field := fields.Everything()
	pods, err := podGetter.GetPods(namespace, label, field)
	if err != nil {
		return nil, err
	}
	var ips []string
	for _, pod := range pods {
		if pod.Status.HostIP == this.nodeAddress && pod.Status.PodIP != this.nodeAddress {
			ips = append(ips, pod.Status.PodIP)
			glog.V(3).Infof("Get pod %s,\t with IP %s", pod.Namespace+"/"+pod.Name, pod.Status.PodIP)
		}
	}
	return ips, nil
}

func (this *K8sConnector) GetServiceNameWithEndpointAddress(address string) (string, error) {
	serviceName, exist := endpointServiceMap[address]
	// A better solution is to start a watcher for the endpoint changes.
	if !exist {
		// Insert new.
		newSvcName, err := this.updateEndpointMap(address)
		if err != nil {
			return "", err
		}
		serviceName = newSvcName
	}
	return serviceName, nil
}

func (this *K8sConnector) getAllEndpointsInK8s() ([]*api.Endpoints, error) {
	epGetter := getter.NewK8sEndpointGetter(this.kubeClient)
	namespace := api.NamespaceAll
	label := labels.Everything()
	return epGetter.GetEndpoints(namespace, label)
}

func (this *K8sConnector) updateEndpointMap(epAddress string) (string, error) {
	endpoints, err := this.getAllEndpointsInK8s()
	if err != nil {
		return "", fmt.Errorf("Cannot update %s: %s", epAddress, err)
	}
	for _, endpoint := range endpoints {
		for _, endpointSubset := range endpoint.Subsets {
			addresses := endpointSubset.Addresses
			for _, address := range addresses {
				if address.IP == epAddress {
					target := address.TargetRef
					if target == nil {
						return "", fmt.Errorf("Cannot update %s due to nil targetRef", epAddress)
					}
					nameWithNamespace := endpoint.Namespace + "/" + endpoint.Name

					endpointServiceMap[epAddress] = nameWithNamespace
					return nameWithNamespace, nil
				}
			}
		}
	}
	return "", fmt.Errorf("Endpoint with IP address %s is not found.", epAddress)
}
