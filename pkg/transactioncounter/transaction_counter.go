package transactioncounter

import (
	"time"

	"github.com/dongyiyang/k8sconnection/pkg/conntrack"
	"github.com/dongyiyang/k8sconnection/pkg/k8sconnector"

	"github.com/golang/glog"
)

type TransactionCounter struct {
	connector k8sconnector.Connector

	// key is service name, value is the transaction related to it.
	counter map[string]map[string]int

	lastPollTimestamp uint64
}

func NewTransactionCounter(connector k8sconnector.Connector) *TransactionCounter {
	counterMap := make(map[string]map[string]int)
	return &TransactionCounter{
		counter:   counterMap,
		connector: connector,
	}
}

// Clear the transaction counter map.
func (tc *TransactionCounter) Reset() {
	glog.V(3).Infof("Inside reset transaction counter")
	counterMap := make(map[string]map[string]int)

	tc.counter = counterMap

	// As after each poll, the counter map is cleaned, so this is the right place to set the lastPollTimestamp.
	tc.lastPollTimestamp = uint64(time.Now().Unix())
}

// Increment the transaction count for a single endpoint.
// Transaction counter map uses serviceName as key and endpoint map as value.
// In endpoint map, key is endpoint IP address, value is the number of transaction happened on the endpoint.
func (tc *TransactionCounter) Count(serviceName, endpointAddress string) {
	epMap, ok := tc.counter[serviceName]
	if !ok {
		epMap = make(map[string]int)
	}
	count, ok := epMap[endpointAddress]
	if !ok {
		glog.Infof("Service %s is not tracked. Now initializing in map", serviceName)
		count = 0
	}
	epMap[endpointAddress] = count + 1
	tc.counter[serviceName] = epMap
	glog.V(5).Infof("counter map is %v", tc)
	glog.V(4).Infof("Transaction count of %s is %d.", endpointAddress, epMap[endpointAddress])
}

func (tc *TransactionCounter) GetAllTransactions() []*Transaction {
	var transactions []*Transaction

	// Here we need to translate the absolute count value into count/second.
	if tc.lastPollTimestamp == 0 {
		// When lastPollTimestamp is 0, meaning that current poll is the first poll. We cannot get the count/s, so just return.
		return transactions
	}
	// Get the time difference between two poll.
	timeDiff := uint64(time.Now().Unix()) - tc.lastPollTimestamp
	glog.V(4).Infof("Time diff is %d", timeDiff)

	for svcName, epMap := range tc.counter {
		// Before append, change count to count per second.
		valueMap := make(map[string]float64)
		countMap := make(map[string]int)
		for ep, count := range epMap {
			valueMap[ep] = float64(count) / float64(timeDiff)
			countMap[ep] = count
		}
		transaction := &Transaction{
			ServiceId:           svcName,
			EndpointsCounterMap: valueMap,
			EpCountAbs:          countMap,
		}
		glog.V(4).Infof("Get transaction data: %++v", transaction)
		transactions = append(transactions, transaction)
	}

	return transactions
}

// Get all the current Established TCP connections from conntrack and add count to transaction counter.
func (tc *TransactionCounter) ProcessConntrackConnections(connections []conntrack.TCPConnection) {
	if len(connections) > 0 {
		glog.V(3).Infof("Connections:\n")
		for _, cn := range connections {
			address := cn.Local
			svcName, err := tc.connector.GetServiceNameWithEndpointAddress(address)
			if err != nil {
				glog.Errorf("\tError getting svc name\n")
			}
			tc.Count(svcName, address)
		}
	}
}
