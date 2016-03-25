package transactioncounter

import (
	"github.com/golang/glog"
)

type TransactionCounter struct {
	// key is service name, value is the transaction related to it.
	counter map[string]*Transaction
}

func NewTransactionCounter() *TransactionCounter {
	counterMap := make(map[string]*Transaction)
	return &TransactionCounter{
		counter: counterMap,
	}
}

func (tc *TransactionCounter) Reset() {
	glog.V(3).Infof("Inside reset transaction counter")
	counterMap := make(map[string]*Transaction)

	tc.counter = counterMap
}

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
	glog.V(5).Infof("Get All transaction from tc %v", tc.counter)

	glog.V(4).Infof("Get All transaction %v", transactions)
	return transactions
}
