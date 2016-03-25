package getter

import (
	"fmt"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client"
	"k8s.io/kubernetes/pkg/labels"

	// "github.com/golang/glog"
)

// Pods Getter is such func that gets all the pods match the provided namespace, labels and fiels.
type ServiceGetter func(namespace string, selector labels.Selector) ([]*api.Service, error)

type EndpointGetter func(namespace string, selector labels.Selector) ([]*api.Endpoints, error)

type ServiceProbe struct {
	serviceGetter  ServiceGetter
	endpointGetter EndpointGetter
}

type K8sServiceGetter struct {
	kubeClient *client.Client
}

func NewK8sServiceGetter(kubeClient *client.Client) *K8sServiceGetter {
	return &K8sServiceGetter{
		kubeClient: kubeClient,
	}
}

// Get service match specified namesapce and label.
func (this *K8sServiceGetter) GetService(namespace string, selector labels.Selector) ([]*api.Service, error) {
	serviceList, err := this.kubeClient.Services(namespace).List(selector)
	if err != nil {
		return nil, fmt.Errorf("Error listing services: %s", err)
	}

	var serviceItems []*api.Service
	for _, service := range serviceList.Items {
		s := service
		serviceItems = append(serviceItems, &s)
	}

	return serviceItems, nil
}

type K8sEndpointGetter struct {
	kubeClient *client.Client
}

func NewK8sEndpointGetter(kubeClient *client.Client) *K8sEndpointGetter {
	return &K8sEndpointGetter{
		kubeClient: kubeClient,
	}
}

// Get endpoints match specified namesapce and label.
func (this *K8sEndpointGetter) GetEndpoints(namespace string, selector labels.Selector) ([]*api.Endpoints, error) {
	epList, err := this.kubeClient.Endpoints(namespace).List(selector)
	if err != nil {
		return nil, fmt.Errorf("Error listing endpoints: %s", err)
	}

	var epItems []*api.Endpoints
	for _, endpoint := range epList.Items {
		ep := endpoint
		epItems = append(epItems, &ep)
	}

	return epItems, nil
}
