#!/usr/bin/env bash
echo "==> Check and Start minikube"
minikube status > /dev/null || {
  echo "minikube n'est pas démarré. 'minikube start' d'abord.";
  minikube start
  eval $(minikube docker-env)
}

echo""
echo "===> Minikube Ok"
minikube status

echo""
echo "==> Build des images"
minikube image build -t contact-backend:local ./backend
echo""
minikube image build -t contact-frontend:local ./frontend

echo""
echo "==> Creation du namespace"
kubectl create namespace contact-form

echo""
echo "==> Creation des secrets pour la BDD"
kubectl -n contact-form create secret generic postgres-secret-form-app \
  --from-literal=POSTGRES_USER=contact \
  --from-literal=POSTGRES_PASSWORD=devlocal \
  --from-literal=POSTGRES_DB=contact_db

kubectl -n contact-form create secret generic backend-form-app-secret \
  --from-literal=DATABASE_URL="postgres://contact:devlocal@postgres:5432/contact_db?sslmode=disable"

echo""
echo "==> Application des secrets pour la BDD"
kubectl apply -f helm/database/postgres.yaml

echo""
echo "==> Backend + frontend, avec les valeurs locales"
helm upgrade --install contact-form-app ./helm/contact-form-app \
  --namespace contact-form \
  -f helm/contact-form-app/values-local.yaml

set -e
echo""
echo "==> Accéder aux applications (port-forward)"
kubectl -n contact-form port-forward service/contact-frontend 8080:8080 &
FRONTEND_PID=$!

kubectl -n contact-form port-forward service/contact-backend 8081:8080 &
BACKEND_PID=$!

trap 'kill $FRONTEND_PID $BACKEND_PID 2>/dev/null' EXIT INT TERM

wait
