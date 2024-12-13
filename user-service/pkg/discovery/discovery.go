package discovery

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/exp/rand"
)

type Registry interface {
	// Register creates a service instance in the registry
	Register(ctx context.Context, instanceID string, serviceName string, hostPort string) error

	// Deregister removes a service instance from the registry
	Deregister(ctx context.Context, instanceID string) error

	// GetServiceAddresses returns the addresses of active instances of a given service
	ServiceAddresses(ctx context.Context, serviceName string) ([]string, error)

	// ReportHealthyState reports the service instance is healthy
	ReportHealthyState(instanceID string, serviceName string) error
}

var ErrNotFound = errors.New("no service addresses found")

func GenerateInstanceID(serviceName string) string {
	return fmt.Sprintf("%s-%d", serviceName, rand.New(rand.NewSource(uint64(time.Now().UnixNano()))).Int())
}
