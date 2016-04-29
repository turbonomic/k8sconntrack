package conntrack

import (
	"github.com/golang/glog"
)

// find the IPs of pods runnign in current host.
func findPodIPs() map[string]struct{} {
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
