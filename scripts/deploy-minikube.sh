#!/usr/bin/env bash
# Déploie l'application sur un minikube déjà démarré (minikube start).
# Regroupe toutes les étapes nécessaires : build des images directement
# dans le node minikube (indépendant du runtime docker/containerd),
# déploiement des manifests, puis forçage sur les images locales sans
# pull réseau (imagePullPolicy: Never).
set -euo pipefail

echo "==> Vérification de minikube"
minikube status > /dev/null || { echo "minikube n'est pas démarré. Lance 'minikube start' d'abord."; exit 1; }

echo "==> Build des images directement dans minikube"
minikube image build -t contact-backend:local ../backend
minikube image build -t contact-frontend:local ../frontend

echo "==> Déploiement des manifests"
kubectl apply -f ../k8s/namespace.yaml
kubectl apply -f ../k8s/postgres/
kubectl apply -f ../k8s/backend/
kubectl apply -f ../k8s/frontend/

echo "==> Pointage sur les images locales (aucun pull réseau)"
kubectl -n contact-form set image deployment/contact-backend contact-backend=contact-backend:local
kubectl -n contact-form set image deployment/contact-frontend contact-frontend=contact-frontend:local
kubectl -n contact-form patch deployment contact-backend --type='json' \
  -p='[{"op":"add","path":"/spec/template/spec/containers/0/imagePullPolicy","value":"Never"}]' 2>/dev/null || true
kubectl -n contact-form patch deployment contact-frontend --type='json' \
  -p='[{"op":"add","path":"/spec/template/spec/containers/0/imagePullPolicy","value":"Never"}]' 2>/dev/null || true

echo "==> Attente que tout soit prêt (timeout 2 min)"
kubectl -n contact-form rollout status deployment/contact-backend --timeout=120s
kubectl -n contact-form rollout status deployment/contact-frontend --timeout=120s

echo ""
echo "C'est prêt. Dans un autre terminal :"
echo "  kubectl -n contact-form port-forward service/contact-frontend 8080:8080"
echo "puis ouvre http://localhost:8080/"
