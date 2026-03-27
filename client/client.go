package client

import (
	"github.com/t0mk/datui/config"
	"k8s.io/client-go/dynamic"
)

// New creates a dynamic client from the loaded config.
func New(cfg *config.Config) (dynamic.Interface, error) {
	return dynamic.NewForConfig(cfg.RestConfig)
}
