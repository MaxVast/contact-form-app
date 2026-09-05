# Observabilité — Prometheus + Grafana

Guide dédié à la stack d'observabilité de contact-form-app : ce qui est
installé, comment l'activer, construire un graphique, poser une
alerte, et où en sont les logs.

---

## 1. Comment ça marche (avant de cliquer sur quoi que ce soit)

- **Prometheus** est une base de données de séries temporelles. Il
  fonctionne en mode *pull* : toutes les 15 secondes, il va chercher
  lui-même les métriques sur `GET /metrics` de chaque pod backend, et
  stocke chaque valeur horodatée.
- **Le backend Go expose `/metrics`** au format texte Prometheus
  (`internal/observability/metrics.go`) : nombre de requêtes HTTP par
  méthode/route/statut (`http_requests_total`) et leur durée
  (`http_request_duration_seconds`), plus les métriques standard du
  runtime Go (mémoire, goroutines, GC).
- **Le `ServiceMonitor`** (`templates/backend-servicemonitor.yaml`)
  est ce qui dit à Prometheus où scraper : il sélectionne le Service
  Kubernetes `contact-backend` via le label `app: contact-backend`.
  C'est le Prometheus Operator (installé avec `kube-prometheus-stack`)
  qui traduit ce CRD en configuration Prometheus réelle.
- **Grafana ne stocke aucune donnée** — c'est uniquement une couche de
  visualisation qui interroge Prometheus et affiche le résultat.

---

## 2. Installation (une fois par cluster)

```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

helm install prometheus prometheus-community/kube-prometheus-stack \
  --namespace monitoring --create-namespace
```

Vérifier que tout est démarré (peut prendre 1-3 minutes) :
```bash
kubectl get pods -n monitoring
```

Récupérer le mot de passe Grafana (généré aléatoirement, pas
`prom-operator` par défaut selon la version du chart) :
```bash
kubectl --namespace monitoring get secrets prometheus-grafana \
  -o jsonpath="{.data.admin-password}" | base64 -d ; echo
```

---

## 3. Activer le monitoring de l'application

Toutes les valeurs de déploiement local (images, secrets, et
`metrics.enabled: true`) sont regroupées dans
`helm/contact-form-app/values-local.yaml` — voir ce fichier plutôt que
d'empiler des `--set` à la main :

```bash
helm upgrade contact-form-app ./helm/contact-form-app \
  --namespace contact-form \
  -f helm/contact-form-app/values-local.yaml
```

Vérifier que la cible est bien scrapée avec succès :
```bash
kubectl -n monitoring port-forward svc/prometheus-kube-prometheus-prometheus 9090:9090
```
Ouvrir `http://localhost:9090/targets`, chercher
`contact-form/contact-backend` → doit afficher `UP`.

---

## 4. Accéder à Grafana

```bash
kubectl -n monitoring port-forward svc/prometheus-grafana 3000:80
```
Ouvrir `http://localhost:3000`, se connecter avec `admin` / le mot de
passe récupéré à l'étape 2.

### Importer un dashboard tout fait

**Dashboards → New → Import**, entrer l'ID `6671` (dashboard "Go
Processes"), **Load**, sélectionner la source Prometheus proposée par
défaut, **Import**.

---

## 5. Construire son propre graphique

1. **Dashboards → New → New Dashboard → Add visualization**
2. Choisir la source de données **Prometheus**
3. Dans l'éditeur de requête, deux modes :
    - **Builder** : menus déroulants (`Select metric`, `+ Operations`)
      — pratique pour découvrir les métriques disponibles sans
      connaître le nom exact.
    - **Code** (onglet à côté de "Builder") : champ texte pour taper du
      PromQL directement — plus rapide une fois qu'on connaît les
      requêtes qu'on veut.
4. Taper la requête puis **Run queries**.

### Requêtes PromQL utiles pour ce projet

| Ce que ça montre | Requête |
|---|---|
| Requêtes/seconde, toutes routes | `rate(http_requests_total[5m])` |
| Soumissions du formulaire de contact | `rate(http_requests_total{path="/api/contact/", method="POST"}[5m])` |
| Taux d'erreur (5xx / total) | `sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m]))` |
| Latence au 95e percentile | `histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))` |
| Nombre de goroutines (fuite mémoire ?) | `go_goroutines{job="contact-backend"}` |
| Mémoire utilisée par le process | `process_resident_memory_bytes{job="contact-backend"}` |

