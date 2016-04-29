package getter

import (
	"fmt"
	"strconv"

	"k8s.io/kubernetes/pkg/api"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"

	"github.com/golang/glog"
)

type K8sNodeGetter struct {
	kubeClient *client.Client
}

func NewK8sNodeGetter(kubeClient *client.Client) *K8sNodeGetter {
	return &K8sNodeGetter{
		kubeClient: kubeClient,
	}
}

func (this *K8sNodeGetter) GetAllNodes() ([]*api.Node, error) {
	return this.GetNodes(labels.Everything(), fields.Everything())
}

// Get all nodes
func (this *K8sNodeGetter) GetNodes(label labels.Selector, field fields.Selector) ([]*api.Node, error) {
	listOption := &api.ListOptions{
		LabelSelector: label,
		FieldSelector: field,
	}
	nodeList, err := this.kubeClient.Nodes().List(*listOption)
	if err != nil {
		//TODO error must be handled
		return nil, fmt.Errorf("Error getting nodes from Kubernetes cluster: %s", err)
	}
	var nodeItems []*api.Node
	for _, node := range nodeList.Items {
		n := node
		nodeItems = append(nodeItems, &n)
	}
	glog.V(3).Infof("Discovering Nodes.. The cluster has " + strconv.Itoa(len(nodeItems)) + " nodes")
	return nodeItems, nil
}
