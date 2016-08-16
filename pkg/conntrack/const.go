package conntrack

const (
	// #defined in libnfnetlink/include/libnfnetlink/linux_nfnetlink.h
	NFNL_SUBSYS_CTNETLINK = 1
	NFNETLINK_V0          = 0

	// #defined in libnfnetlink/include/libnfnetlink/linux_nfnetlink_compat.h
	NF_NETLINK_CONNTRACK_NEW     = 0x00000001
	NF_NETLINK_CONNTRACK_UPDATE  = 0x00000002
	NF_NETLINK_CONNTRACK_DESTROY = 0x00000004

	// #defined in libnfnetlink/include/libnfnetlink/libnfnetlink.h
	NLA_F_NESTED        = uint16(1 << 15)
	NLA_F_NET_BYTEORDER = uint16(1 << 14)
	NLA_TYPE_MASK       = ^(NLA_F_NESTED | NLA_F_NET_BYTEORDER)
)

// Event types.
type NfConntrackEventType int

const (
	NfctMsgUnknown NfConntrackEventType = 0
	NfctMsgNew     NfConntrackEventType = 1 << 0
	NfctMsgUpdate  NfConntrackEventType = 1 << 1
	NfctMsgDestroy NfConntrackEventType = 1 << 2
)

type TCPState uint8

// taken from libnetfilter_conntrack/src/conntrack/snprintf.c
const (
	TCPState_NONE        TCPState = 0
	TCPState_SYN_SENT    TCPState = 1
	TCPState_SYN_RECV    TCPState = 2
	TCPState_ESTABLISHED TCPState = 3
	TCPState_FIN_WAIT    TCPState = 4
	TCPState_CLOSE_WAIT  TCPState = 5
	TCPState_LAST_ACK    TCPState = 6
	TCPState_TIME_WAIT   TCPState = 7
	TCPState_CLOSE       TCPState = 8
	TCPState_LISTEN      TCPState = 9
	TCPState_MAX         TCPState = 10
	TCPState_IGNORE      TCPState = 11
)
