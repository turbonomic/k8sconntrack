package flowcollector

import (
	"fmt"
	"time"

	"github.com/dongyiyang/k8sconnection/pkg/conntrack"
	"github.com/dongyiyang/k8sconnection/pkg/k8sconnector"

	"github.com/golang/glog"
)

// Flow collector requires user to turn on nf_conntrack_acct and nf_conntrack_timestamp

type FlowCollector struct {
	connector *k8sconnector.K8sConnector

	// flows is a map, key is flow UID, value is Flow instance.
	flows []*Flow
}

func NewFlowCollector(connector *k8sconnector.K8sConnector) *FlowCollector {
	return &FlowCollector{
		connector: connector,
	}
}

// TODO: For POC: key is src:srcPort->dest:destPort#startTimestamp
var prevConnectionMap map[string]*conntrack.ConntrackInfo = make(map[string]*conntrack.ConntrackInfo)

func keyFunc(info *conntrack.ConntrackInfo) string {
	if info == nil {
		return ""
	}
	return fmt.Sprintf("%s:%d->%s:%d#%d",
		info.Src, info.SrcPort, info.Dst, info.DstPort, info.StartTimestamp)
}

func (this *FlowCollector) TrackFlow() {
	// Track flow
	infos, err := conntrack.ListConntrackInfos(this.flowConnectionFilterFunc)
	if err != nil {
		panic(err)
	}
	if len(infos) < 1 {
		return
	}

	// build flow based on connections
	var currConntrackInfos map[string]*conntrack.ConntrackInfo = make(map[string]*conntrack.ConntrackInfo)
	for _, i := range infos {
		info := &i
		key := keyFunc(info)
		currConntrackInfos[key] = info
		if prevInfo, exist := prevConnectionMap[key]; exist {
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
	prevConnectionMap = currConntrackInfos

}

func (this *FlowCollector) flowConnectionFilterFunc(c conntrack.ConntrackInfo) bool {
	// Here we only care about updated info
	if c.MsgType != conntrack.NfctMsgUpdate {
		glog.V(4).Infof("Message isn't an update: %d\n", c.MsgType)
		return false
	}

	src := c.Src.String()
	dst := c.Dst.String()
	podIPs := this.connector.FindPodsIP()
	_, srcPodLocal := podIPs[src]
	_, dstPodLocal := podIPs[dst]

	// Only monitor connections between pods.
	if !srcPodLocal || !dstPodLocal {
		return false
	}

	// As for updated info, we only care about ESTABLISHED for now.
	if c.TCPState != "ESTABLISHED" {
		glog.V(4).Infof("State isn't in ESTABLISHED: %s\n", c.TCPState)
		return false
	}

	if c.TCPState == "ESTABLISHED" {
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

// Get the flows within the time between current call and previous call,
// calculate the average flow values for each unique connection
func (this *FlowCollector) GetAllUniqueFlows() []*Flow {
	uniqueConnectionToFlows := make(map[string][]*Flow)
	for _, flow := range this.flows {
		var flowList []*Flow
		if fs, exist := uniqueConnectionToFlows[flow.UID]; exist {
			flowList = fs
		}
		flowList = append(flowList, flow)
		uniqueConnectionToFlows[flow.UID] = flowList
	}

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
