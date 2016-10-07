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

type CtattrType int

const (
	CtaUnspec         CtattrType = 0
	CtaTupleOrig      CtattrType = 1
	CtaTupleReply     CtattrType = 2
	CtaStatus         CtattrType = 3
	CtaProtoinfo      CtattrType = 4
	CtaHelp           CtattrType = 5
	CtaNatSrc         CtattrType = 6
	CtaTimeout        CtattrType = 7
	CtaMark           CtattrType = 8
	CtaCountersOrig   CtattrType = 9
	CtaCountersReply  CtattrType = 10
	CtaUse            CtattrType = 11
	CtaId             CtattrType = 12
	CtaNatDst         CtattrType = 13
	CtaTupleMaster    CtattrType = 14
	CtaNatSeqAdjOrig  CtattrType = 15
	CtaNatSeqAdjReply CtattrType = 16
	CtaSecmark        CtattrType = 17
	CtaZone           CtattrType = 18
	CtaSecctx         CtattrType = 19
	CtaTimestamp      CtattrType = 20
	CtaMarkMask       CtattrType = 21
	CtaLabels         CtattrType = 22
	CtaLabelsMask     CtattrType = 23
	CtaMax            CtattrType = 24
)

type CtattrTuple int

const (
	CtaTupleUnspec CtattrTuple = 0
	CtaTupleIp     CtattrTuple = 1
	CtaTupleProto  CtattrTuple = 2
	CtaTupleMax    CtattrTuple = 3
)

type CtattrIp int

const (
	CtaIpUnspec CtattrIp = 0
	CtaIpV4Src  CtattrIp = 1
	CtaIpV4Dst  CtattrIp = 2
	CtaIpV6Src  CtattrIp = 3
	CtaIpV6Dst  CtattrIp = 4
	CtaIpMax    CtattrIp = 5
)

type CtattrL4proto int

const (
	CtaProtoUnspec     CtattrL4proto = 0
	CtaProtoNum        CtattrL4proto = 1
	CtaProtoSrcPort    CtattrL4proto = 2
	CtaProtoDstPort    CtattrL4proto = 3
	CtaProtoIcmpId     CtattrL4proto = 4
	CtaProtoIcmpType   CtattrL4proto = 5
	CtaProtoIcmpCode   CtattrL4proto = 6
	CtaProtoIcmpv6Id   CtattrL4proto = 7
	CtaProtoIcmpv6Type CtattrL4proto = 8
	CtaProtoIcmpv6Code CtattrL4proto = 9
	CtaProtoMax        CtattrL4proto = 10
)

type CtattrProtoinfo int

const (
	CtaProtoinfoUnspec CtattrProtoinfo = 0
	CtaProtoinfoTcp    CtattrProtoinfo = 1
	CtaProtoinfoDccp   CtattrProtoinfo = 2
	CtaProtoinfoSctp   CtattrProtoinfo = 3
	CtaProtoinfoMax    CtattrProtoinfo = 4
)

type CtattrProtoinfoTcp int

const (
	CtaProtoinfoTcpUnspec         CtattrProtoinfoTcp = 0
	CtaProtoinfoTcpState          CtattrProtoinfoTcp = 1
	CtaProtoinfoTcpWscaleOriginal CtattrProtoinfoTcp = 2
	CtaProtoinfoTcpWscaleReply    CtattrProtoinfoTcp = 3
	CtaProtoinfoTcpFlagsOriginal  CtattrProtoinfoTcp = 4
	CtaProtoinfoTcpFlagsReply     CtattrProtoinfoTcp = 5
	CtaProtoinfoTcpMax            CtattrProtoinfoTcp = 6
)

type CtattrCounters int

const (
	CtaCountersUnspec    CtattrCounters = 0
	CtaCountersPackets   CtattrCounters = 1
	CtaCountersBytes     CtattrCounters = 2
	CtaCounters32Packets CtattrCounters = 3
	CtaCoutners32Bytes   CtattrCounters = 4
	CtaCountersMax       CtattrCounters = 5
)

type CtattrTimestamp int

const (
	CtaTimestampUnspec CtattrTimestamp = 0
	CtaTimestampStart  CtattrTimestamp = 1
	CtaTimestampStop   CtattrTimestamp = 2
	CtaTimestampMax    CtattrTimestamp = 3
)

type NfConntrackAttrGrp int

const (
	AttrGrpOrigIpv4     NfConntrackAttrGrp = 0
	AttrGrpReplIpv4     NfConntrackAttrGrp = 1
	AttrGrpOrigIpv6     NfConntrackAttrGrp = 2
	AttrGrpReplIpv6     NfConntrackAttrGrp = 3
	AttrGrpOrigPort     NfConntrackAttrGrp = 4
	AttrGrpReplPort     NfConntrackAttrGrp = 5
	AttrGrpIcmp         NfConntrackAttrGrp = 6
	AttrGrpMasterIpv4   NfConntrackAttrGrp = 7
	AttrGrpMasterIpv6   NfConntrackAttrGrp = 8
	AttrGrpMasterPort   NfConntrackAttrGrp = 9
	AttrGrpOrigCounters NfConntrackAttrGrp = 10
	AttrGrpReplCounters NfConntrackAttrGrp = 11
	AttrGrpOrigAddrSrc  NfConntrackAttrGrp = 12
	AttrGrpOrigAddrDst  NfConntrackAttrGrp = 13
	AttrGrpReplAddrSrc  NfConntrackAttrGrp = 14
	AttrGrpReplAddrDst  NfConntrackAttrGrp = 15
	AttrGrpMax          NfConntrackAttrGrp = 16
)

type NfConntrackQuery int

const (
	NfctQCreate          NfConntrackQuery = 0
	NfctQUpdate          NfConntrackQuery = 1
	NfctQDestroy         NfConntrackQuery = 2
	NfctQGet             NfConntrackQuery = 3
	NfctQFlush           NfConntrackQuery = 4
	NfctQDump            NfConntrackQuery = 5
	NfctQDumpReset       NfConntrackQuery = 6
	NfctQCreateUpdate    NfConntrackQuery = 7
	NfctQDumpFilter      NfConntrackQuery = 8
	NfctQDumpFilterReset NfConntrackQuery = 9
)

type CntlMsgTypes int

const (
	IpctnlMsgCtNew            CntlMsgTypes = 0
	IpctnlMsgCtGet            CntlMsgTypes = 1
	IpctnlMsgCtDelete         CntlMsgTypes = 2
	IpctnlMsgCtGetCtrzero     CntlMsgTypes = 3
	IpctnlMsgCtGetStatsCpu    CntlMsgTypes = 4
	IpctnlMsgCtGetStats       CntlMsgTypes = 5
	IpctnlMsgCtGetDying       CntlMsgTypes = 6
	IpctnlMsgCtGetUnconfirmed CntlMsgTypes = 7
	IpctnlMsgMax              CntlMsgTypes = 8
)
