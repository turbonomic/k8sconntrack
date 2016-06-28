package storage

import (
	"sync"
)

// Endpoint keeps track of current ip of Pods running in cluster.
type PodIPStore struct {
	lock sync.RWMutex
	// Key is PodIP
	items map[string]interface{}
}

func NewPodIPStore() *PodIPStore {

	podIPStore := &PodIPStore{
		items: make(map[string]interface{}),
	}
	return podIPStore
}

func (this *PodIPStore) Add(ip string) {
	this.lock.Lock()
	defer this.lock.Unlock()

	this.items[ip] = struct{}{}
}

func (this *PodIPStore) GetPodIPSets() map[string]interface{} {
	this.lock.RLock()
	defer this.lock.RUnlock()

	return this.items
}

func (this *PodIPStore) DeleteAll() {
	this.lock.Lock()
	defer this.lock.Unlock()

	this.items = make(map[string]interface{})
}
