package options

import (
	"github.com/spf13/pflag"
)

type K8sConntrackConfig struct {
	Master     string
	Kubeconfig string
}

func NewK8sConntrackConfig() *K8sConntrackConfig {
	return &K8sConntrackConfig{}
}

// Add parameters passed from command line.
func (s *K8sConntrackConfig) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.Master, "master", s.Master, "The address of the Kubernetes API server (overrides any value in kubeconfig)")
	fs.StringVar(&s.Kubeconfig, "kubeconfig", s.Kubeconfig, "Path to kubeconfig file with authorization and master location information.")
}
