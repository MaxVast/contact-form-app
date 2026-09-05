# Contact Form App — Go + Vue.js + PostgreSQL + Kubernetes

Petite application de prise de contact, pensée comme un exercice
d'architecture microservices : un service API en Go, un front Vue.js,
une base PostgreSQL, le tout conteneurisé et déployable sur Kubernetes
via une pipeline CI/CD GitHub Actions.

## Architecture

```
[ Vue.js (Nginx) ] --/api--> [ Go API (chi) ] ---> [ PostgreSQL ]
   contact-frontend            contact-backend           postgres
```

- **contact-frontend** : SPA Vue 3 + Vite, servie en prod par Nginx qui
  fait aussi office de reverse-proxy vers le backend (`/api/*`).
- **contact-backend** : API REST en Go (framework `chi`), driver
  `pgx`, migration SQL exécutée au démarrage, arrêt propre sur
  `SIGTERM` (important pour les rolling updates Kubernetes).
- **postgres** : une seule table `contact_messages`.

Chaque brique a son propre `Dockerfile` et est déployée comme un
service Kubernetes indépendant : c'est cette séparation (front / API /
base, scalables et déployables indépendamment) qui constitue
l'architecture microservices ici — volontairement minimaliste pour un
formulaire de contact, mais le découpage (routeur HTTP, repository,
service métier) permet d'ajouter facilement un vrai microservice
supplémentaire (ex: un `notification-service` qui consomme une file
de messages pour envoyer l'email) sans toucher au reste.

## Lancer en local (Docker Compose)

```bash
docker compose up --build
```

- Frontend : http://localhost:8081
- API : http://localhost:8080/api/contact
- Postgres : localhost:5432 (contact / contact / contact_db)

Pour développer le front avec hot-reload (proxy vers le backend Docker) :

```bash
cd frontend
npm install
npm run dev
```

## Tests

Backend (tests unitaires : validation du modèle, service avec un
repository mocké, handlers HTTP avec `httptest`) :

```bash
task install   # docker-compose up -d --build
task ci        # tests backend + frontend, couverture dans reports/ et à la fin arrête et nettoie les conteneurs
```
Ou directement, sans Task — backend (tests unitaires : validation du modèle, service avec un repository mocké, handlers HTTP avec httptest) :

```bash
cd backend
go test ./... -v -cover
```
Frontend (Vitest + Vue Test Utils, `fetch` mocké) :

```bash
cd frontend
npm install
npm run test
```

Les deux sont exécutés automatiquement dans la CI (`test-backend` et
`test-frontend`) avant tout build/push d'image.

Trois workflows GitHub Actions sont fournis :

`.github/workflows/build-and-test.yml` — sur chaque PR vers `develop` : Build de l'application (valide que les Dockerfiles buildent correctement), et lancement de  l'application (valide que l'application fonctionne correctement),
vérification des principaux endpoints, et pour finir execution des tests avec: task ci (tests backend + frontend). Volontairement minimal : nos tests sont unitaires avec mocks, donc pas besoin de démarrer la stack Docker pour les faire tourner — contrairement à un projet où les tests s'exécutent contre l'app et sa base réellement démarrées.

`.github/workflows/validate-k8s.yml` — sur chaque PR vers `develop` : La configuration Helm est lint et tester, ensuite un smoke test et  réaliser avant tout déploiement, dans le second jobs l'application est déployé avec Helm (Backend, Frontend & DB) sur kubernetes, et ensuite vérifié que l'application fonctionne correctement.

`.github/workflows/build-and-push.yml` — sur push vers `main` : build/push des images Docker vers GHCR.

## Générer go.sum

Le `go.sum` n'est pas fourni (pas d'accès au proxy Go dans cet
environnement). Avant le premier build :

```bash
cd backend
go mod tidy
```

## Node.js requis (front)

Le front nécessite Node ≥20.19 (idéalement 22, comme dans le
Dockerfile). Avec une version plus ancienne ou un paquet système
générique (Ubuntu/WSL), Vite/Vitest peut échouer avec une erreur du
type `crypto.getRandomValues is not a function`. Utilise `nvm` :

```bash
cd frontend
nvm install
nvm use 22.22
```

! Note importe le projet comporte des containers Docker avec les versions exact de Goland et de Node.js !

Vous pouvez utiliser ces containers pour executer build, ci, format et etc

## Déployer sur Kubernetes
Deux façons équivalentes de déployer backend + frontend + Ingress :
manifests YAML bruts (`k8s/backend/`, `k8s/frontend/`, `k8s/ingress.yaml`, `k8s/postgres/`)
ou chart Helm (`helm/contact-form-app/`) et une chart Helm pour Postgres (`helm/database`).
--------------------
Pense à :
- remplacer `ghcr.io/OWNER/...` dans `k8s/backend/deployment.yaml` et
  `k8s/frontend/deployment.yaml` par ton propre repo GHCR ;
- changer le mot de passe Postgres dans `k8s/postgres/secret.yaml` et
  `k8s/backend/secret.yaml` (idéalement via un outil comme Sealed
  Secrets ou un secret manager, pas en clair dans le repo) ;
- adapter `host: contact.example.com` dans `k8s/ingress.yaml`.

## Option A - sans Helm (manifests bruts)

# 1. Démarrer minikube et builder les images dans son environnement
```bash
minikube start
eval $(minikube docker-env)

minikube image build -t contact-backend:local ./backend
minikube image build -t contact-frontend:local ./frontend
```
# 2. Déploiement des manifests
```bash
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/postgres/
kubectl apply -f k8s/backend/
kubectl apply -f k8s/frontend/
kubectl apply -f k8s/ingress.yaml
```

# 3. Créer le namespace et les secrets, puis déployer la base de données
```bash
kubectl create namespace contact-form
kubectl -n contact-form create secret generic postgres-secret \
--from-literal=POSTGRES_USER=contact \
--from-literal=POSTGRES_PASSWORD=<mot_de_passe> \
--from-literal=POSTGRES_DB=contact_db
kubectl -n contact-form create secret generic backend-secret \
--from-literal=DATABASE_URL="postgres://contact:<mot_de_passe>@postgres:5432/contact_db?sslmode=disable"
kubectl apply -n contact-form -f k8s/postgres/statefulset.yaml -f k8s/postgres/service.yaml
```

# 4. Pointage sur les images locales (aucun pull réseau)
```bash
kubectl -n contact-form set image deployment/contact-backend contact-backend=contact-backend:local
kubectl -n contact-form set image deployment/contact-frontend contact-frontend=contact-frontend:local
kubectl -n contact-form patch deployment contact-backend --type='json' \
  -p='[{"op":"add","path":"/spec/template/spec/containers/0/imagePullPolicy","value":"Never"}]' 2>/dev/null || true
kubectl -n contact-form patch deployment contact-frontend --type='json' \
  -p='[{"op":"add","path":"/spec/template/spec/containers/0/imagePullPolicy","value":"Never"}]' 2>/dev/null || true
```

# 5. Attente que tout soit prêt (timeout 2 min)
```bash
kubectl -n contact-form rollout status deployment/contact-backend --timeout=120s
kubectl -n contact-form rollout status deployment/contact-frontend --timeout=120s
```

## Option B - avec Helm
# 1. Namespace + secrets (noms attendus par les fichiers du repo ; adapte-les si tu les as personnalisés dans ton values.yaml)
```bash
kubectl create namespace contact-form

kubectl -n contact-form create secret generic postgres-secret-form-app \
  --from-literal=POSTGRES_USER=contact \
  --from-literal=POSTGRES_PASSWORD=devlocal \
  --from-literal=POSTGRES_DB=contact_db

kubectl -n contact-form create secret generic backend-form-app-secret \
  --from-literal=DATABASE_URL="postgres://contact:devlocal@postgres:5432/contact_db?sslmode=disable"
```
# 2. Base de données
```bash
kubectl apply -f helm/database/postgres.yaml
```
# 3. Backend + frontend, avec les images locales
```bash
helm upgrade --install contact-form ./helm/contact-form-app \
  --namespace contact-form \
  --set backend.image.repository=contact-backend \
  --set backend.image.tag=local \
  --set backend.image.pullPolicy=Never \
  --set backend.existingSecret=backend-form-app-secret \
  --set frontend.image.repository=contact-frontend \
  --set frontend.image.tag=local \
  --set frontend.image.pullPolicy=Never
```
# Accéder à l'application (port-forward)
**Le plus simple — port-forward direct** :
```bash
kubectl -n contact-form port-forward service/contact-frontend 8080:8080
```
puis ouvrir http://localhost:8080/


**Plus proche de la prod — via l'Ingress**, pour tester le chemin
complet navigateur → Ingress → Service → Pod :
```bash
minikube addons enable ingress
minikube tunnel   # dans un terminal séparé, laisse tourner (demande sudo)
```
Dans un autre terminal, vérifie que l'Ingress a bien une adresse :
```bash
kubectl get ingress -n contact-form
```
Puis ajoute l'entrée DNS locale, avec le host défini dans
`helm/contact-form-app/values.yaml` (`contact.example.com` par défaut) :
```bash
echo "$(minikube ip) contact.example.com" | sudo tee -a /etc/hosts
```

> **Sous WSL** : si tu ouvres le navigateur depuis **Windows** et non
> depuis WSL, l'`/etc/hosts` modifié ci-dessus (celui de WSL) est
> invisible pour Windows — tu auras `DNS_PROBE_FINISHED_NXDOMAIN`.
> Ajoute la même ligne dans
> `C:\Windows\System32\drivers\etc\hosts` (Bloc-notes en admin), ou
> vérifie d'abord sans toucher aux hosts avec :
> ```bash
> curl -H "Host: contact.example.com" http://$(minikube ip)/
> ```
> Si ça renvoie le HTML de la page, l'Ingress fonctionne — le souci
> n'est alors que la résolution DNS côté navigateur.

`minikube image build` fonctionne quel que soit le runtime de minikube
(docker ou containerd) — vérifiable avec `minikube profile list`.

Pour tout supprimer et repartir de zéro :
```bash
# avec Helm
helm uninstall contact-form --namespace contact-form
# sans Helm
kubectl delete -f k8s/backend/ -f k8s/frontend/ -f k8s/ingress.yaml
# dans tous les cas
kubectl delete namespace contact-form
```

## CI/CD

`.github/workflows/build-and-test.yml` :
1. Sur chaque PR, l'application est build;
2. L'application est lancé et testé par différent endpoints:
   1. Pour le back-end : `/health`, `/ready`;
   2. Pour le front-end: `/`;
3. Les lint, format et tests sont lancé
4. Tous les conteneurs sont arrêté

`.github/workflows/validate-k8s.yml` :
1. Sur chaque PR, l'application est lint et test avec Helm et kubeconform;
2. Un deploiement est réalisé en structure smoke test
3. Les images sont build
4. Postgres est déployé avec les charts Helm
5. L'application est déployé avec les charts Helm
6. Redémarrage des pods
7. L'application est lancé et testé par différent endpoints:
    1. Pour le back-end : `/health`, `/ready`;
    2. Pour le front-end: `/`;

`.github/workflows/build-and-push.yml` :
1. Connexion à ghcr.io
2. Build and push backend image sur ghcr.io
3. Build and push frontend image sur ghcr.io

## Pistes d'évolution
- [ ] workflow github pour déploiement sur GCP/GKE
- [ ] ajout d'un outil d'observabilité comme Prometheus & Grafana
- [ ] tests d'intégration (vraie base Postgres via testcontainers) ;
- [ ] rate limiting sur `/api/contact` (anti-spam) ;
- [ ] notification par email en découplant via une file (NATS/RabbitMQ) et
  un second microservice.
- [ ] ajout d'un model user
- [ ] ajout d'une page de connexion
- [ ] ajout d'une connexion avec JWT
- [ ] ajout d'un back-office
- [ ] changer le (framework `chi`) par `gin`
- [ ] ajouter un orm (exemple `gorm`)
