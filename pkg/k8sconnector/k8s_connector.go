package k8sconnector

import (
	"fmt"
	"net"
	"time"

	"k8s.io/kubernetes/pkg/api"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"

	"github.com/dongyiyang/k8sconnection/pkg/k8sconnector/getter"
	"github.com/dongyiyang/k8sconnection/pkg/k8sconnector/storage"

	"github.com/golang/glog"
)

// A map translating endpoint address to corresponding service.
// We build this so that we do not need to iterate over all the endpoint set every time.
var endpointServiceMap map[string]string = make(map[string]string)

// A K8sConnector is used to connect to the Kubernetes cluster and get info from cluster.
type K8sConnector struct {
	kubeClient  *client.Client
	nodeAddress map[string]struct{}

	podIPStore *storage.PodIPStore
}

func NewK8sConnector(c *client.Client) (*K8sConnector, error) {
	glog.V(3).Infof("Building Kubernetes client")
	initEndpointMap(c)
	nodeIPs, err := findNodeIPs()
	if err != nil {
		return nil, err
	}
	connector := &K8sConnector{
		kubeClient:  c,
		nodeAddress: nodeIPs,
		podIPStore:  storage.NewPodIPStore(),
	}

	go connector.syncK8s()

	return connector, nil
}

// Sync Kubernetes and update info periodically.
// So that we don't have to make API calls to Kubernetes API server every time
// for some operations, such as getting pods IPs.
func (this *K8sConnector) syncK8s() {
	// TODO: use time.sleep for now...
	for {
		// sync pod ip
		podIPs, err := this.retreivePodsIPsOnNode()
		if err != nil {
			glog.Fatalf("Error getting pod ips from Kubernets cluster: %s.\n Exit...", err)
		}
		this.podIPStore.DeleteAll()
		for _, podIP := range podIPs {
			this.podIPStore.Add(podIP)
		}

		time.Sleep(time.Second * 10)
	}
}

// Find the all valid IP address of the node.
func findNodeIPs() (map[string]struct{}, error) {
	var l = map[string]struct{}{}
	if localNets, err := net.InterfaceAddrs(); err == nil {
		// Not all networks are IP networks.
		for _, localNet := range localNets {
			if network, ok := localNet.(*net.IPNet); ok {
				netAddress := network.IP.String()
				glog.Infof("Find network address %s", netAddress)

				ipAddress := net.ParseIP(netAddress)
				if ipAddress.To4() == nil {
					// Only support IPv4 now
					continue
				}
				glog.Infof("Find valid IPv4 address %s", netAddress)

				l[netAddress] = struct{}{}
			}
		}
		return l, nil
	} else {
		glog.Errorf("Error is %s", err)
		return nil, fmt.Errorf("Error finding IP address of current node: %s", err)
	}
}

// Initialize the endpoint to service map.
// First get all the endpoint from Kubernetes cluster; Then create map.
func initEndpointMap(c *client.Client) error {
	glog.V(3).Infof("Initialize endpoint map.")
	endpoints, err := GetAllEpsInK8s(c)
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

	return this.retreivePodsIPsOnNode()
}

// Get the ip of pods those are running on the current node.
func (this *K8sConnector) FindPodsIP() map[string]interface{} {
	return this.podIPStore.GetPodIPSets()
}

// Get the ip of pods those are running on the current node.
func (this *K8sConnector) retreivePodsIPsOnNode() ([]string, error) {
	// glog.V(3).Infof("Now get pod running on current node.")

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
func extractValidPodsIPs(pods []*api.Pod, hostAddresses map[string]struct{}) []string {
	var ips []string
	for _, pod := range pods {
		if _, exist := hostAddresses[pod.Status.HostIP]; exist {
			if _, has := hostAddresses[pod.Status.PodIP]; !has {
				ips = append(ips, pod.Status.PodIP)
				// glog.V(3).Infof("Get pod %s,\t with IP %s", pod.Namespace+"/"+pod.Name, pod.Status.PodIP)
				podIPMap[pod.Status.PodIP] = pod.Name
			}
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
func GetAllEpsInK8s(c *client.Client) ([]*api.Endpoints, error) {
	epGetter := getter.NewK8sEndpointGetter(c)
	namespace := api.NamespaceAll
	label := labels.Everything()
	return epGetter.GetEndpoints(namespace, label)
}

func (this *K8sConnector) updateEndpointMap(epAddress string) (string, error) {
	endpoints, err := GetAllEpsInK8s(this.kubeClient)
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

var podIPMap map[string]string = make(map[string]string)

func GetPodNameFromPodIP(PodIP string) string {
	return podIPMap[PodIP]
}
