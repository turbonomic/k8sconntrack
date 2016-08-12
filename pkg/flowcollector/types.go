package flowcollector

import (
	"net"
)

// Network flow between two endpoints. Value is in bytes/s
type Flow struct {
	UID                  string `json:"uid,omitempty"`
	Src                  net.IP `json:"source,omitempty"`
	Dst                  net.IP `json:"destination,omitempty"`
	Value                uint64 `json:"value,omitempty"`
	LastUpdatedTimestamp uint64 `json:"timestamp,omitempty"`
}
