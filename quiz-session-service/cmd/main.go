package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/richardktran/realtime-quiz/gen"
	"github.com/richardktran/realtime-quiz/pkg/discovery"
	"github.com/richardktran/realtime-quiz/pkg/discovery/consul"
	"github.com/richardktran/realtime-quiz/pkg/message-broker/kafka"
	idGenerationGateway "github.com/richardktran/realtime-quiz/quiz-session-service/internal/gateway/idgeneration"
	"github.com/richardktran/realtime-quiz/quiz-session-service/internal/gateway/socketio"
	grpcHandler "github.com/richardktran/realtime-quiz/quiz-session-service/internal/handler/grpc"
	httpHandler "github.com/richardktran/realtime-quiz/quiz-session-service/internal/handler/http"
	"github.com/richardktran/realtime-quiz/quiz-session-service/internal/repository/postgres"
	"github.com/richardktran/realtime-quiz/quiz-session-service/internal/service/quizsession"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gopkg.in/yaml.v3"
)

const serviceName = "quiz-session"

type serviceConfig struct {
	APIConfig apiConfig `yaml:"api"`
	DBConfig  dbConfig  `yaml:"db"`
}

type apiConfig struct {
	Port string `yaml:"port"`
}

type dbConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
}

// CORS middleware
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Upgrade, Connection")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "3600")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Handle WebSocket upgrade
		if r.Header.Get("Upgrade") == "websocket" {
			w.Header().Set("Upgrade", "websocket")
			w.Header().Set("Connection", "Upgrade")
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	f, err := os.Open("quiz-session-service/configs/base.yaml")

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

	// Service registry
	registry, err := consul.NewRegistry("localhost:8500")
	if err != nil {
		panic(err)
	}

	// Register the service
	instanceId := discovery.GenerateInstanceID(serviceName)
	ctx := context.Background()

	if err := registry.Register(ctx, instanceId, serviceName, fmt.Sprintf("localhost:%v", port)); err != nil {
		panic(err)
	}

	// Health check
	go func() {
		for {
			if err := registry.ReportHealthyState(instanceId, serviceName); err != nil {
				log.Printf("Failed to report healthy state: %v", err)
			}
			time.Sleep(5 * time.Second)
		}
	}()
	defer registry.Deregister(ctx, instanceId)

	// Kafka producer
	producer, err := kafka.NewProducer()
	if err != nil {
		panic(err)
	}
	defer producer.Close()

	idGenerationGateway := idGenerationGateway.New(registry)

	repo, err := postgres.New(postgres.Config{
		Host:     cfg.DBConfig.Host,
		Port:     cfg.DBConfig.Port,
		User:     cfg.DBConfig.User,
		Password: cfg.DBConfig.Password,
		DBName:   cfg.DBConfig.DBName,
	})

	if err != nil {
		panic(err)
	}

	svc := quizsession.New(repo, producer)
	h := grpcHandler.New(svc, idGenerationGateway)

	server := grpc.NewServer()
	reflection.Register(server)
	gen.RegisterQuizSessionServiceServer(server, h)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Start gRPC server in a goroutine
	go func() {
		if err := server.Serve(listener); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	// Initialize HTTP handlers
	roomHandler := httpHandler.NewRoomHandler(svc)

	// Create a new mux for HTTP handlers
	mux := http.NewServeMux()

	// Register routes with CORS middleware
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

	// Initialize WebSocket server
	wsServer := socketio.NewServer(svc)
	// Register WebSocket handler with CORS middleware
	mux.HandleFunc("/ws", corsMiddleware(http.HandlerFunc(wsServer.HandleWebSocket)).ServeHTTP)

	// Start HTTP server with CORS middleware
	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", corsMiddleware(mux)); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}
