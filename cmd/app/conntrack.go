package app

import (
	"k8s.io/kubernetes/pkg/util/sysctl"

	"github.com/golang/glog"
)

type Conntracker interface {
	EnableAcct() error
	EnableTimestamp() error
}

type realConntracker struct{}

func (realConntracker) EnableAcct() error {
	glog.Infof("Enabling nf_conntrack_acct.")
	return sysctl.New().SetSysctl("net/netfilter/nf_conntrack_acct", 1)
}

func (realConntracker) EnableTimestamp() error {
	glog.Infof("Enabling nf_conntrack_timestamp.")
	return sysctl.New().SetSysctl("net/netfilter/nf_conntrack_timestamp", 1)
}

func (realConntracker) SetSocketReadBuffer(desiredBufferSize int) error {
	glog.Infof("Changing rmem_default to %d", desiredBufferSize)
	return sysctl.New().SetSysctl("net/core/rmem_default", desiredBufferSize)
}
