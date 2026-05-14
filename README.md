# Lambare API

API REST en Go + MongoDB avec JWT et documentation Swagger.

Objectif de ce repo:
- un seul fichier Docker Compose;
- image API build/push via GitHub Actions;
- déploiement via Portainer en mode stack depuis Git.

## Démarrage rapide

```bash
docker compose up -d
curl http://localhost:8080/health
```

## Swagger (HTML)

<p>
  <a href="http://localhost:8080/swagger"><strong>Ouvrir Swagger UI</strong></a><br />
  <code>http://localhost:8080/swagger</code>
</p>

<p>
  Spécification OpenAPI:<br />
  <a href="http://localhost:8080/openapi.yaml">http://localhost:8080/openapi.yaml</a>
</p>

## Un seul Docker Compose

Le fichier unique est `docker-compose.yml`.

Le service `api` utilise une image distante (GHCR), pas un `build` local:

```yaml
api:
  image: ${API_IMAGE:-ghcr.io/lambare/go-mongo-api:latest}
  pull_policy: always
```

Donc, lors d'un `docker compose up -d` ou d'un redeploy Portainer, la stack tire l'image publiée par GitHub Actions.

## Comment est buildé le binaire Go

Le binaire est compilé dans le `Dockerfile` (multi-stage):

1. Stage `builder` basé sur `golang:1.22-alpine`
2. `go mod download`
3. `go build ... -o /app/api ./cmd/api`
4. Copie du binaire dans une image finale `scratch`

Important: pas besoin d'avoir Go installé sur ta machine pour déployer la stack.

## GitHub Actions (build image uniquement)

Workflow: `.github/workflows/docker-image.yml`

Déclenchement:
- push sur `main`
- lancement manuel (`workflow_dispatch`)

Ce workflow:
1. build l'image Docker multi-arch (`linux/amd64`, `linux/arm64`)
2. push vers GHCR (`ghcr.io/<owner>/<repo>`)
3. publie `latest` + un tag SHA court

## Déploiement Portainer depuis Git

Dans Portainer:
1. Créer une stack depuis repository Git.
2. Pointer vers ce repo et le fichier `docker-compose.yml`.
3. Renseigner les variables d'environnement (UI Portainer ou fichier `.env`).
4. Déployer.

Pour mettre à jour:
- relancer un redeploy/re-pull de la stack après un push sur `main`.

## Variables d'environnement principales

| Variable | Valeur par défaut | Description |
|---|---|---|
| `API_IMAGE` | `ghcr.io/lambare/go-mongo-api:latest` | Image API tirée par Compose |
| `API_BIND` | `8080` | Port hôte mappé vers `8080` dans le conteneur |
| `MONGO_BIND` | `27017` | Port hôte mappé vers Mongo |
| `MONGO_ROOT_USER` | `admin` | Utilisateur root Mongo |
| `MONGO_ROOT_PASSWORD` | `changeme` | Mot de passe root Mongo |
| `MONGO_DB_NAME` | `lambare` | Nom base Mongo |
| `JWT_SECRET` | `change-me-in-production` | Secret JWT |
| `GIN_MODE` | `release` | Mode Gin |
