package health

import (
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

// Handler returns 200 OK for Kubernetes HTTP probes.
func Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}
}

// RegisterGRPC attaches the standard gRPC health service so kubelet can probe with grpc.
func RegisterGRPC(s *grpc.Server) *health.Server {
	h := health.NewServer()
	healthpb.RegisterHealthServer(s, h)
	h.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	return h
}

// MountHTTP registers /health and /ready on mux.
func MountHTTP(mux *http.ServeMux) {
	mux.HandleFunc("/health", Handler())
	mux.HandleFunc("/ready", Handler())
}
