package getter

import (
	"k8s.io/kubernetes/pkg/api"
	"testing"
)

type FakeNodeGetter struct {
	nodeList *NodeList
}

func NewFakeNodeGetter(nodes []*api.Node) *FakeNodeGetter {
	return &FakeNodeGetter{
		nodeList: &NodeList{
			Items: nodes,
		},
	}
}

func (this *FakeNodeGetter) GetType() EntityType {
	return EntityType_Node
}

func (this *FakeNodeGetter) GetAllEntities() (EntityList, error) {
	return this.nodeList, nil
}

func makeNodeList(nodeNames []string) []*api.Node {
	result := make([]*api.Node, len(nodeNames))
	for ix := range nodeNames {
		result[ix].Name = nodeNames[ix]
	}
	return result
}

type FakePodGetter struct {
	podList *PodList
}

func NewFakePodGetter(pods []*api.Pod) *FakePodGetter {
	return &FakePodGetter{
		podList: &PodList{
			Items: pods,
		},
	}
}

func (this *FakePodGetter) GetType() EntityType {
	return EntityType_Pod
}

func (this *FakePodGetter) GetAllEntities() (EntityList, error) {
	return this.podList, nil
}

func TestRegisterEntityGetter(t *testing.T) {
	k8sEntityGetter := NewK8sEntityGetter()
	tests := []struct {
		getterType EntityType
		getter     EntityGetter
		register   bool
		found      bool
	}{
		{
			getterType: EntityType_Node,
			getter:     NewFakeNodeGetter(nil),
			register:   true,
			found:      true,
		},
		{
			getterType: EntityType_Pod,
			getter:     NewFakePodGetter(nil),
			register:   false,
			found:      false,
		},
	}
	for _, test := range tests {
		if test.register {
			k8sEntityGetter.RegisterEntityGetter(test.getter)
		}
		_, registered := k8sEntityGetter.getters[test.getterType]
		if test.found != registered {
			t.Errorf("Registered: %t; Found in getters: %t", registered, test.found)
		}
	}
}

func TestRetrieveGetterOfType(t *testing.T) {
	k8sEntityGetter := NewK8sEntityGetter()
	k8sEntityGetter.RegisterEntityGetter(NewFakeNodeGetter(nil))
	tests := []struct {
		expectedEntityType EntityType
		expectsError       bool
	}{
		{
			expectedEntityType: EntityType_Node,
			expectsError:       false,
		},
		{
			expectedEntityType: EntityType_Pod,
			expectsError:       true,
		},
	}
	for _, test := range tests {
		got, err := k8sEntityGetter.RetrieveGetterOfType(test.expectedEntityType)
		if test.expectsError {
			if err == nil {
				t.Error("Unexpected non-error")
			}
		} else {
			if err != nil {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
			if got.GetType() != test.expectedEntityType {
				t.Errorf("Expect type: %v; Got type: %v", test.expectedEntityType, got.GetType())
			}
		}
	}
}
