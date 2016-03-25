package conntrack

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"syscall"
	"unsafe"

	"github.com/dongyiyang/k8sconnection/conntrack/k8sconnector"

	"github.com/golang/glog"
)

var k8sConnector *k8sconnector.K8sConnector

func SetK8sConnector(connector *k8sconnector.K8sConnector) {
	k8sConnector = connector
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

// Established lists all established TCP connections.
func Established() ([]ConnTCP, error) {
	s, lsa, err := connectNetfilter(0)
	if err != nil {
		return nil, err
	}
	defer syscall.Close(s)

	var conns []ConnTCP
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
	// fmt.Printf("msg bytes: %q\n", wb)
	if err := syscall.Sendto(s, wb, 0, lsa); err != nil {
		return nil, err
	}

	local := findPodIPs()
	readMsgs(s, func(c Conn) {
		if c.MsgType != NfctMsgUpdate {
			glog.V(3).Infof("msg isn't an update: %d\n", c.MsgType)
			return
		}
		if c.TCPState != "ESTABLISHED" {
			// fmt.Printf("state isn't ESTABLISHED: %s\n", c.TCPState)
			return
		}
		if tc := c.ConnTCP(local); len(tc) > 0 {
			for _, conn := range tc {
				conns = append(conns, *conn)
			}
		}
	})
	return conns, nil
}

// Follow gives a channel with all changes.
func Follow() (<-chan Conn, func(), error) {
	s, _, err := connectNetfilter(NF_NETLINK_CONNTRACK_NEW | NF_NETLINK_CONNTRACK_UPDATE | NF_NETLINK_CONNTRACK_DESTROY)
	stop := func() {
		syscall.Close(s)
	}
	if err != nil {
		return nil, stop, err
	}

	res := make(chan Conn, 1)
	go func() {
		defer syscall.Close(s)
		if err := readMsgs(s, func(c Conn) {
			if c.TCPState != "ESTABLISHED" {
				// 3 is TCP established.
				// fmt.Printf("NOT ESTABLISHED. conn is %v \n", c)

				return
			}
			if c.TCPState == "ESTABLISHED" {

				glog.V(4).Infof("conn is %v \n", c)
			}
			res <- c
		}); err != nil {
			panic(err)
		}
	}()
	return res, stop, nil
}

func readMsgs(s int, cb func(Conn)) error {
	for {
		rb := make([]byte, syscall.Getpagesize()) // TODO: re-use
		nr, _, err := syscall.Recvfrom(s, rb, 0)
		if err != nil {
			return err
		}

		msgs, err := syscall.ParseNetlinkMessage(rb[:nr])
		if err != nil {
			return err
		}
		for _, msg := range msgs {
			if err := nfnlIsError(msg.Header); err != nil {
				return fmt.Errorf("msg is some error: %s\n", err)
			}
			if nflnSubsysID(msg.Header.Type) != NFNL_SUBSYS_CTNETLINK {
				return fmt.Errorf(
					"unexpected subsys_id: %d\n",
					nflnSubsysID(msg.Header.Type),
				)
			}
			conn, err := parsePayload(msg.Data[sizeofGenmsg:])
			if err != nil {
				return err
			}
			if conn.Proto != syscall.IPPROTO_TCP {
				continue
			}

			// Taken from conntrack/parse.c:__parse_message_type
			switch CntlMsgTypes(nflnMsgType(msg.Header.Type)) {
			case IpctnlMsgCtNew:
				conn.MsgType = NfctMsgUpdate
				if msg.Header.Flags&(syscall.NLM_F_CREATE|syscall.NLM_F_EXCL) > 0 {
					conn.MsgType = NfctMsgNew
				}
			case IpctnlMsgCtDelete:
				conn.MsgType = NfctMsgDestroy
			}

			cb(*conn)
		}
	}
}

type Conn struct {
	MsgType  NfConntrackMsg
	Proto    int
	Src      net.IP
	SrcPort  uint16
	Dst      net.IP
	DstPort  uint16
	TCPState string
}

// ConnTCP decides which way this connection is going and makes a ConnTCP.
func (c Conn) ConnTCP(addressSet map[string]struct{}) []*ConnTCP {
	var res []*ConnTCP
	// conntrack gives us all connections, even things passing through. But here we only
	// care connection those are sourced from or destinated to address defined in addressSet
	src := c.Src.String()
	dst := c.Dst.String()
	_, srcLocal := addressSet[src]
	_, dstLocal := addressSet[dst]
	// If both are found in addressSet we must just order things predictably.
	if srcLocal && dstLocal {
		srcLocal = c.SrcPort < c.DstPort
	}
	if srcLocal {
		srcConn := &ConnTCP{
			Local:      src,
			LocalPort:  strconv.Itoa(int(c.SrcPort)),
			Remote:     dst,
			RemotePort: strconv.Itoa(int(c.DstPort)),
		}
		res = append(res, srcConn)
	}
	if dstLocal {
		dstConn := &ConnTCP{
			Local:      dst,
			LocalPort:  strconv.Itoa(int(c.DstPort)),
			Remote:     src,
			RemotePort: strconv.Itoa(int(c.SrcPort)),
		}
		res = append(res, dstConn)
	}
	// Neither is in addressSet. conntrack also reports NAT connections.
	return res
}

func parsePayload(b []byte) (*Conn, error) {
	// Most of this comes from libnetfilter_conntrack/src/conntrack/parse_mnl.c
	conn := &Conn{}
	attrs, err := parseAttrs(b)
	if err != nil {
		return conn, err
	}
	for _, attr := range attrs {
		switch CtattrType(attr.Typ) {
		case CtaTupleOrig:
		case CtaTupleReply:
			// fmt.Printf("It's a reply\n")
			// We take the reply, nor the orig.... Sure?
			parseTuple(attr.Msg, conn)
		case CtaStatus:
			// These are ip_conntrack_status
			// status := binary.BigEndian.Uint32(attr.Msg)
			// fmt.Printf("It's status %d\n", status)
		case CtaProtoinfo:
			parseProtoinfo(attr.Msg, conn)
		}
	}
	return conn, nil
}

func parseTuple(b []byte, conn *Conn) error {
	attrs, err := parseAttrs(b)
	if err != nil {
		return fmt.Errorf("invalid tuple attr: %s", err)
	}
	for _, attr := range attrs {
		// fmt.Printf("pl: %d, type: %d, multi: %t, bigend: %t\n", len(attr.Msg), attr.Typ, attr.IsNested, attr.IsNetByteorder)
		switch CtattrTuple(attr.Typ) {
		case CtaTupleUnspec:
			// fmt.Printf("It's a tuple unspec\n")
		case CtaTupleIp:
			// fmt.Printf("It's a tuple IP\n")
			if err := parseIP(attr.Msg, conn); err != nil {
				return err
			}
		case CtaTupleProto:
			// fmt.Printf("It's a tuple proto\n")
			parseProto(attr.Msg, conn)
		}
	}
	return nil
}

func parseIP(b []byte, conn *Conn) error {
	attrs, err := parseAttrs(b)
	if err != nil {
		return fmt.Errorf("invalid tuple attr: %s", err)
	}
	for _, attr := range attrs {
		switch CtattrIp(attr.Typ) {
		case CtaIpV4Src:
			conn.Src = net.IP(attr.Msg) // TODO: copy so we can reuse the buffer?
		case CtaIpV4Dst:
			conn.Dst = net.IP(attr.Msg) // TODO: copy so we can reuse the buffer?
		case CtaIpV6Src:
			// TODO
		case CtaIpV6Dst:
			// TODO
		}
	}
	return nil
}

func parseProto(b []byte, conn *Conn) error {
	attrs, err := parseAttrs(b)
	if err != nil {
		return fmt.Errorf("invalid tuple attr: %s", err)
	}
	for _, attr := range attrs {
		switch CtattrL4proto(attr.Typ) {
		case CtaProtoNum:
			conn.Proto = int(uint8(attr.Msg[0]))
		case CtaProtoSrcPort:
			conn.SrcPort = binary.BigEndian.Uint16(attr.Msg)
		case CtaProtoDstPort:
			conn.DstPort = binary.BigEndian.Uint16(attr.Msg)
		}
	}
	return nil
}

func parseProtoinfo(b []byte, conn *Conn) error {
	attrs, err := parseAttrs(b)
	if err != nil {
		return fmt.Errorf("invalid tuple attr: %s", err)
	}
	for _, attr := range attrs {
		switch CtattrProtoinfo(attr.Typ) {
		case CtaProtoinfoTcp:
			if err := parseProtoinfoTCP(attr.Msg, conn); err != nil {
				return err
			}
		default:
			// we're not interested in other protocols
		}
	}
	return nil
}

func parseProtoinfoTCP(b []byte, conn *Conn) error {
	attrs, err := parseAttrs(b)
	if err != nil {
		return fmt.Errorf("invalid tuple attr: %s", err)
	}
	for _, attr := range attrs {
		switch CtattrProtoinfoTcp(attr.Typ) {
		case CtaProtoinfoTcpState:
			conn.TCPState = tcpState[uint8(attr.Msg[0])]
		default:
			// not interested
		}
	}
	return nil
}
