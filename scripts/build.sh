#!/bin/sh

set -e

echo "Building Docker images on the host, then loading them into Minikube..."

if ! command -v minikube >/dev/null 2>&1; then
  echo "minikube is required. Install it from https://minikube.sigs.k8s.io/docs/start/"
  exit 1
fi

if ! minikube status >/dev/null 2>&1; then
  echo "Minikube is not running. Start it first:"
  echo "  minikube start --memory=8192 --cpus=4"
  exit 1
fi

if ! command -v docker >/dev/null 2>&1; then
  echo "docker is required (Docker Desktop / host daemon)."
  exit 1
fi

# Do not use `minikube docker-env`. Builds inside Minikube often cannot reach
# Debian mirrors (apt-get: Could not resolve / connection timed out).
unset DOCKER_TLS_VERIFY DOCKER_HOST DOCKER_CERT_PATH MINIKUBE_ACTIVE_DOCKERD

IMAGES=""

build_go() {
  name="$1"
  pkg="$2"
  cgo="${3:-0}"
  echo "Building ${name}..."
  docker build \
    -t "kootha/${name}:latest" \
    -f docker/Dockerfile.go \
    --build-arg "SERVICE_PKG=${pkg}" \
    --build-arg "CGO=${cgo}" \
    .
  IMAGES="${IMAGES} kootha/${name}:latest"
}

# HTTP / gRPC services without Kafka (no CGO)
build_go api-gateway ./api-gateway/cmd 0
build_go user-service ./user-service/cmd 0
build_go id-generation-service ./id-generation-service/cmd 0
build_go quiz-bank-service ./quiz-bank-service/cmd 0

# Kafka-backed services (librdkafka / CGO)
build_go quiz-session-service ./quiz-session-service/cmd 1
build_go notification-service ./notification-service/cmd 1
build_go leaderboard-service ./leaderboard-service/cmd 1
build_go session-created-consumer ./quiz-session-service/cmd/session-created-consumer 1
build_go user-joined-consumer ./quiz-session-service/cmd/user-joined-consumer 1
build_go kafka-topics ./cmd/create-kafka-topics 1

echo "Building migrate job image..."
docker build -t kootha/migrate:latest -f docker/Dockerfile.migrate .
IMAGES="${IMAGES} kootha/migrate:latest"

echo "Building web frontend..."
docker build -t kootha/web:latest -f web/Dockerfile .
IMAGES="${IMAGES} kootha/web:latest"

echo "Loading images into Minikube..."
for img in ${IMAGES}; do
  echo "  ${img}"
  minikube image load "${img}"
done

echo "Docker images built and loaded into Minikube."
