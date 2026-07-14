package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/richardktran/realtime-quiz/api-gateway/internal/gateway/quizsession"
	userGateway "github.com/richardktran/realtime-quiz/api-gateway/internal/gateway/user"
	httpHandler "github.com/richardktran/realtime-quiz/api-gateway/internal/handler/http"
	"github.com/richardktran/realtime-quiz/pkg/discovery/consul"
	"gopkg.in/yaml.v3"
)

const serviceName = "api-gateway"

type serviceConfig struct {
	APIConfig apiConfig `yaml:"api"`
}

type apiConfig struct {
	Port int `yaml:"port"`
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "3600")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	f, err := os.Open("api-gateway/configs/base.yaml")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var cfg serviceConfig
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		panic(err)
	}

	port := cfg.APIConfig.Port
	log.Printf("Starting the %s on port %v...", serviceName, port)

	registry, err := consul.NewRegistry("localhost:8500")
	if err != nil {
		panic(err)
	}

	userGW := userGateway.New(registry)
	quizSessionGW := quizsession.New(registry)

	userHandler := httpHandler.NewUserHandler(userGW)
	roomHandler := httpHandler.NewRoomHandler(quizSessionGW)

	mux := http.NewServeMux()

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

	addr := fmt.Sprintf(":%v", port)
	log.Printf("Listening on %s", addr)
	if err := http.ListenAndServe(addr, corsMiddleware(mux)); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}
