package registry

import (
	"fmt"
	"strings"

	"github.com/richardktran/realtime-quiz/pkg/discovery"
	"github.com/richardktran/realtime-quiz/pkg/discovery/consul"
	"github.com/richardktran/realtime-quiz/pkg/discovery/k8s"
	"github.com/richardktran/realtime-quiz/pkg/env"
)

// New returns Consul (local Makefile) or Kubernetes DNS (in-cluster).
//
//	DISCOVERY_BACKEND=consul  (default) — CONSUL_ADDR, default localhost:8500
//	DISCOVERY_BACKEND=k8s     — ClusterIP DNS, optional *_SERVICE_ADDR overrides
func New() (discovery.Registry, error) {
	backend := strings.ToLower(env.GetString("DISCOVERY_BACKEND", "consul"))
	switch backend {
	case "k8s", "kubernetes", "static":
		return k8s.NewRegistry(), nil
	case "consul", "":
		addr := env.GetString("CONSUL_ADDR", "localhost:8500")
		return consul.NewRegistry(addr)
	default:
		return nil, fmt.Errorf("unknown DISCOVERY_BACKEND %q (use consul or k8s)", backend)
	}
}
