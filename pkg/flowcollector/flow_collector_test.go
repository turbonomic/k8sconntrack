package flowcollector

import (
	"net"
	"testing"

	"github.com/dongyiyang/k8sconnection/pkg/conntrack"
	"github.com/dongyiyang/k8sconnection/pkg/k8sconnector"
)

type FakeConnInfoBuilder struct {
	MsgType        conntrack.NfConntrackEventType
	Proto          int
	Src            net.IP
	SrcPort        uint16
	Dst            net.IP
	DstPort        uint16
	Packets        uint64
	Bytes          uint64
	StartTimestamp uint64
	DeltaTime      uint64
	TCPState       conntrack.TCPState
}

func NewFakeConnInfoBuilder() *FakeConnInfoBuilder {
	return &FakeConnInfoBuilder{}
}

func (this *FakeConnInfoBuilder) WithMsgType(msgType conntrack.NfConntrackEventType) *FakeConnInfoBuilder {
	this.MsgType = msgType
	return this
}

func (this *FakeConnInfoBuilder) WithProto(p int) *FakeConnInfoBuilder {
	this.Proto = p
	return this
}

func (this *FakeConnInfoBuilder) WithSrc(src net.IP) *FakeConnInfoBuilder {
	this.Src = src
	return this
}

func (this *FakeConnInfoBuilder) WithSrcPort(sPort uint16) *FakeConnInfoBuilder {
	this.SrcPort = sPort
	return this
}

func (this *FakeConnInfoBuilder) WithDst(dst net.IP) *FakeConnInfoBuilder {
	this.Dst = dst
	return this
}

func (this *FakeConnInfoBuilder) WithDstPort(dstPort uint16) *FakeConnInfoBuilder {
	this.DstPort = dstPort
	return this
}

func (this *FakeConnInfoBuilder) WithPackets(p uint64) *FakeConnInfoBuilder {
	this.Packets = p
	return this
}

func (this *FakeConnInfoBuilder) WithBytes(b uint64) *FakeConnInfoBuilder {
	this.Bytes = b
	return this
}

func (this *FakeConnInfoBuilder) WithStartTimestamp(ts uint64) *FakeConnInfoBuilder {
	this.StartTimestamp = ts
	return this
}

func (this *FakeConnInfoBuilder) WithTCPState(state conntrack.TCPState) *FakeConnInfoBuilder {
	this.TCPState = state
	return this
}

func (this *FakeConnInfoBuilder) WithDeltaTime(dtime uint64) *FakeConnInfoBuilder {
	this.DeltaTime = dtime
	return this
}

func (this *FakeConnInfoBuilder) Build() *conntrack.ConntrackInfo {
	return &conntrack.ConntrackInfo{
		MsgType:        this.MsgType,
		Proto:          this.Proto,
		Src:            this.Src,
		SrcPort:        this.SrcPort,
		Dst:            this.Dst,
		DstPort:        this.DstPort,
		Packets:        this.Packets,
		Bytes:          this.Bytes,
		StartTimestamp: this.StartTimestamp,
		DeltaTime:      this.DeltaTime,
		TCPState:       this.TCPState,
	}
}

func TestKeyFunc(t *testing.T) {
	tests := []struct {
		HasData     bool
		SrcIP       string
		SrcPort     uint16
		DstIP       string
		DstPort     uint16
		Timestamp   uint64
		ExpectedKey string
	}{
		{
			HasData:     true,
			SrcIP:       "10.2.3.123",
			SrcPort:     10,
			DstIP:       "183.123.12.2",
			DstPort:     8080,
			Timestamp:   1471017354,
			ExpectedKey: "10.2.3.123:10->183.123.12.2:8080#1471017354",
		},
		{
			HasData:     false,
			ExpectedKey: "",
		},
	}

	for _, test := range tests {
		var connInfo *conntrack.ConntrackInfo
		if test.HasData {
			srcIP := net.ParseIP(test.SrcIP)
			dstIP := net.ParseIP(test.DstIP)
			connInfo = NewFakeConnInfoBuilder().WithSrc(srcIP).WithSrcPort(test.SrcPort).WithDst(dstIP).WithDstPort(test.DstPort).WithStartTimestamp(test.Timestamp).Build()
		}
		key := keyFunc(connInfo)
		if test.ExpectedKey != key {
			t.Errorf("Expect key: %s, get key: %s", test.ExpectedKey, key)
		}
	}
}

func TestFlowConnectionFilterFunc(t *testing.T) {
	tests := []struct {
		MsgType              conntrack.NfConntrackEventType
		SrcIP                net.IP
		IncludeSrcIP         bool
		DstIP                net.IP
		IncludeDstIP         bool
		TCPState             conntrack.TCPState
		ExpectedFilterResult bool
	}{
		{
			MsgType:              conntrack.NfctMsgUpdate,
			SrcIP:                net.ParseIP("10.0.0.3"),
			IncludeSrcIP:         true,
			DstIP:                net.ParseIP("10.0.0.6"),
			IncludeDstIP:         true,
			TCPState:             conntrack.TCPState_ESTABLISHED,
			ExpectedFilterResult: true,
		},
		{
			MsgType:              conntrack.NfctMsgUnknown,
			SrcIP:                net.ParseIP("10.0.0.3"),
			IncludeSrcIP:         true,
			DstIP:                net.ParseIP("10.0.0.6"),
			IncludeDstIP:         true,
			TCPState:             conntrack.TCPState_ESTABLISHED,
			ExpectedFilterResult: false,
		},
		{
			MsgType:              conntrack.NfctMsgUpdate,
			SrcIP:                net.ParseIP("10.0.0.3"),
			IncludeSrcIP:         false,
			DstIP:                net.ParseIP("10.0.0.6"),
			IncludeDstIP:         true,
			TCPState:             conntrack.TCPState_ESTABLISHED,
			ExpectedFilterResult: false,
		},
		{
			MsgType:              conntrack.NfctMsgUpdate,
			SrcIP:                net.ParseIP("10.0.0.3"),
			IncludeSrcIP:         true,
			DstIP:                net.ParseIP("10.0.0.6"),
			IncludeDstIP:         false,
			TCPState:             conntrack.TCPState_ESTABLISHED,
			ExpectedFilterResult: false,
		},
		{
			MsgType:              conntrack.NfctMsgUpdate,
			SrcIP:                net.ParseIP("10.0.0.3"),
			IncludeSrcIP:         true,
			DstIP:                net.ParseIP("10.0.0.6"),
			IncludeDstIP:         false,
			TCPState:             conntrack.TCPState_NONE,
			ExpectedFilterResult: false,
		},
	}

	for _, test := range tests {
		connector := k8sconnector.NewFakeConnector()
		if test.IncludeSrcIP {
			connector.AddPodIP(test.SrcIP.String())
		}
		if test.IncludeDstIP {
			connector.AddPodIP(test.DstIP.String())
		}
		flowCollector := NewFlowCollector(connector)
		connInfo := NewFakeConnInfoBuilder().WithMsgType(test.MsgType).WithSrc(test.SrcIP).WithDst(test.DstIP).WithTCPState(test.TCPState).Build()
		filterResult := flowCollector.flowConnectionFilterFunc(*connInfo)
		if filterResult != test.ExpectedFilterResult {
			t.Errorf("Expected filterFunc result %t, got %t", test.ExpectedFilterResult, filterResult)
		}
	}
}
