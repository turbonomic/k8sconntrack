package conntrack

// Netlink attr parsing.

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"errors"
)

func parsePayload(b []byte) (*ConntrackInfo, error) {
	// Most of this comes from libnetfilter_conntrack/src/conntrack/parse_mnl.c
	conn := &ConntrackInfo{}
	// glog.V(4).Infof("Before parse Attrs %++v", b)
	attrs, err := parseAttrs(b)
	if err != nil {
		return conn, err
	}
	for _, attr := range attrs {
		// glog.V(4).Infof("Parse each Attrs %++v", attr)

		switch CtattrType(attr.Typ) {
		case CtaTupleOrig: //1
		case CtaTupleReply: //2
			// fmt.Printf("It's a reply\n")
			parseTuple(attr.Msg, conn)
		case CtaStatus: //3
			// These are ip_conntrack_status
			// status := binary.BigEndian.Uint32(attr.Msg)
			// fmt.Printf("It's status %d\n", status)
		case CtaProtoinfo: //4
			parseProtoinfo(attr.Msg, conn)
		case CtaCountersOrig: // 9
			// parseCounters(attr.Msg, conn)
		case CtaCountersReply: //10
			parseCounters(attr.Msg, conn)
		case CtaTimestamp: // 20
			parseTimestamp(attr.Msg, conn)
		}
	}
	return conn, nil
}

func parseTuple(b []byte, conn *ConntrackInfo) error {
	attrs, err := parseAttrs(b)
	if err != nil {
		return fmt.Errorf("invalid tuple attr: %s", err)
	}
	for _, attr := range attrs {
		// fmt.Printf("pl: %d, type: %d, multi: %t, bigend: %t\n", len(attr.Msg), attr.Typ, attr.IsNested, attr.IsNetByteorder)
		switch CtattrTuple(attr.Typ) {
		case CtaTupleUnspec: //0
			// fmt.Printf("It's a tuple unspec\n")
		case CtaTupleIp: //1
			// fmt.Printf("It's a tuple IP\n")
			if err := parseIP(attr.Msg, conn); err != nil {
				return err
			}
		case CtaTupleProto: //2
			// fmt.Printf("It's a tuple proto\n")
			parseProto(attr.Msg, conn)
		}
	}
	return nil
}

func parseIP(b []byte, conn *ConntrackInfo) error {
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

func parseProto(b []byte, conn *ConntrackInfo) error {
	attrs, err := parseAttrs(b)
	if err != nil {
		return fmt.Errorf("invalid tuple attr: %s", err)
	}
	for _, attr := range attrs {
		switch CtattrL4proto(attr.Typ) {
		case CtaProtoNum: //0
			conn.Proto = int(uint8(attr.Msg[0]))
		case CtaProtoSrcPort: //1
			conn.SrcPort = binary.BigEndian.Uint16(attr.Msg)
		case CtaProtoDstPort: //2
			conn.DstPort = binary.BigEndian.Uint16(attr.Msg)
		}
	}
	return nil
}

func parseProtoinfo(b []byte, conn *ConntrackInfo) error {
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

func parseProtoinfoTCP(b []byte, conn *ConntrackInfo) error {
	attrs, err := parseAttrs(b)
	if err != nil {
		return fmt.Errorf("invalid tuple attr: %s", err)
	}
	for _, attr := range attrs {
		switch CtattrProtoinfoTcp(attr.Typ) {
		case CtaProtoinfoTcpState: //1
			conn.TCPState = TCPState(uint8(attr.Msg[0]))
		default:
			// not interested
		}
	}
	return nil
}

func parseCounters(b []byte, conn *ConntrackInfo) error {
	attrs, err := parseAttrs(b)
	if err != nil {
		return fmt.Errorf("invalid counters attr: %s", err)
	}
	for _, attr := range attrs {
		// fmt.Printf("pl: %d, type: %d, multi: %t, bigend: %t\n", len(attr.Msg), attr.Typ, attr.IsNested, attr.IsNetByteorder)
		switch CtattrCounters(attr.Typ) {
		case CtaCountersPackets: //1
			packet := binary.BigEndian.Uint64(attr.Msg)
			// glog.V(4).Infof("packets = %d", packet)
			conn.Packets = packet
		case CtaCountersBytes: //2
			bytes := binary.BigEndian.Uint64(attr.Msg)
			// glog.V(4).Infof("bytes = %d", bytes)
			conn.Bytes = bytes
		}
	}
	return nil
}

func parseTimestamp(b []byte, conn *ConntrackInfo) error {
	attrs, err := parseAttrs(b)
	if err != nil {
		return fmt.Errorf("invalid timestamp attr: %s", err)
	}
	for _, attr := range attrs {
		// fmt.Printf("pl: %d, type: %d, multi: %t, bigend: %t\n", len(attr.Msg), attr.Typ, attr.IsNested, attr.IsNetByteorder)
		switch CtattrTimestamp(attr.Typ) {
		case CtaTimestampStart: //1
			startTime := binary.BigEndian.Uint64(attr.Msg)
			// startTime returned here is in nanoseconds; convert it to seconds
			conn.StartTimestamp = startTime / 1E9
			// fmt.Println(time.Now().Unix())
			conn.DeltaTime = uint64(time.Now().Unix()) - conn.StartTimestamp
		case CtaTimestampStop: //2
			// Don't care stop time right now.
		}
	}
	return nil
}

const attrHdrLength = 4

type Attr struct {
	Msg            []byte
	Typ            int
	IsNested       bool
	IsNetByteorder bool
}

func parseAttrs(b []byte) ([]Attr, error) {
	var attrs []Attr
	for len(b) >= attrHdrLength {
		var attr Attr
		attr, b = parseAttr(b)
		attrs = append(attrs, attr)
	}
	if len(b) != 0 {
		return nil, errors.New("leftover attr bytes")
	}
	return attrs, nil
}

func parseAttr(b []byte) (Attr, []byte) {
	l := binary.LittleEndian.Uint16(b[0:2])
	// length is header + payload
	l -= uint16(attrHdrLength)

	typ := binary.LittleEndian.Uint16(b[2:4])
	attr := Attr{
		Msg:            b[attrHdrLength : attrHdrLength+int(l)],
		Typ:            int(typ & NLA_TYPE_MASK),
		IsNested:       typ&NLA_F_NESTED > 0,
		IsNetByteorder: typ&NLA_F_NET_BYTEORDER > 0,
	}
	// glog.V(4).Infof("Attr is %++v", attr)
	return attr, b[rtaAlignOf(attrHdrLength+int(l)):]
}
