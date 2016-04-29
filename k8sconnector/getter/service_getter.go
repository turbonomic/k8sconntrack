package getter

import (
	"fmt"

	"k8s.io/kubernetes/pkg/api"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/labels"
)

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
	listOption := &api.ListOptions{
		LabelSelector: selector,
	}
	serviceList, err := this.kubeClient.Services(namespace).List(*listOption)
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
	listOption := &api.ListOptions{
		LabelSelector: selector,
	}
	epList, err := this.kubeClient.Endpoints(namespace).List(*listOption)
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
