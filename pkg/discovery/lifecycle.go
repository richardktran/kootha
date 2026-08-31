package discovery

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/richardktran/realtime-quiz/pkg/env"
)

// AdvertiseAddr is host:port registered with Consul. In Kubernetes, Register is a no-op.
func AdvertiseAddr(port string) string {
	host := env.GetString("SERVICE_ADVERTISE_HOST", "localhost")
	return fmt.Sprintf("%s:%s", host, port)
}

// StartHeartbeat registers the instance and reports a TTL health check every 5s.
// Safe to call with a k8s Registry (Register / ReportHealthyState are no-ops).
func StartHeartbeat(ctx context.Context, registry Registry, serviceName, hostPort string) (stop func(), err error) {
	instanceID := GenerateInstanceID(serviceName)
	if err := registry.Register(ctx, instanceID, serviceName, hostPort); err != nil {
		return nil, err
	}

	done := make(chan struct{})
	go func() {
		_ = registry.ReportHealthyState(instanceID, serviceName)
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := registry.ReportHealthyState(instanceID, serviceName); err != nil {
					log.Printf("Failed to report healthy state: %v", err)
				}
			}
		}
	}()

	return func() {
		close(done)
		_ = registry.Deregister(ctx, instanceID)
	}, nil
}
