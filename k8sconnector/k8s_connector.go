package k8sconnector

import (
	"fmt"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"

	"github.com/dongyiyang/k8sconnection/k8sconnector/getter"

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

// Initialize the endpoint to service map.
// First get all the endpoint from Kubernetes cluster; Then create map.
func initEndpointMap(c *client.Client) error {
	glog.V(3).Infof("Initialize endpoint map.")
	endpoints, err := getAllEpsInK8s(c)
	if err != nil {
		return err
	}
	initEpSvcMap(endpointServiceMap, endpoints)

	return nil
}

func initEpSvcMap(epSvcMap map[string]string, endpoints []*api.Endpoints) {
	for _, endpoint := range endpoints {
		for _, endpointSubset := range endpoint.Subsets {
			addresses := endpointSubset.Addresses
			for _, address := range addresses {
				target := address.TargetRef
				if target == nil {
					continue
				}
				nameWithNamespace := endpoint.Namespace + "/" + endpoint.Name

				epSvcMap[address.IP] = nameWithNamespace
			}
		}
	}
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
	ips := extractValidPodsIPs(pods, this.nodeAddress)
	return ips, nil
}

// Here we only want to get the IP of those pods,
// which are hosted by the current node with its own unique ip.
func extractValidPodsIPs(pods []*api.Pod, hostAddress string) []string {
	var ips []string
	for _, pod := range pods {
		if pod.Status.HostIP == hostAddress && pod.Status.PodIP != hostAddress {
			ips = append(ips, pod.Status.PodIP)
			glog.V(4).Infof("Get pod %s,\t with IP %s", pod.Namespace+"/"+pod.Name, pod.Status.PodIP)
		}
	}
	return ips
}

// Retrieve service name based on endpoint address.
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

// Create Kuberntes endpoint getter and retrieve all the endpoints.
func getAllEpsInK8s(c *client.Client) ([]*api.Endpoints, error) {
	epGetter := getter.NewK8sEndpointGetter(c)
	namespace := api.NamespaceAll
	label := labels.Everything()
	return epGetter.GetEndpoints(namespace, label)
}

func (this *K8sConnector) updateEndpointMap(epAddress string) (string, error) {
	endpoints, err := getAllEpsInK8s(this.kubeClient)
	if err != nil {
		return "", fmt.Errorf("Cannot update %s: %s", epAddress, err)
	}
	return updateEpSvcMap(endpointServiceMap, epAddress, endpoints)
}

func updateEpSvcMap(epSvcMap map[string]string, epAddress string, endpoints []*api.Endpoints) (string, error) {
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

					epSvcMap[epAddress] = nameWithNamespace
					return nameWithNamespace, nil
				}
			}
		}
	}
	return "", fmt.Errorf("Endpoint with IP address %s is not found.", epAddress)
}