> **Repère utile** : les "quatre signaux dorés" (Google SRE) —
> latence, trafic, erreurs, saturation — suffisent à avoir une vue de
> santé correcte de n'importe quel service. Les trois premiers sont
> couverts par les requêtes ci-dessus ; la saturation (CPU/mémoire vs
> les `resources.limits` du pod) vient de `kube-state-metrics`, déjà
> installé avec la stack.

5. Une fois satisfait du graphique : **Apply** (en haut à droite du
   panel), puis **Save dashboard** (en haut du tableau de bord).

---

## 6. Poser une alerte

**Alerting → Alert rules → New alert rule** :

1. Nommer la règle (ex: `Taux d'erreur élevé`)
2. Définir la requête (reprendre celle du taux d'erreur ci-dessus)
3. Définir la condition : ex. "la valeur est au-dessus de `0.05`" (5%)
4. Définir la durée pendant laquelle la condition doit être vraie
   avant de déclencher (ex: `5m`) — évite les fausses alertes sur un
   pic d'une seconde
5. Associer un point de contact (**Alerting → Contact points**) : email,
   Slack, webhook... à configurer une fois, réutilisable pour toutes
   les règles

Sans ça, un dashboard n'a qu'une valeur limitée : il faut le regarder
activement pour être utile. Une alerte inverse la logique — c'est elle
qui vient te chercher.

---

[//]: # (## 7. Et les logs ?)

[//]: # ()
[//]: # (`kube-prometheus-stack` ne couvre que les **métriques**, pas les logs.)

[//]: # (Actuellement, les logs de l'application restent accessibles seulement)

[//]: # (via :)

[//]: # (```bash)

[//]: # (kubectl -n contact-form logs deployment/contact-backend)

[//]: # (kubectl -n contact-form logs deployment/contact-backend --previous  # après un crash)

[//]: # (```)

[//]: # ()
[//]: # (Pour les voir **dans Grafana**, au même endroit que les graphiques)

[//]: # (&#40;utile pour corréler "pic d'erreurs à 14h32" avec "qu'est-ce qui)

[//]: # (apparaît dans les logs à ce moment-là"&#41;, il faudrait ajouter **Loki**)

[//]: # (&#40;l'équivalent de Prometheus, mais pour les logs — même éditeur,)

[//]: # (Grafana, mêmes créateurs&#41; :)

[//]: # ()
[//]: # (```bash)

[//]: # (helm repo add grafana https://grafana.github.io/helm-charts)

[//]: # (helm repo update)

[//]: # ()
[//]: # (helm install loki grafana/loki-stack )

[//]: # (  --namespace monitoring )

[//]: # (  --set grafana.enabled=false )

[//]: # (  --set promtail.enabled=true)

[//]: # (```)

[//]: # ()
[//]: # (&#40;`grafana.enabled=false` car on a déjà Grafana via)

[//]: # (`kube-prometheus-stack` — `loki-stack` ajoute juste Loki + Promtail,)

[//]: # (qui collecte les logs de tous les pods automatiquement&#41;)

[//]: # ()
[//]: # (Non fait pour l'instant — piste pour plus tard, pas nécessaire pour un)

[//]: # (premier niveau d'observabilité fonctionnel.)

---

## Dépannage — pièges rencontrés en pratique

| Symptôme | Cause |
|---|---|
| `ServiceMonitor` créé mais absent de `/targets` dans Prometheus | Labels du `ServiceMonitor` ne matchent pas `serviceMonitorSelector` de l'objet `Prometheus` (vérifier `kubectl -n monitoring get prometheus -o yaml \| grep -A3 serviceMonitorSelector`) |
| Cible visible mais `0/0 up`, "No active targets" | Le **Service** Kubernetes lui-même n'a pas le label `app: contact-backend` dans ses `metadata.labels` (différent du `spec.selector`, qui lui sélectionne les pods) |
| Cible en `UNKNOWN` / `Last scrape: never` | Souvent juste le temps du premier cycle de scrape (15-30s) — rafraîchir la page avant de creuser plus loin |
| `helm upgrade --reuse-values` ignore une nouvelle valeur ajoutée à `values.yaml` | `--reuse-values` reprend les valeurs de la **release précédente**, pas le fichier local — utiliser `-f values-local.yaml` à la place, relu en entier à chaque fois |
| `panic: chi: all middlewares must be defined before routes on a mux` | Une route (ex: `/metrics`) déclarée avant un `r.Use(...)` sur le même routeur — soit remonter tous les `Use()` avant les routes, soit isoler les routes non-instrumentées dans un `r.Group(...)` séparé |
| 404 sur `/metrics` alors que le code semble correct | L'image du pod n'a pas été reconstruite après modification du code (`minikube image build` + `kubectl rollout restart` nécessaires — un pod ne recharge jamais son code tout seul) |