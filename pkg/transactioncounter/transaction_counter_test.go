package transactioncounter

import (
	"testing"
)

func TestCount(t *testing.T) {
	tests := []struct {
		CounterMap      map[string]map[string]int
		ServiceName     string
		EndpointAddress string
		ExpectedCount   int
	}{
		{
			CounterMap:      map[string]map[string]int{},
			ServiceName:     "service1",
			EndpointAddress: "10.0.0.2",
			ExpectedCount:   1,
		},
		{
			CounterMap: map[string]map[string]int{
				"service1": map[string]int{
					"10.0.1.2": 3,
				},
			},
			ServiceName:     "service1",
			EndpointAddress: "10.0.0.2",
			ExpectedCount:   1,
		},
		{
			CounterMap: map[string]map[string]int{
				"service1": map[string]int{
					"10.0.0.2": 3,
				},
			},
			ServiceName:     "service1",
			EndpointAddress: "10.0.0.2",
			ExpectedCount:   4,
		},
	}

	for _, test := range tests {
		transactionCounter := NewTransactionCounter(nil)
		transactionCounter.counter = test.CounterMap
		transactionCounter.Count([]*countInfo{&countInfo{test.ServiceName, test.EndpointAddress}})
		var c int
		if epMap, exist := transactionCounter.counter[test.ServiceName]; exist {
			if epCounts, has := epMap[test.EndpointAddress]; has {
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
