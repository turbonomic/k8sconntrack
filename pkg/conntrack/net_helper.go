package conntrack

import (
	"net"

	"github.com/dongyiyang/k8sconnection/pkg/k8sconnector"

	"github.com/golang/glog"
)

var k8sConnector *k8sconnector.K8sConnector

func SetK8sConnector(connector *k8sconnector.K8sConnector) {
	k8sConnector = connector
}

// find the IPs of pods runnign in current host.
func FindPodIPs() map[string]struct{} {
	var l = map[string]struct{}{}

	ips, err := k8sConnector.GetPodsIPsOnNode()
	if err != nil {
		glog.Fatalf("Cannot get pods IPs: %s", err)
	}
	for _, ip := range ips {
		l[ip] = struct{}{}
	}
	return l
}

func localIPs() map[string]struct{} {
	var l = map[string]struct{}{}
	if localNets, err := net.InterfaceAddrs(); err == nil {
		// Not all networks are IP networks.
		for _, localNet := range localNets {
			if net, ok := localNet.(*net.IPNet); ok {
				l[net.IP.String()] = struct{}{}
			}
		}
	}
	return l
}
