package k8sconnector

import (
	"fmt"
)

type Connector interface {
	FindPodsIP() map[string]interface{}
	GetServiceNameWithEndpointAddress(address string) (string, error)
}

type FakeConnector struct {
	PodIPs map[string]interface{}

	// A map storing endpoint to service info.
	epSvcMap map[string]string
}

func NewFakeConnector() *FakeConnector {
	return &FakeConnector{
		PodIPs:   make(map[string]interface{}),
		epSvcMap: make(map[string]string),
	}
}

func (this *FakeConnector) AddPodIP(ip string) {
	this.PodIPs[ip] = struct{}{}
}

func (this *FakeConnector) FindPodsIP() map[string]interface{} {
	return this.PodIPs
}

func (this *FakeConnector) AddEndpointServiceMapping(ep, svc string) {
	this.epSvcMap[ep] = svc
}

func (this *FakeConnector) GetServiceNameWithEndpointAddress(address string) (string, error) {
	svc, exist := this.epSvcMap[address]
	if !exist {
		return "", fmt.Errorf("Endpoint address %s does not link to any service.", address)
	}
	return svc, nil
}
