package conntrack

import (
	"fmt"
	"syscall"

	"github.com/golang/glog"
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

// Follow returns a channel with all changes.
// NOTE: currently we only return connection is ESTABLISHED state.
func Follow() (<-chan ConntrackInfo, func(), error) {
	s, _, err := connectNetfilter(NF_NETLINK_CONNTRACK_NEW | NF_NETLINK_CONNTRACK_UPDATE | NF_NETLINK_CONNTRACK_DESTROY)
	stop := func() {
		syscall.Close(s)
	}
	if err != nil {
		return nil, stop, err
	}

	res := make(chan ConntrackInfo, 1)
	go func() {
		defer syscall.Close(s)
		if err := readMessagesFromNetfilter(s, func(c ConntrackInfo) {
			if c.TCPState != TCPState_ESTABLISHED {
				// Only track the connection state in ESTABLISHED for now.
				return
			}
			res <- c
		}); err != nil {
			glog.Fatalf("Error reading message from Netfilter: %++v", err)
			panic(err)
		}
	}()
	return res, stop, nil
}

// Read from Netfilter and parse the result into ConntrackInfo object.
// The resulting ConntrackInfo object is then passed into callback for further processing.
func readMessagesFromNetfilter(s int, callback func(ConntrackInfo)) error {
	for {
		rb := make([]byte, syscall.Getpagesize()) // TODO: re-use
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
			if nflnSubsysID(msg.Header.Type) != NFNL_SUBSYS_CTNETLINK {
				return fmt.Errorf("Unexpected subsys_id: %d\n",
					nflnSubsysID(msg.Header.Type))
			}

			// Now we can parse the raw message got from Netfilter.
			conn, err := parsePayload(msg.Data[sizeofGenmsg:])
			if err != nil {
				return err
			}

			if conn.Proto != syscall.IPPROTO_TCP {
				// We only process tcp connection right now.
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
