package transactioncounter

import (
	"github.com/dongyiyang/k8sconnection/pkg/conntrack"
	"github.com/dongyiyang/k8sconnection/pkg/k8sconnector"

	"github.com/golang/glog"
)

type TransactionCounter struct {
	connector *k8sconnector.K8sConnector

	// key is service name, value is the transaction related to it.
	counter map[string]*Transaction
}

func NewTransactionCounter(connector *k8sconnector.K8sConnector) *TransactionCounter {
	counterMap := make(map[string]*Transaction)
	return &TransactionCounter{
		counter:   counterMap,
		connector: connector,
	}
}

// Clear the transaction counter map.
func (tc *TransactionCounter) Reset() {
	glog.V(3).Infof("Inside reset transaction counter")
	counterMap := make(map[string]*Transaction)

	tc.counter = counterMap
}

// Increment the transaction count for a single endpoint.
// Transaction counter map uses serviceName as key and endpoint map as value.
// In endpoint map, key is endpoint IP address, value is the number of transaction happened on the endpoint.
func (tc *TransactionCounter) Count(serviceName, endpointAddress string) {
	trans, ok := tc.counter[serviceName]
	if !ok {
		epMap := make(map[string]int)
		trans = &Transaction{
			ServiceId:           serviceName,
			EndpointsCounterMap: epMap,
		}
		tc.counter[serviceName] = trans
	}
	count, ok := trans.EndpointsCounterMap[endpointAddress]
	if !ok {
		glog.Infof("Service %s is not tracked. Now initializing in map", serviceName)
		count = 0
	}
	trans.EndpointsCounterMap[endpointAddress] = count + 1
	glog.V(5).Infof("counter map is %v", tc)
	glog.V(4).Infof("Transaction value of %s is %d.", endpointAddress, trans.EndpointsCounterMap[endpointAddress])
}

func (tc *TransactionCounter) GetAllTransactions() []*Transaction {
	var transactions []*Transaction

	for _, value := range tc.counter {
		transactions = append(transactions, value)
	}

	glog.V(4).Infof("Get All transaction %v", transactions)
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
