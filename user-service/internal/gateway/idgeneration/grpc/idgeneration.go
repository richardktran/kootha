package grpc

import (
	"context"

	"github.com/richardktran/realtime-quiz/gen"
	"github.com/richardktran/realtime-quiz/id-generation-service/pkg/model"
	"github.com/richardktran/realtime-quiz/pkg/discovery"
	"github.com/richardktran/realtime-quiz/utils/grpcutil"
)

var serviceName = "id-generation-service"

type Gateway struct {
	registry discovery.Registry
}

func New(registry discovery.Registry) *Gateway {
	return &Gateway{
		registry: registry,
	}
}

func (g *Gateway) GenerateId(ctx context.Context, entity string) (*model.IDGenerator, error) {
	conn, err := grpcutil.ServiceConnection(ctx, serviceName, g.registry)

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	client := gen.NewIdGenerationServiceClient(conn)
	response, err := client.GenerateId(ctx, &gen.IdGenerationRequest{
		Entity: entity,
	})

	if err != nil {
		return nil, err
	}

	return model.IDGeneratorFromProto(response.GetIdGenerator()), nil
}
