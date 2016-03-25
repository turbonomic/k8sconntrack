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

// Pods Getter is such func that gets all the pods match the provided namespace, labels and fiels.
type PodsGetter func(namespace string, label labels.Selector, field fields.Selector) ([]*api.Pod, error)

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
	podList, err := this.kubeClient.Pods(namespace).List(label, field)
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
