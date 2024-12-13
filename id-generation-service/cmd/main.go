package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	httpHandler "github.com/richardktran/realtime-quiz/id-generation-service/internal/handler/http"
	idgeneration "github.com/richardktran/realtime-quiz/id-generation-service/internal/service/idGeneration"
	"github.com/richardktran/realtime-quiz/user-service/pkg/discovery"
	"github.com/richardktran/realtime-quiz/user-service/pkg/discovery/consul"
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

	svc := idgeneration.New()
	h := httpHandler.New(svc)
	http.Handle("/v1/id", http.HandlerFunc(h.GenerateId))

	if err := http.ListenAndServe(fmt.Sprintf(":%v", port), nil); err != nil {
		panic(err)
	}
}
