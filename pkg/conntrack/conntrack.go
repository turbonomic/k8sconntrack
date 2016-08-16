package conntrack

import (
	"fmt"
	"time"

	"github.com/golang/glog"
)

// TCPConnection is a connection
type TCPConnection struct {
	Local      string // net.IP
	LocalPort  string // int
	Remote     string // net.IP
	RemotePort string // int
}

func (c TCPConnection) String() string {
	return fmt.Sprintf("%s:%s->%s:%s", c.Local, c.LocalPort, c.Remote, c.RemotePort)
}

// ConnTrack monitors the connections.
type ConnTrack struct {
	connReq chan chan []TCPConnection
	quit    chan struct{}
}

// New returns a ConnTrack.
func New() (*ConnTrack, error) {
	c := ConnTrack{
		connReq: make(chan chan []TCPConnection),
		quit:    make(chan struct{}),
	}
	go func() {
		for {
			err := c.track()
			select {
			case <-c.quit:
				return
			default:
			}
			if err != nil {
				glog.Errorf("conntrack: %s\n", err)
			}
			time.Sleep(1 * time.Second)
		}
	}()

	return &c, nil
}

// Close stops all monitoring and executables.
func (c ConnTrack) Close() {
	close(c.quit)
}

// Connections returns the list of all connections seen since last time you
// called it.
func (c *ConnTrack) Connections() []TCPConnection {
	r := make(chan []TCPConnection)
	c.connReq <- r
	return <-r
}

// track is the main loop
func (c *ConnTrack) track() error {
	// We use Follow() to keep track of conn state changes, but it doesn't give
	// us the initial state.
	events, stop, err := Follow()
	if err != nil {
		return err
	}

	establishedConns, err := ListConnections(func(c ConntrackInfo) bool {
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
		glog.V(4).Infof("conn is %++v \n", c)
		return true
	})
	if err != nil {
		return fmt.Errorf("Error listing ESTABLISHED connections: %++v.", err)
	}

	established := map[TCPConnection]struct{}{}
	for _, c := range establishedConns {
		established[c] = struct{}{}
	}

	podIPs := FindPodIPs()
	updateLocalIPs := time.Tick(time.Minute)

	for {
		select {

		case <-c.quit:
			stop()
			return nil

		case <-updateLocalIPs:
			podIPs = FindPodIPs()

		case e, ok := <-events:
			if !ok {
				return nil
			}
			switch {

			default:
				// not interested

			case e.TCPState == TCPState_ESTABLISHED:
				conns := e.BuildTCPConn(podIPs)
				for _, cn := range conns {
					if cn == nil {
						// log.Printf("not a local connection: %+v\n", e)
						continue
					}
					established[*cn] = struct{}{}
					glog.V(4).Infof("Established Connection payload is %++v", cn)
				}

			case e.MsgType == NfctMsgDestroy, e.TCPState == TCPState_TIME_WAIT, e.TCPState == TCPState_CLOSE:
				// NOTE Since in Follow(), it only sends back conneciton with ESTABLISHED state,
				// So this part of code would never be hit under current logic.
			}

		case r := <-c.connReq:
			cs := make([]TCPConnection, 0, len(established))
			for c := range established {
				cs = append(cs, c)
			}
			r <- cs
			established = map[TCPConnection]struct{}{}
		}
	}
}
