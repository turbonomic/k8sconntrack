package transactioncounter

import (
	//	"fmt"
	"testing"

	"github.com/dongyiyang/k8sconnection/pkg/k8sconnector"
)

func TestCount(t *testing.T) {
	tests := []struct {
		CounterMap      map[string]*Transaction
		ServiceName     string
		EndpointAddress string
		ExpectedCount   int
	}{
		{
			CounterMap:      map[string]*Transaction{},
			ServiceName:     "service1",
			EndpointAddress: "10.0.0.2",
			ExpectedCount:   1,
		},
		{
			CounterMap: map[string]*Transaction{
				"service1": &Transaction{ServiceId: "service1",
					EndpointsCounterMap: map[string]int{
						"10.0.1.2": 3,
					}},
			},
			ServiceName:     "service1",
			EndpointAddress: "10.0.0.2",
			ExpectedCount:   1,
		},
		{
			CounterMap: map[string]*Transaction{
				"service1": &Transaction{ServiceId: "service1",
					EndpointsCounterMap: map[string]int{
						"10.0.0.2": 3,
					}},
			},
			ServiceName:     "service1",
			EndpointAddress: "10.0.0.2",
			ExpectedCount:   4,
		},
	}

	for _, test := range tests {
		transactionCounter := NewTransactionCounter(k8sconnector.NewFakeConnector())
		transactionCounter.counter = test.CounterMap
		transactionCounter.Count(test.ServiceName, test.EndpointAddress)
		var c int
		if transaction, exist := transactionCounter.counter[test.ServiceName]; exist {
			if epCounts, has := transaction.EndpointsCounterMap[test.EndpointAddress]; has {
				c = epCounts
			} else {
				t.Errorf("Endpoint %s is not found in transaction info for service %s.", test.EndpointAddress, test.ServiceName)
			}
		} else {
			t.Errorf("Service %s is not found in transaction counter map.", test.ServiceName)
		}
		if c != test.ExpectedCount {
			t.Errorf("Expected count is %d, got %d", test.ExpectedCount, c)
		}
	}
}
