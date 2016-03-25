package getter

import (
	"fmt"
	"strconv"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"

	"github.com/golang/glog"
)

type NodesGetter func(label labels.Selector, field fields.Selector) ([]*api.Node, error)

type K8sNodeGetter struct {
	kubeClient *client.Client
}

func NewK8sNodeGetter(kubeClient *client.Client) *K8sNodeGetter {
	return &K8sNodeGetter{
		kubeClient: kubeClient,
	}
}

// Get all nodes
func (this *K8sNodeGetter) GetNodes(label labels.Selector, field fields.Selector) ([]*api.Node, error) {
	nodeList, err := this.kubeClient.Nodes().List(label, field)
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
