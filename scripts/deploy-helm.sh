#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="contact-form"

echo "==> Check and Start minikube"
if ! minikube status > /dev/null 2>&1; then
  echo "minikube n'est pas démarré. Démarrage..."
  minikube start
fi

echo ""
echo "===> Minikube Ok"
minikube status

echo ""
echo "==> Build des images"
echo ""
echo "==> Build image backend"
minikube image build -t contact-backend:local ./backend
echo ""
echo "==> Build image frontend"
minikube image build -t contact-frontend:local ./frontend

echo ""
echo "==> Creation du namespace (idempotent)"
kubectl get namespace "$NAMESPACE" > /dev/null 2>&1 || kubectl create namespace "$NAMESPACE"

echo ""
echo "==> Creation des secrets pour la BDD (idempotent)"
kubectl -n "$NAMESPACE" create secret generic postgres-secret-form-app \
  --from-literal=POSTGRES_USER=contact \
  --from-literal=POSTGRES_PASSWORD=devlocal \
  --from-literal=POSTGRES_DB=contact_db \
  --dry-run=client -o yaml | kubectl apply -f -

kubectl -n "$NAMESPACE" create secret generic backend-form-app-secret \
  --from-literal=DATABASE_URL="postgres://contact:devlocal@postgres:5432/contact_db?sslmode=disable" \
  --dry-run=client -o yaml | kubectl apply -f -

echo ""
echo "==> Creation du secret JWT"
kubectl -n "$NAMESPACE" create secret generic jwt-secret \
  --from-literal=JWT_SECRET="XGeNrs2NhavFOcKTht6HnfWwvmDyiCBm81yr3Rejdgq" \
  --from-literal=JWT_EXPIRATION="1h" \
  --dry-run=client -o yaml | kubectl apply -f -

echo ""
echo "==> Creation du compte admin (local)"
if ! kubectl -n "$NAMESPACE" get secret admin-secret > /dev/null 2>&1; then
  read -rp "Email admin: " ADMIN_EMAIL
  read -rsp "Mot de passe admin: " ADMIN_PWD
  echo ""
  kubectl -n "$NAMESPACE" create secret generic admin-secret \
    --from-literal=ADMIN_EMAIL="$ADMIN_EMAIL" \
    --from-literal=ADMIN_PASSWORD="$ADMIN_PWD"
  unset ADMIN_PWD
else
  echo "admin-secret existe déjà, aucune modification."
fi

echo ""
echo "==> Application des secrets pour la BDD"
kubectl apply -n "$NAMESPACE" -f helm/database/postgres.yaml

echo ""
echo "==> Attente de PostgreSQL"
kubectl -n "$NAMESPACE" rollout status deployment/postgres --timeout=120s

echo ""
echo "==> Déploiement Helm"
echo "    Chart      : ./helm/contact-form-app"
echo "    Namespace  : $NAMESPACE"
echo "    Values     : values-local.yaml"
echo ""

helm upgrade --install contact-form-app ./helm/contact-form-app \
  --namespace "$NAMESPACE" \
  -f helm/contact-form-app/values-local.yaml \
  --wait \
  --timeout 2m

echo ""
echo "==> Helm deployment terminé"
helm status contact-form-app --namespace "$NAMESPACE"

echo ""
echo "==> Rebuild Pod Backend"
kubectl -n "$NAMESPACE" rollout restart deployment/contact-backend
kubectl -n "$NAMESPACE" rollout status deployment/contact-backend

echo ""
echo "==> Rebuild Pod Frontend"
kubectl -n "$NAMESPACE" rollout restart deployment/contact-frontend
kubectl -n "$NAMESPACE" rollout status deployment/contact-frontend

echo ""
echo "==> Vérification des ressources"
kubectl get pods -n "$NAMESPACE"
kubectl get services -n "$NAMESPACE"
kubectl get jobs -n "$NAMESPACE"

echo ""
echo "==> Accéder aux applications (port-forward)"
kubectl -n "$NAMESPACE" port-forward service/contact-frontend 8080:8080 &
FRONTEND_PID=$!

kubectl -n "$NAMESPACE" port-forward service/contact-backend 8081:8080 &
BACKEND_PID=$!

trap 'kill $FRONTEND_PID $BACKEND_PID 2>/dev/null' EXIT INT TERM

wait