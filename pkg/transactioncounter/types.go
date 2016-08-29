package transactioncounter

type Transaction struct {
	ServiceId           string             `json:"serviceID,omitempty"`
	EndpointsCounterMap map[string]float64 `json:"endpointCounter,omitempty"`
	EpCountAbs          map[string]int     `json:"endpointAbs,omitempty"`
}

func (this *Transaction) GetEndpointsCounterMap() map[string]float64 {
	return this.EndpointsCounterMap
}
