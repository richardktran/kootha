package k8s

import (
	"context"

	"github.com/richardktran/realtime-quiz/pkg/discovery"
	"github.com/richardktran/realtime-quiz/pkg/env"
)

// Registry resolves services via Kubernetes DNS (or explicit env overrides).
// Register / Deregister / ReportHealthyState are no-ops — ClusterIP Services handle routing.
type Registry struct {
	addrs map[string]string
}

func NewRegistry() *Registry {
	return &Registry{
		addrs: map[string]string{
			"user-service":          env.GetString("USER_SERVICE_ADDR", "user-service:8082"),
			"id-generation-service": env.GetString("ID_GENERATION_SERVICE_ADDR", "id-generation-service:8083"),
			"quiz-bank":             env.GetString("QUIZ_BANK_SERVICE_ADDR", "quiz-bank-service:8085"),
			"quiz-session":          env.GetString("QUIZ_SESSION_SERVICE_ADDR", "quiz-session-service:8084"),
			"notification":          env.GetString("NOTIFICATION_SERVICE_ADDR", "notification-service:8086"),
		},
	}
}

func (r *Registry) Register(ctx context.Context, instanceID, serviceName, hostPort string) error {
	return nil
}

func (r *Registry) Deregister(ctx context.Context, instanceID string) error {
	return nil
}

func (r *Registry) ServiceAddresses(ctx context.Context, serviceName string) ([]string, error) {
	addr, ok := r.addrs[serviceName]
	if !ok || addr == "" {
		return nil, discovery.ErrNotFound
	}
	return []string{addr}, nil
}

func (r *Registry) ReportHealthyState(instanceID, serviceName string) error {
	return nil
}
