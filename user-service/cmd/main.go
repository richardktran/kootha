package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/richardktran/realtime-quiz/user-service/pkg/discovery"
	"github.com/richardktran/realtime-quiz/user-service/pkg/discovery/consul"
	"gopkg.in/yaml.v3"
)

const serviceName = "user"

type serviceConfig struct {
	APIConfig apiConfig `yaml:"api"`
}

type apiConfig struct {
	Port string `yaml:"port"`
}

func main() {
	f, err := os.Open("user-service/configs/base.yaml")
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
				time.Sleep(1 * time.Second)
			}
		}
	}()
	defer registry.Deregister(ctx, instanceId)

}
