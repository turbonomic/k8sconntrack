package transactioncounter

type Transaction struct {
	ServiceId           string         `json:"serviceID,omitempty"`
	EndpointsCounterMap map[string]int `json:"endpointCounter,omitempty"`
}

func (this *Transaction) GetEndpointsCounterMap() map[string]int {
	return this.EndpointsCounterMap
}
