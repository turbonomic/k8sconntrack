package conntrack

import (
	"net"
	"reflect"
	"testing"
)

func TestBuildTCPConnection(t *testing.T) {
	tests := []struct {
		ConnInfo     ConntrackInfo
		AddressSet   map[string]struct{}
		ExpectedSize int
		ExpectedList []*TCPConnection
	}{
		{
			ConnInfo: ConntrackInfo{
				Src:     net.ParseIP("10.1.1.0"),
				SrcPort: 8090,
				Dst:     net.ParseIP("10.4.3.1"),
				DstPort: 4221,
			},
			AddressSet: map[string]struct{}{
				"10.1.1.0": struct{}{},
				"10.4.3.1": struct{}{},
			},
			ExpectedSize: 2,
			ExpectedList: []*TCPConnection{
				&TCPConnection{
					Local:      "10.1.1.0",
					LocalPort:  "8090",
					Remote:     "10.4.3.1",
					RemotePort: "4221",
				},
				&TCPConnection{
					Local:      "10.4.3.1",
					LocalPort:  "4221",
					Remote:     "10.1.1.0",
					RemotePort: "8090",
				},
			},
		},
		{
			ConnInfo: ConntrackInfo{
				Src:     net.ParseIP("10.1.1.0"),
				SrcPort: 8090,
				Dst:     net.ParseIP("10.4.3.1"),
				DstPort: 4221,
			},
			AddressSet: map[string]struct{}{
				"10.1.1.0": struct{}{},
			},
			ExpectedSize: 1,
			ExpectedList: []*TCPConnection{
				&TCPConnection{
					Local:      "10.1.1.0",
					LocalPort:  "8090",
					Remote:     "10.4.3.1",
					RemotePort: "4221",
				},
			},
		},
	}

	for _, test := range tests {
		tcpConns := test.ConnInfo.BuildTCPConn(test.AddressSet)
		if len(tcpConns) != test.ExpectedSize {
			t.Errorf("Expect %d TCPConnection build from ConntrackInfo, but got %d", test.ExpectedSize, len(tcpConns))
		}
		if !reflect.DeepEqual(tcpConns, test.ExpectedList) {
			t.Errorf("Results are not as expected. Expect %++v, got %++v", tcpConns, test.ExpectedList)
		}
	}
}
