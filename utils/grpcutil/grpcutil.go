package grpcutil

import (
	"context"

	"github.com/richardktran/realtime-quiz/pkg/discovery"
	"golang.org/x/exp/rand"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func ServiceConnection(ctx context.Context, serviceName string, registry discovery.Registry) (*grpc.ClientConn, error) {
	addrs, err := registry.ServiceAddresses(ctx, serviceName)
	if err != nil {
		return nil, err
	}

	addr := addrs[rand.Intn(len(addrs))]

	return grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
}
