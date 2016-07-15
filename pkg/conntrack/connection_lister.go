package conntrack

import (
	"fmt"
	"syscall"
	"unsafe"
)

type nfgenmsg struct {
	Family  uint8  /* AF_xxx */
	Version uint8  /* nfnetlink version */
	ResID   uint16 /* resource id */
}

const (
	sizeofGenmsg = uint32(unsafe.Sizeof(nfgenmsg{})) // TODO
)

type ConntrackListReq struct {
	Header syscall.NlMsghdr
	Body   nfgenmsg
}

func (c *ConntrackListReq) toWireFormat() []byte {
	// adapted from syscall/NetlinkRouteRequest.toWireFormat
	b := make([]byte, c.Header.Len)
	*(*uint32)(unsafe.Pointer(&b[0:4][0])) = c.Header.Len
	*(*uint16)(unsafe.Pointer(&b[4:6][0])) = c.Header.Type
	*(*uint16)(unsafe.Pointer(&b[6:8][0])) = c.Header.Flags
	*(*uint32)(unsafe.Pointer(&b[8:12][0])) = c.Header.Seq
	*(*uint32)(unsafe.Pointer(&b[12:16][0])) = c.Header.Pid
	b[16] = byte(c.Body.Family)
	b[17] = byte(c.Body.Version)
	*(*uint16)(unsafe.Pointer(&b[18:20][0])) = c.Body.ResID
	return b
}

type FilterFunc func(c ConntrackInfo) bool

// lists all established TCP connections.
func ListConnections(filter FilterFunc) ([]TCPConnection, error) {
	s, lsa, err := connectNetfilter(0)
	if err != nil {
		return nil, fmt.Errorf("Error listing connections: %s", err)
	}
	defer syscall.Close(s)

	// build the requist.
	msg := ConntrackListReq{
		Header: syscall.NlMsghdr{
			Len:   syscall.NLMSG_HDRLEN + sizeofGenmsg,
			Type:  (NFNL_SUBSYS_CTNETLINK << 8) | uint16(IpctnlMsgCtGet),
			Flags: syscall.NLM_F_REQUEST | syscall.NLM_F_DUMP,
			Pid:   0,
			Seq:   0,
		},
		Body: nfgenmsg{
			Family:  syscall.AF_INET,
			Version: NFNETLINK_V0,
			ResID:   0,
		},
	}
	wb := msg.toWireFormat()
	if err := syscall.Sendto(s, wb, 0, lsa); err != nil {
		return nil, err
	}

	var conns []TCPConnection
	local := FindPodIPs()

	readMessagesFromNetfilter(s, func(c ConntrackInfo) {
		pass := filter(c)
		if !pass {
			return
		}

		if tc := c.BuildTCPConn(local); len(tc) > 0 {
			for _, conn := range tc {
				conns = append(conns, *conn)
			}
		}
	})
	return conns, nil
}

func ListConntrackInfos(filter FilterFunc) ([]ConntrackInfo, error) {
	s, lsa, err := connectNetfilter(0)
	if err != nil {
		return nil, fmt.Errorf("Error listing connections: %s", err)
	}
	defer syscall.Close(s)

	// build the requist.
	msg := ConntrackListReq{
		Header: syscall.NlMsghdr{
			Len:   syscall.NLMSG_HDRLEN + sizeofGenmsg,
			Type:  (NFNL_SUBSYS_CTNETLINK << 8) | uint16(IpctnlMsgCtGet),
			Flags: syscall.NLM_F_REQUEST | syscall.NLM_F_DUMP,
			Pid:   0,
			Seq:   0,
		},
		Body: nfgenmsg{
			Family:  syscall.AF_INET,
			Version: NFNETLINK_V0,
			ResID:   0,
		},
	}
	wb := msg.toWireFormat()
	if err := syscall.Sendto(s, wb, 0, lsa); err != nil {
		return nil, err
	}

	var conns []ConntrackInfo

	readMessagesFromNetfilter(s, func(c ConntrackInfo) {
		if pass := filter(c); pass {
			conns = append(conns, c)
		}

	})
	return conns, nil
}
