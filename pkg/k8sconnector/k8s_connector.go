package k8sconnector

import (
	"fmt"
	"reflect"
	"time"

	"k8s.io/kubernetes/pkg/api"
	client "k8s.io/kubernetes/pkg/client/unversioned"

	"github.com/dongyiyang/k8sconnection/pkg/k8sconnector/getter"
	"github.com/dongyiyang/k8sconnection/pkg/k8sconnector/storage"

	"github.com/golang/glog"
)

// A map translating endpoint address to corresponding service.
// We build this so that we do not need to iterate over all the endpoint set every time.
var endpointServiceMap map[string]string = make(map[string]string)

// A K8sConnector is used to connect to the Kubernetes cluster and get info from cluster.
type K8sConnector struct {
	nodeAddress map[string]struct{}

	podIPStore *storage.K8sInfoStore

	// K8sEntityGetter is used to get all Kubernetes entities from Kuberentes cluster.
	k8sEntityGetter *getter.K8sEntityGetter
}

func NewK8sConnector(c *client.Client) (*K8sConnector, error) {
	glog.V(3).Infof("Building Kubernetes connector")

	k8sEntityGetter := getter.NewK8sEntityGetter()

	// Create NodeGetter
	nodeGetter := getter.NewK8sNodeGetter(c)
	k8sEntityGetter.RegisterEntityGetter(getter.EntityType_Node, nodeGetter)
	// Create PodGetter and register to k8sEntityGetter
	podGetter := getter.NewK8sPodGetter(c)
	k8sEntityGetter.RegisterEntityGetter(getter.EntityType_Pod, podGetter)

	// Create EndpointGetter and register to k8sEntityGetter
	endpointGetter := getter.NewK8sEndpointGetter(c)
	k8sEntityGetter.RegisterEntityGetter(getter.EntityType_Endpoint, endpointGetter)

	nodeIPs, err := findIPsOfCurrentNode()
	if err != nil {
		return nil, err
	}

	connector := &K8sConnector{
		nodeAddress: nodeIPs,
		podIPStore:  storage.NewK8sInfoStore(),

		k8sEntityGetter: k8sEntityGetter,
	}

	connector.initEndpointMap()

	go connector.syncK8s()

	return connector, nil
}

// Sync Kubernetes and update info periodically.
// So that we don't have to make API calls to Kubernetes api-server every time
// for some operations, such as getting pods IPs.
func (this *K8sConnector) syncK8s() {
	// TODO: use time.sleep for now...
	for {
		// sync pod ip
		podIPs, err := this.retreivePodsIPsOnNode()
		if err != nil {
			glog.Fatalf("Error getting pod IPs from Kubernetes cluster: %s.\n Exit...", err)
		}
		this.podIPStore.DeleteAll()
		for _, podIP := range podIPs {
			this.podIPStore.Add(podIP, struct{}{})
		}

		// Synchronize every 10 seconds
		time.Sleep(time.Second * 10)
	}
}

// Initialize the endpoint to service map.
// First get all the endpoint from Kubernetes cluster; Then create map.
func (this *K8sConnector) initEndpointMap() error {
	glog.V(3).Infof("Initialize endpoint map.")
	endpoints, err := this.GetAllEpsInK8s()
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

	ipSet := this.FindPodsIP()
	var ipList []string
	for ip, _ := range ipSet {
		ipList = append(ipList, ip)
	}
	return ipList, nil
}

// Get the ip of pods those are running on the current node.
func (this *K8sConnector) FindPodsIP() map[string]interface{} {
	return this.podIPStore.GetAll()
}

// Get the ip of pods those are running on the current node.
func (this *K8sConnector) retreivePodsIPsOnNode() ([]string, error) {
	// glog.V(3).Infof("Now get pod running on current node.")
	entityList, err := this.k8sEntityGetter.GetAllEntitiesOfType(getter.EntityType_Pod)
	if err != nil {
		return nil, err
	}
	podList, ok := entityList.(*getter.PodList)
	if !ok {
		glog.Errorf("Type of entityList is %v, but got %v", reflect.TypeOf(entityList), reflect.TypeOf(podList))
		return nil, fmt.Errorf("Casting Error.")
	}
	ips := extractValidPodsIPs(podList.Items, this.nodeAddress)
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
				podIP2NameMap[pod.Status.PodIP] = pod.Name
			}
		}
	}
	return ips
}

// Retrieve service name based on endpoint address.
// First try to find the service name from the endpoint2service map. If not found,
// trigger an update for all the entries in the map.
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
func (this *K8sConnector) GetAllEpsInK8s() ([]*api.Endpoints, error) {
	entityList, err := this.k8sEntityGetter.GetAllEntitiesOfType(getter.EntityType_Endpoint)
	if err != nil {
		return nil, fmt.Errorf("Cannot get all endpoints from Kubernetes cluster.")
	}
	endpointList, ok := entityList.(*getter.EndpointList)
	if !ok {
		return nil, fmt.Errorf("Casting Error.")
	}
	return endpointList.Items, nil
}

// Find the name of the endpoint according to its address and retrieve its corresponding service, then update endpoint-service map.
func (this *K8sConnector) updateEndpointMap(epAddress string) (string, error) {
	endpoints, err := this.GetAllEpsInK8s()
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

var podIP2NameMap map[string]string = make(map[string]string)

func GetPodNameFromPodIP(PodIP string) string {
	return podIP2NameMap[PodIP]
}
