package server

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	fcollector "github.com/dongyiyang/k8sconnection/pkg/flowcollector"
	tcounter "github.com/dongyiyang/k8sconnection/pkg/transactioncounter"

	"github.com/golang/glog"
)

// Server is a http.Handler which exposes kubelet functionality over HTTP.
type Server struct {
	counter       *tcounter.TransactionCounter
	flowCollector *fcollector.FlowCollector
	mux           *http.ServeMux
}

// NewServer initializes and configures a kubelet.Server object to handle HTTP requests.
func NewServer(counter *tcounter.TransactionCounter, flowCollector *fcollector.FlowCollector) Server {
	server := Server{
		counter:       counter,
		flowCollector: flowCollector,
		mux:           http.NewServeMux(),
	}
	server.InstallDefaultHandlers()
	return server
}

// InstallDefaultHandlers registers the default set of supported HTTP request patterns with the mux.
func (s *Server) InstallDefaultHandlers() {
	s.mux.HandleFunc("/", handler)
	s.mux.HandleFunc("/transactions/count", s.getTransactionsCount)
	s.mux.HandleFunc("/transactions", s.getAllTransactionsAndReset)
	s.mux.HandleFunc("/flows", s.getAllFlows)
}

// ServeHTTP responds to HTTP requests on the Kubelet.
func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.mux.ServeHTTP(w, req)
}

func (s *Server) getAllTransactionsAndReset(w http.ResponseWriter, r *http.Request) {
	transactions := s.counter.GetAllTransactions()

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(transactions); err != nil {
		panic(err)
	}

	s.resetCounter()
}

func (s *Server) getTransactionsCount(w http.ResponseWriter, r *http.Request) {
	transactions := s.counter.GetAllTransactions()

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(transactions); err != nil {
		panic(err)
	}
}

func (s *Server) getAllFlows(w http.ResponseWriter, r *http.Request) {
	flows := s.flowCollector.GetAllFlows()

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(flows); err != nil {
		panic(err)
	}
}

func (s *Server) resetCounter() {
	s.counter.Reset()
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Vmturbo k8sconntrack Service.")
}

// TODO: For now the address and port number is hardcoded. The actual port number need to be discussed.
func ListenAndServeProxyServer(counter *tcounter.TransactionCounter, flowCollector *fcollector.FlowCollector) {
	glog.V(3).Infof("Start VMT Kube-proxy server")
	handler := NewServer(counter, flowCollector)
	s := &http.Server{
		Addr:           net.JoinHostPort("0.0.0.0", "2223"),
		Handler:        &handler,
		MaxHeaderBytes: 1 << 20,
	}
	glog.Fatal(s.ListenAndServe())
}
