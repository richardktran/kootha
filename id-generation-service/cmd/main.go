package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/richardktran/realtime-quiz/gen"
	grpcHandler "github.com/richardktran/realtime-quiz/id-generation-service/internal/handler/grpc"
	idgeneration "github.com/richardktran/realtime-quiz/id-generation-service/internal/service/idgeneration"
	"github.com/richardktran/realtime-quiz/pkg/discovery"
	"github.com/richardktran/realtime-quiz/pkg/discovery/consul"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gopkg.in/yaml.v3"
)

const serviceName = "id-generation-service"

type serviceConfig struct {
	APIConfig apiConfig `yaml:"api"`
}

type apiConfig struct {
	Port string `yaml:"port"`
}

func main() {
	f, err := os.Open("id-generation-service/configs/base.yaml")
	if err != nil {
		panic(err)
	}

	defer f.Close()

	var cfg serviceConfig

	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		panic(err)
	}

	port := cfg.APIConfig.Port

	log.Printf("Starting the %s service with port %v...", serviceName, port)

	registry, err := consul.NewRegistry("localhost:8500")
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	instanceId := discovery.GenerateInstanceID(serviceName)

	if err := registry.Register(ctx, instanceId, serviceName, fmt.Sprintf("localhost:%v", port)); err != nil {
		panic(err)
	}

	go func() {
		for {
			if err := registry.ReportHealthyState(instanceId, serviceName); err != nil {
				log.Printf("Failed to report healthy state: %v", err)
				time.Sleep(5 * time.Second)
			}
		}
	}()
	defer registry.Deregister(ctx, instanceId)

	// =============== This section is for gRPC server ===============
	svc := idgeneration.New()
	h := grpcHandler.New(svc)
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)
	gen.RegisterIdGenerationServiceServer(grpcServer, h)

	if err := grpcServer.Serve(listener); err != nil {
		panic(err)
	}

	// =============== This section is for HTTP handler ===============
	// svc := idgeneration.New()
	// h := httpHandler.New(svc)
	// http.Handle("/v1/id", http.HandlerFunc(h.GenerateId))

	// if err := http.ListenAndServe(fmt.Sprintf(":%v", port), nil); err != nil {
	// 	panic(err)
	// }
}
