package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/richardktran/realtime-quiz/api-gateway/internal/gateway/quizsession"
	userGateway "github.com/richardktran/realtime-quiz/api-gateway/internal/gateway/user"
	httpHandler "github.com/richardktran/realtime-quiz/api-gateway/internal/handler/http"
	"github.com/richardktran/realtime-quiz/pkg/config"
	"github.com/richardktran/realtime-quiz/pkg/cors"
	"github.com/richardktran/realtime-quiz/pkg/env"
	"github.com/richardktran/realtime-quiz/pkg/health"
	"github.com/richardktran/realtime-quiz/pkg/registry"
)

const serviceName = "api-gateway"

type serviceConfig struct {
	APIConfig apiConfig `yaml:"api"`
}

type apiConfig struct {
	Port int `yaml:"port"`
}

func main() {
	var cfg serviceConfig
	if err := config.LoadYAML("api-gateway/configs/base.yaml", &cfg); err != nil {
		panic(err)
	}

	port := env.OverlayInt(cfg.APIConfig.Port, "API_PORT")
	log.Printf("Starting the %s on port %v...", serviceName, port)

	reg, err := registry.New()
	if err != nil {
		panic(err)
	}

	userGW := userGateway.New(reg)
	quizSessionGW := quizsession.New(reg)

	userHandler := httpHandler.NewUserHandler(userGW)
	roomHandler := httpHandler.NewRoomHandler(quizSessionGW)

	mux := http.NewServeMux()
	health.MountHTTP(mux)

	mux.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			userHandler.CreateUser(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/rooms", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			roomHandler.CreateRoom(w, r)
		case http.MethodGet:
			roomHandler.GetRoom(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/rooms/join", roomHandler.JoinRoom)

	addr := fmt.Sprintf(":%s", strconv.Itoa(port))
	log.Printf("Listening on %s", addr)
	if err := http.ListenAndServe(addr, cors.Middleware(mux)); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}
