package options

import (
	"github.com/spf13/pflag"
)

type K8sConntrackConfig struct {
	Master     string
	Kubeconfig string

	ConntrackBindAddress string
	ConntrackPort        string

	EnableConnectionCounter bool
	EnableFlowCollector     bool
	SocketBufferSize        string
}

func NewK8sConntrackConfig() *K8sConntrackConfig {
	return &K8sConntrackConfig{
		ConntrackBindAddress: "0.0.0.0",
	}
}

// Add parameters passed from command line.
func (s *K8sConntrackConfig) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.Master, "master", s.Master, "The address of the Kubernetes API server (overrides any value in kubeconfig)")
	fs.StringVar(&s.Kubeconfig, "kubeconfig", s.Kubeconfig, "Path to kubeconfig file with authorization and master location information.")
	fs.StringVar(&s.ConntrackBindAddress, "conntrack-bind-address", s.ConntrackBindAddress, "The IP address for the conntrack server to serve on, defaulting to 0.0.0.0 (set to 127.0.0.1 for local).")
	fs.StringVar(&s.ConntrackPort, "conntrack-port", "2222", "The port to bind the k8sconntrack server.")
	fs.BoolVar(&s.EnableConnectionCounter, "enable-connection-counter", true, "If set false, explicitly disable connection connector.")
	fs.BoolVar(&s.EnableFlowCollector, "enable-flow-collector", true, "If set false, explicitly disable flow collector.")
	fs.StringVar(&s.SocketBufferSize, "buffer-size", "", "Set the buffer size when communicating with kernel. Can set small, medium, large.")
}
