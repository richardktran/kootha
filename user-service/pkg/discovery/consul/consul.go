package consul

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	consul "github.com/hashicorp/consul/api"
	"github.com/richardktran/realtime-quiz/user-service/pkg/discovery"
)

type Registry struct {
	client *consul.Client
}

// Define port for Consul
func NewRegistry(addr string) (*Registry, error) {
	config := consul.DefaultConfig()
	config.Address = addr

	client, err := consul.NewClient(config)

	if err != nil {
		return nil, err
	}

	return &Registry{client: client}, nil
}

func (r *Registry) Register(ctx context.Context, instanceID, serviceName, hostPort string) error {
	address, port, err := splitHostPort(hostPort)

	if err != nil {
		return err
	}

	portString, err := strconv.Atoi(port)

	if (err) != nil {
		return err
	}

	agent := r.client.Agent()

	return agent.ServiceRegister(&consul.AgentServiceRegistration{
		Address: address,
		ID:      instanceID,
		Name:    serviceName,
		Port:    portString,
		Check: &consul.AgentServiceCheck{
			CheckID: instanceID,
			TTL:     "5s",
		},
	})
}

func (r *Registry) Deregister(ctx context.Context, instanceId string) error {
	agent := r.client.Agent()

	return agent.ServiceDeregister(instanceId)
}

// ServiceAddresses returns the list of addresses of active instances of a given service
func (r *Registry) ServiceAddresses(ctx context.Context, serviceName string) ([]string, error) {
	entries, _, err := r.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, err
	} else if len(entries) == 0 {
		return nil, discovery.ErrNotFound
	}

	var res []string
	for _, entry := range entries {
		res = append(res, fmt.Sprintf("%s:%d", entry.Service.Address, entry.Service.Port))
	}

	return res, nil
}

func (r *Registry) ReportHealthyState(instanceID, serviceName string) error {
	agent := r.client.Agent()

	return agent.PassTTL(instanceID, "")
}

func splitHostPort(hostPort string) (string, string, error) {
	parts := strings.Split(hostPort, ":")
	if len(parts) != 2 {
		return "", "", errors.New("hostPort must be in a form of <host>:<port>, example: localhost:8081")
	}

	return parts[0], parts[1], nil
}
