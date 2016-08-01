package getter

import (
	"fmt"

	"k8s.io/kubernetes/pkg/api"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/labels"
)

type EndpointList struct {
	Items []*api.Endpoints
}

func (this *EndpointList) IsEntityList() {}

type K8sEndpointGetter struct {
	getterType EntityType
	kubeClient *client.Client
}

func NewK8sEndpointGetter(kubeClient *client.Client) *K8sEndpointGetter {
	return &K8sEndpointGetter{
		getterType: EntityType_Endpoint,
		kubeClient: kubeClient,
	}
}

func (this *K8sEndpointGetter) GetType() EntityType {
	return this.getterType
}

// Get endpoints match specified namesapce and label.
func (this *K8sEndpointGetter) GetEndpoints(namespace string, selector labels.Selector) ([]*api.Endpoints, error) {
	endpointList, err := this.getEndpoints(namespace, selector)
	if err != nil {
		return nil, err
	}
	return endpointList.Items, nil
}

func (this *K8sEndpointGetter) getEndpoints(namespace string, selector labels.Selector) (*EndpointList, error) {

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

	return &EndpointList{
		Items: epItems}, nil
}

// Implement EntityGetter Interface.
func (this *K8sEndpointGetter) GetAllEntities() (EntityList, error) {
	return this.getEndpoints(api.NamespaceAll, labels.Everything())
}

// Register current node getter to K8sEntityGetter.
func (this *K8sEndpointGetter) register(k8sEntityGetter *K8sEntityGetter) {
	k8sEntityGetter.RegisterEntityGetter(this)
}
