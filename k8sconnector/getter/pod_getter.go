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

type K8sPodGetter struct {
	kubeClient *client.Client
}

func NewK8sPodGetter(kubeClient *client.Client) *K8sPodGetter {
	return &K8sPodGetter{
		kubeClient: kubeClient,
	}
}

// Get pods match specified namesapce, label and field.
func (this *K8sPodGetter) GetPods(namespace string, label labels.Selector, field fields.Selector) ([]*api.Pod, error) {
	listOption := &api.ListOptions{
		LabelSelector: label,
		FieldSelector: field,
	}
	podList, err := this.kubeClient.Pods(namespace).List(*listOption)
	if err != nil {
		return nil, fmt.Errorf("Error getting all the desired pods from Kubernetes cluster: %s", err)
	}
	var podItems []*api.Pod
	for _, pod := range podList.Items {
		p := pod
		podItems = append(podItems, &p)
	}
	glog.V(3).Infof("Discovering Pods, now the cluster has " + strconv.Itoa(len(podItems)) + " pods")

	return podItems, nil
}
