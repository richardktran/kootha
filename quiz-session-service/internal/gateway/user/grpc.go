package user

import (
	"context"

	"github.com/richardktran/realtime-quiz/gen"
	"github.com/richardktran/realtime-quiz/pkg/discovery"
	userModel "github.com/richardktran/realtime-quiz/user-service/pkg/model"
	"github.com/richardktran/realtime-quiz/utils/grpcutil"
)

var serviceName = "user-service"

type Gateway struct {
	registry discovery.Registry
}

func New(registry discovery.Registry) *Gateway {
	return &Gateway{
		registry: registry,
	}
}

func (g *Gateway) CreateUser(ctx context.Context, name string) (*userModel.User, error) {
	conn, err := grpcutil.ServiceConnection(ctx, serviceName, g.registry)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := gen.NewUserServiceClient(conn)

	response, err := client.CreateUser(ctx, &gen.CreateUserRequest{
		Name: name,
	})
	if err != nil {
		return nil, err
	}

	return userModel.UserFromProto(response.GetUser()), nil
}
