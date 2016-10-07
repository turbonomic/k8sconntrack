package conntrack

import (
	"fmt"
	"syscall"
	"unsafe"
)

func connectNetfilter(groups uint32) (int, *syscall.SockaddrNetlink, error) {
	s, err := syscall.Socket(syscall.AF_NETLINK, syscall.SOCK_RAW, syscall.NETLINK_NETFILTER)
	if err != nil {
		return 0, nil, err
	}
	lsa := &syscall.SockaddrNetlink{
		Family: syscall.AF_NETLINK,
		Groups: groups,
	}
	if err := syscall.Bind(s, lsa); err != nil {
		return 0, nil, err
	}
	return s, lsa, nil
}

// Read from Netfilter and parse the result into ConntrackInfo object.
// The resulting ConntrackInfo object is then passed into callback for further processing.
func readMessagesFromNetfilter(s int, callback func(ConntrackInfo)) error {
	for {
		rb := make([]byte, syscall.Getpagesize())
		nr, _, err := syscall.Recvfrom(s, rb, 0)
		if err != nil {
			return err
		}

		msgs, err := syscall.ParseNetlinkMessage(rb[:nr])
		if err != nil {
			return fmt.Errorf("Error parsing netlink message: %s", err)
		}
		for _, msg := range msgs {
			if err := nfnlIsError(msg.Header); err != nil {
				return fmt.Errorf("Got an error message: %s\n", err)
			}
			if nfnlSubsysID(msg.Header.Type) != NFNL_SUBSYS_CTNETLINK {
				return fmt.Errorf("Unexpected subsys_id: %d\n",
					nfnlSubsysID(msg.Header.Type))
			}

			// Now we can parse the raw message got from Netfilter.
			conn, err := parsePayload(msg.Data[sizeofGenmsg:])
			if err != nil {
				return err
			}

			if conn.Proto != syscall.IPPROTO_TCP {
				// NOTE: We only process tcp connection right now.
				continue
			}

			// Set connection type: Taken from conntrack/parse.c:__parse_message_type.
			switch CntlMsgTypes(nflnMsgType(msg.Header.Type)) {
			case IpctnlMsgCtNew:
				conn.MsgType = NfctMsgUpdate
				if msg.Header.Flags&(syscall.NLM_F_CREATE|syscall.NLM_F_EXCL) > 0 {
					conn.MsgType = NfctMsgNew
				}
			case IpctnlMsgCtDelete:
				conn.MsgType = NfctMsgDestroy
			}

			callback(*conn)
		}
	}
}

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

func buildConntrackListRequest() []byte {
	// build the request.
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
	return msg.toWireFormat()

}

func sendRequestToNetfilter() (int, error) {
	fd, sa, err := connectNetfilter(0)
	if err != nil {
		return -1, fmt.Errorf("Error connecting Netfilter: %s", err)
	}

	p := buildConntrackListRequest()

	if err := syscall.Sendto(fd, p, 0, sa); err != nil {
		return -1, err
	}
	return fd, nil
}
