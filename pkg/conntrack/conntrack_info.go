package conntrack

import (
	"fmt"
	"net"
	//	"strconv"
)

// Struct for storing retreived conntrack information. The fields are chosen according the output of nf_conntrack.
type ConntrackInfo struct {
	MsgType        NfConntrackEventType
	Proto          int
	Src            net.IP
	SrcPort        uint16
	Dst            net.IP
	DstPort        uint16
	Packets        uint64
	Bytes          uint64
	StartTimestamp uint64
	DeltaTime      uint64
	TCPState       TCPState
}

func (c ConntrackInfo) String() string {
	return fmt.Sprintf("%s:%d->%s:%d, packets=%d, bytes=%d, start_time=%d, delta_time=%d",
		c.Src, c.SrcPort, c.Dst, c.DstPort, c.Packets, c.Bytes, c.StartTimestamp, c.DeltaTime)
}
