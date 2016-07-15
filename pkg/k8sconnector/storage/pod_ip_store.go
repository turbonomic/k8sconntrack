package storage

import (
	"sync"
)

// Endpoint keeps track of current ip of Pods running in cluster.
type K8sInfoStore struct {
	lock  sync.RWMutex
	items map[string]interface{}
}

func NewK8sInfoStore() *K8sInfoStore {

	k8sInfoStore := &K8sInfoStore{
		items: make(map[string]interface{}),
	}
	return k8sInfoStore
}

func (this *K8sInfoStore) Add(key string, val interface{}) {
	this.lock.Lock()
	defer this.lock.Unlock()

	this.items[key] = val
}

func (this *K8sInfoStore) Get(key string) interface{} {
	this.lock.RLock()
	defer this.lock.RUnlock()

	return this.items[key]
}

func (this *K8sInfoStore) GetAll() map[string]interface{} {
	this.lock.RLock()
	defer this.lock.RUnlock()

	return this.items
}

func (this *K8sInfoStore) DeleteAll() {
	this.lock.Lock()
	defer this.lock.Unlock()

	this.items = make(map[string]interface{})
}
