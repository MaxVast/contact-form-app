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
task ci        # tests backend + frontend, couverture dans reports/
task down      # arrête et nettoie les conteneurs
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

Deux workflows GitHub Actions sont fournis :

`.github/workflows/build-and-test.yml` — sur chaque PR vers `main` : task ci (tests backend + frontend, avec couverture archivée en artifact). Volontairement minimal : nos tests sont unitaires avec mocks, donc pas besoin de démarrer la stack Docker pour les faire tourner — contrairement à un projet où les tests s'exécutent contre l'app et sa base réellement démarrées.
`.github/workflows/ci-cd.yml` — sur push vers `main` : build/push des images Docker vers GHCR puis déploiement Kubernetes. C'est ce job qui valide que les Dockerfiles buildent correctement (pas vérifié sur les PR, par choix de rapidité).

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
nvm install   # lit frontend/.nvmrc
nvm use 22.22
```


## Déployer sur Kubernetes

```bash
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/postgres/
kubectl apply -f k8s/backend/
kubectl apply -f k8s/frontend/
kubectl apply -f k8s/ingress.yaml
```

Pense à :
- remplacer `ghcr.io/OWNER/...` dans `k8s/backend/deployment.yaml` et
  `k8s/frontend/deployment.yaml` par ton propre repo GHCR ;
- changer le mot de passe Postgres dans `k8s/postgres/secret.yaml` et
  `k8s/backend/secret.yaml` (idéalement via un outil comme Sealed
  Secrets ou un secret manager, pas en clair dans le repo) ;
- adapter `host: contact.example.com` dans `k8s/ingress.yaml`.

## CI/CD

`.github/workflows/ci-cd.yml` :
1. `go vet` + `go build` sur chaque push/PR ;
2. sur `main` : build & push des deux images Docker vers GHCR ;
3. déploiement sur le cluster via `kubectl` (nécessite un secret de
   repo `KUBE_CONFIG`, même principe que pour le cluster GKE de
   shop-menace-kult).

## Pistes d'évolution

- tests d'intégration (vraie base Postgres via testcontainers) ;
- Helm chart au lieu des manifests YAML bruts, pour templatiser
  images/hosts/secrets entre environnements ;
- rate limiting sur `/api/contact` (anti-spam) ;
- notification par email en découplant via une file (NATS/RabbitMQ) et
  un second microservice.
