package util

import (
	"fmt"
	"net"

	"github.com/golang/glog"
)

// Find the all valid IP address of the node where k8sConntrack runs.
func findIPsOfCurrentNode() (map[string]struct{}, error) {
	var l = map[string]struct{}{}
	if localNets, err := net.InterfaceAddrs(); err == nil {
		// Not all networks are IP networks.
		for _, localNet := range localNets {
			if network, ok := localNet.(*net.IPNet); ok {
				netAddress := network.IP.String()
				glog.Infof("Find network address %s", netAddress)

				ipAddress := net.ParseIP(netAddress)
				if ipAddress.To4() == nil {
					// Only support IPv4 now
					continue
				}
				glog.Infof("Find valid IPv4 address %s", netAddress)

				l[netAddress] = struct{}{}
			}
		}
		return l, nil
	} else {
		glog.Errorf("Error is %s", err)
		return nil, fmt.Errorf("Error finding IP address of current node: %s", err)
	}
}
