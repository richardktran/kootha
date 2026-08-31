#!/bin/sh

set -e

NAMESPACE="${NAMESPACE:-kootha}"
K8S_DIR="k8s/dev"

echo "Deploying Kootha to Kubernetes (${NAMESPACE})..."

kubectl apply -f "${K8S_DIR}/namespace.yaml"
sleep 2
kubectl apply -f "${K8S_DIR}/secrets.yaml"

# Jobs are immutable; delete so re-deploys can recreate them.
kubectl delete job migrate kafka-topics -n "${NAMESPACE}" --ignore-not-found

echo "Applying infrastructure..."
kubectl apply -f "${K8S_DIR}/postgres.yaml"
kubectl apply -f "${K8S_DIR}/redis.yaml"
kubectl apply -f "${K8S_DIR}/apache-kafka.yaml"
kubectl apply -f "${K8S_DIR}/jobs.yaml"

echo "Waiting for infrastructure..."
kubectl wait --for=condition=ready pod -l app=postgres -n "${NAMESPACE}" --timeout=180s
kubectl wait --for=condition=available deployment/redis -n "${NAMESPACE}" --timeout=120s
kubectl wait --for=condition=ready pod -l app=apache-kafka -n "${NAMESPACE}" --timeout=180s

echo "Waiting for migrations and Kafka topics..."
kubectl wait --for=condition=complete job/migrate -n "${NAMESPACE}" --timeout=180s
kubectl wait --for=condition=complete job/kafka-topics -n "${NAMESPACE}" --timeout=180s

echo "Applying application services..."
kubectl apply -f "${K8S_DIR}/id-generation-service.yaml"
kubectl apply -f "${K8S_DIR}/user-service.yaml"
kubectl apply -f "${K8S_DIR}/quiz-bank-service.yaml"
kubectl apply -f "${K8S_DIR}/quiz-session-service.yaml"
kubectl apply -f "${K8S_DIR}/api-gateway.yaml"
kubectl apply -f "${K8S_DIR}/notification-service.yaml"
kubectl apply -f "${K8S_DIR}/leaderboard-service.yaml"
kubectl apply -f "${K8S_DIR}/session-created-consumer.yaml"
kubectl apply -f "${K8S_DIR}/user-joined-consumer.yaml"
kubectl apply -f "${K8S_DIR}/web.yaml"

echo "Waiting for application deployments..."
for deploy in \
  api-gateway \
  user-service \
  id-generation-service \
  quiz-bank-service \
  quiz-session-service \
  notification-service \
  leaderboard-service \
  session-created-consumer \
  user-joined-consumer \
  web
do
  kubectl wait --for=condition=available "deployment/${deploy}" -n "${NAMESPACE}" --timeout=300s
done

echo ""
echo "Deployment finished."
echo ""
echo "Next steps:"
echo "  1. In another terminal:  minikube tunnel"
echo "  2. Open the web app:     http://localhost:3000"
echo "  3. Cluster status:       make k8s-status"
echo ""
echo "If LoadBalancer IPs stay pending, port-forward instead:"
echo "  kubectl port-forward -n ${NAMESPACE} svc/web 3000:3000"
echo "  kubectl port-forward -n ${NAMESPACE} svc/api-gateway 8080:8080"
echo "  kubectl port-forward -n ${NAMESPACE} svc/notification-service 8086:8086"
