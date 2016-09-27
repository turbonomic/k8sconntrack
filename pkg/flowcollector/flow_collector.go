package flowcollector

import (
	"fmt"
	"sync"
	"time"

	"k8s.io/kubernetes/pkg/api"

	"github.com/dongyiyang/k8sconnection/pkg/conntrack"
	"github.com/dongyiyang/k8sconnection/pkg/k8sconnector"

	"github.com/golang/glog"
)

// Flow collector requires user to turn on nf_conntrack_acct and nf_conntrack_timestamp

type FlowCollector struct {
	connector k8sconnector.Connector

	// Protects endpointsSet.
	mu sync.Mutex

	// endpintsMap keeps track of current endpoints in K8s cluster.
	endpointsSet map[string]bool

	// A map keeps track of ConntrackInfo
	// TODO: For POC: key is src:srcPort->dest:destPort#startTimestamp
	conntrackInfoMap map[string]*conntrack.ConntrackInfo

	// flows is a map, key is flow UID, value is Flow instance.
	flows []*Flow
}

func NewFlowCollector(connector k8sconnector.Connector) *FlowCollector {
	return &FlowCollector{
		connector:        connector,
		endpointsSet:     make(map[string]bool),
		conntrackInfoMap: make(map[string]*conntrack.ConntrackInfo),
	}
}

// Implement k8s.io/pkg/proxy/config/EndpointsConfigHandler Interface.
func (this *FlowCollector) OnEndpointsUpdate(allEndpoints []api.Endpoints) {
	start := time.Now()
	defer func() {
		glog.V(4).Infof("OnEndpointsUpdate took %v for %d endpoints", time.Since(start), len(allEndpoints))
	}()

	this.mu.Lock()
	defer this.mu.Unlock()

	// Clear the current endpoints set.
	this.endpointsSet = make(map[string]bool)

	for i := range allEndpoints {
		endpoints := &allEndpoints[i]
		for j := range endpoints.Subsets {
			ss := &endpoints.Subsets[j]
			for k := range ss.Addresses {
				addr := &ss.Addresses[k]
				this.endpointsSet[addr.IP] = true
			}
		}
	}

	this.syncConntrackInfo()
}

func keyFunc(info *conntrack.ConntrackInfo) string {
	if info == nil {
		return ""
	}
	return fmt.Sprintf("%s:%d->%s:%d#%d",
		info.Src, info.SrcPort, info.Dst, info.DstPort, info.StartTimestamp)
}

func (this *FlowCollector) TrackFlow() {
	this.mu.Lock()
	defer this.mu.Unlock()

	this.syncConntrackInfo()
}

// Get valid ConntrackInfo from Conntrack and build Flow Objects.
func (this *FlowCollector) syncConntrackInfo() {
	// Track flow
	infos, err := conntrack.ListConntrackInfos(this.flowConnectionFilterFunc)
	if err != nil {
		panic(err)
	}
	if len(infos) < 1 {
		glog.Infof("No Data")
		return
	}

	// build flow based on connections
	var currConntrackInfos map[string]*conntrack.ConntrackInfo = make(map[string]*conntrack.ConntrackInfo)
	for _, i := range infos {
		info := i
		key := keyFunc(&info)
		currConntrackInfos[key] = &info
		if prevInfo, exist := this.conntrackInfoMap[key]; exist {
			bytesDiff := info.Bytes - prevInfo.Bytes
			timeDiff := info.DeltaTime - prevInfo.DeltaTime
			if timeDiff == 0 {
				continue
			}

			flowValue := bytesDiff / timeDiff
			flow := &Flow{
				UID:                  key,
				Src:                  info.Src,
				Dst:                  info.Dst,
				Value:                flowValue,
				LastUpdatedTimestamp: uint64(time.Now().Unix()),
			}
			glog.V(4).Infof("Flow (UID: %s) between %s and %s is %d",
				flow.UID, flow.Src, flow.Dst, flow.Value)
			this.flows = append(this.flows, flow)
		}
	}
	this.conntrackInfoMap = currConntrackInfos
}

func (this *FlowCollector) flowConnectionFilterFunc(c conntrack.ConntrackInfo) bool {
	// Here we only care about updated info
	if c.MsgType != conntrack.NfctMsgUpdate {
		glog.V(4).Infof("Message isn't an update: %d\n", c.MsgType)
		return false
	}

	src := c.Src.String()
	dst := c.Dst.String()
	//	podIPs := this.connector.FindPodsIP()
	//	_, srcPodLocal := podIPs[src]
	//	_, dstPodLocal := podIPs[dst]
	_, srcPodLocal := this.endpointsSet[src]
	_, dstPodLocal := this.endpointsSet[dst]

	// NOTE: Only monitor connections between endpoints.
	if !srcPodLocal || !dstPodLocal {
		return false
	}

	// As for updated info, we only care about ESTABLISHED for now.
	if c.TCPState != conntrack.TCPState_ESTABLISHED {
		glog.V(4).Infof("State isn't in ESTABLISHED: %s\n", c.TCPState)
		return false
	}

	if c.TCPState == conntrack.TCPState_ESTABLISHED {
		glog.V(4).Infof("EEEEEEEE  ESTABLISHED conn is %v \n", c)
	}

	return true
}

func (this *FlowCollector) GetAllFlows() []*Flow {
	var result []*Flow
	for _, flow := range this.flows {
		result = append(result, flow)
	}
	glog.V(4).Infof("Get All flows %++v", result)

	return result
}

func (this *FlowCollector) Reset() {
	var newFlows []*Flow
	this.flows = newFlows
}
