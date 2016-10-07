package conntrack

import (
	"fmt"
	//	"time"
	"syscall"

	"github.com/golang/glog"
)

// FilterFunc is used against each ConntrackInfo. If pass return true; otherwise return false.
type FilterFunc func(c ConntrackInfo) bool

func DefaultFilter(c ConntrackInfo) bool {
	// Here we only care about updated info
	if c.MsgType != NfctMsgUpdate {
		glog.V(4).Infof("Message isn't an update: %d\n", c.MsgType)
		return false
	}
	// As for updated info, we only care about ESTABLISHED for now.
	if c.TCPState != TCPState_ESTABLISHED {
		glog.V(4).Infof("State isn't in ESTABLISHED: %s\n", c.TCPState)
		return false
	}
	glog.V(4).Infof("Established TCP connection is %++v \n", c)
	return true
}

// ConnTrack monitors the network connections.
type ConnTrack struct {
	connReq chan chan []ConntrackInfo
	quit    chan struct{}

	filterFunc FilterFunc
}

// New returns a ConnTrack.
func New(filterFunc FilterFunc) (*ConnTrack, error) {
	c := ConnTrack{
		connReq: make(chan chan []ConntrackInfo),
		quit:    make(chan struct{}),

		filterFunc: filterFunc,
	}
	go func() {
		//		for {
		err := c.track()
		select {
		case <-c.quit:
			return
		default:
		}
		if err != nil {
			glog.Errorf("conntrack: %s\n", err)
		}
		//			time.Sleep(1 * time.Second)
		//		}
	}()

	return &c, nil
}

// Close stops all monitoring and executables.
func (c ConnTrack) Close() {
	close(c.quit)
}

// track is the main loop
func (c *ConnTrack) track() error {
	// We use Follow() to keep track of conn state changes, but it doesn't give
	// us the initial state.
	events, stop, err := c.Follow()
	if err != nil {
		return err
	}

	// Use ListTCPConnections to get current established tcp connections.
	establishedConns, err := c.ListConntrackInfos()
	if err != nil {
		return fmt.Errorf("Error listing existing ESTABLISHED connections: %++v.", err)
	}

	established := map[string]ConntrackInfo{}
	for _, c := range establishedConns {
		established[c.String()] = c
	}

	for {
		select {

		case <-c.quit:
			stop()
			return nil

		case e, ok := <-events:
			if !ok {
				return nil
			}
			switch {

			default:
				// not interested

			case e.TCPState == TCPState_ESTABLISHED:
				established[e.String()] = e
				glog.V(4).Infof("track() - Established Connection payload is %++v", e)
			}

		case r := <-c.connReq:
			cs := make([]ConntrackInfo, 0, len(established))
			for _, c := range established {
				cs = append(cs, c)
			}
			r <- cs
			established = map[string]ConntrackInfo{}
		}
	}
}

func (c *ConnTrack) ListConntrackInfos() ([]ConntrackInfo, error) {
	s, err := sendRequestToNetfilter()
	defer syscall.Close(s)

	if err != nil {
		return nil, err
	}
	var conns []ConntrackInfo

	readMessagesFromNetfilter(s, func(conntrackInfo ConntrackInfo) {
		if pass := c.filterFunc(conntrackInfo); pass {
			conns = append(conns, conntrackInfo)
		}

	})
	return conns, nil
}

// Connections gets the list of all connection track events seen since last time you
// called it and return them as a list of ConntrackInfo.
func (c *ConnTrack) ConnectionEvents() []ConntrackInfo {
	r := make(chan []ConntrackInfo)
	c.connReq <- r
	return <-r
}

// Follow returns a channel with all changes.
// NOTE: currently we only return connection is ESTABLISHED state.
func (c *ConnTrack) Follow() (<-chan ConntrackInfo, func(), error) {
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
		err := readMessagesFromNetfilter(s, func(conntrackInfo ConntrackInfo) {
			if c.filterFunc(conntrackInfo) {
				res <- conntrackInfo
			}
		})
		if err != nil {
			glog.Fatalf("Error reading message from Netfilter: %++v", err)
			panic(err)
		}
	}()
	return res, stop, nil
}
