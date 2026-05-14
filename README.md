# 🦁 Lambare API

API REST Go + MongoDB, sécurisée par JWT, documentée Swagger, déployée sur **lambare.fr** via Cloudflare Tunnel sur un ZimaBoard.

## Stack technique

| Composant | Choix |
|---|---|
| Language | Go 1.22 |
| Framework HTTP | Gin |
| Base de données | MongoDB 7 |
| Auth | JWT (HS256) avec access + refresh tokens |
| Documentation | OpenAPI 3.0 / Swagger UI |
| Conteneurisation | Docker + Docker Compose |
| CI/CD | GitHub Actions → GHCR |
| Tunnel | Cloudflare Tunnel |

---

## Démarrage rapide (local)

### Prérequis
- Docker & Docker Compose v2
- Go 1.22+ (optionnel, pour développer sans Docker)

```bash
# 1. Cloner le repo
git clone https://github.com/your-username/go-mongo-api.git
cd go-mongo-api

# 2. Copier et adapter les variables d'environnement
cp .env.example .env
# Éditez .env : changez les mots de passe et le JWT_SECRET

# 3. Lancer la stack
docker compose up -d

# 4. Vérifier
curl http://localhost:8080/health
```

Swagger UI disponible sur → **http://localhost:8080/swagger**

---

## Connexion MongoDB Compass (local)

```
mongodb://admin:changeme@localhost:27017/lambare?authSource=admin
```

> Le port 27017 est exposé sur l'hôte en développement (`docker-compose.yml`).  
> En production (`docker-compose.prod.yml`) il est restreint à `127.0.0.1` seulement.

---

## Structure du projet

```
go-mongo-api/
├── cmd/api/main.go              # Point d'entrée
├── internal/
│   ├── auth/jwt.go              # Génération & validation JWT
│   ├── config/config.go         # Chargement des variables d'env
│   ├── handlers/
│   │   ├── auth.go              # Register, Login, Refresh
│   │   ├── items.go             # CRUD items
│   │   └── swagger.go           # Serve OpenAPI spec + Swagger UI
│   ├── middleware/auth.go       # Middleware JWT + CORS
│   ├── models/models.go         # Structures de données
│   └── repository/
│       ├── item_repository.go   # Accès MongoDB items
│       └── user_repository.go   # Accès MongoDB users
├── docs/openapi.yaml            # Spec OpenAPI 3.0
├── scripts/mongo-init.js        # Init DB au premier démarrage
├── Dockerfile                   # Multi-stage build (scratch final)
├── docker-compose.yml           # Dev (Mongo port ouvert)
└── docker-compose.prod.yml      # Prod (ports restreints + cloudflared)
```

---

## Routes API

### Publiques
| Méthode | Route | Description |
|---|---|---|
| `GET` | `/health` | Health check |
| `GET` | `/swagger` | Swagger UI |
| `GET` | `/openapi.yaml` | Spec brute |
| `POST` | `/auth/register` | Créer un compte |
| `POST` | `/auth/login` | Se connecter |
| `POST` | `/auth/refresh` | Rafraîchir le token |

### Protégées (Bearer JWT requis)
| Méthode | Route | Description |
|---|---|---|
| `GET` | `/api/v1/items` | Lister (paginé, filtrable) |
| `POST` | `/api/v1/items` | Créer un item |
| `GET` | `/api/v1/items/:id` | Récupérer un item |
| `PUT` | `/api/v1/items/:id` | Mettre à jour |
| `DELETE` | `/api/v1/items/:id` | Supprimer |

### Paramètres de liste
```
GET /api/v1/items?page=1&page_size=20&search=foo&tags=go,api
```

---

## Authentification

```bash
# 1. Créer un compte
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@lambare.fr","password":"motdepasse"}'

# 2. Se connecter
TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@lambare.fr","password":"motdepasse"}' | jq -r .access_token)

# 3. Appel protégé
curl http://localhost:8080/api/v1/items \
  -H "Authorization: Bearer $TOKEN"
```

---

## CI/CD GitHub Actions

### Sur une Pull Request → `.github/workflows/ci.yml`
- Lint (`golangci-lint`)
- Tests avec MongoDB de service
- Build Docker (sans push)

### Sur merge dans `main` → `.github/workflows/cd.yml`
1. Build image multi-arch (`linux/amd64` + `linux/arm64`)
2. Push sur **GitHub Container Registry** (`ghcr.io`)
3. Deploy SSH sur le ZimaBoard

### Secrets GitHub à configurer
| Secret | Description |
|---|---|
| `ZIMABOARD_HOST` | IP ou hostname du ZimaBoard |
| `ZIMABOARD_USER` | Utilisateur SSH |
| `ZIMABOARD_SSH_KEY` | Clé privée SSH (ED25519 recommandé) |
| `ZIMABOARD_PORT` | Port SSH (défaut 22) |
| `GHCR_TOKEN` | Personal Access Token avec scope `read:packages` |

---

## Déploiement production sur ZimaBoard

### 1. Préparer le serveur

```bash
# Sur le ZimaBoard
sudo mkdir -p /opt/lambare
sudo chown $USER:$USER /opt/lambare
cd /opt/lambare

# Copier les fichiers de compose
scp docker-compose.prod.yml .env user@zimaboard:/opt/lambare/
```

### 2. Configurer le `.env` de production

```bash
# /opt/lambare/.env
MONGO_ROOT_USER=admin
MONGO_ROOT_PASSWORD=<mot_de_passe_fort>
MONGO_DB_NAME=lambare
JWT_SECRET=<secret_tres_long_et_aleatoire>
CLOUDFLARE_TUNNEL_TOKEN=<votre_token_cloudflare>
GITHUB_REPOSITORY=your-username/go-mongo-api
IMAGE_TAG=latest
```

### 3. Configurer Cloudflare Tunnel

Dans le dashboard Cloudflare Zero Trust :
```
api.lambare.fr  →  http://lambare-api:8080
```

> Le service `cloudflared` dans `docker-compose.prod.yml` se connecte automatiquement au tunnel existant via le token.

### 4. Premier démarrage

```bash
cd /opt/lambare
echo "$GHCR_TOKEN" | docker login ghcr.io -u your-username --password-stdin
docker compose -f docker-compose.prod.yml up -d
```

Les déploiements suivants se font automatiquement à chaque push sur `main`.

---

## Développement

```bash
# Lancer uniquement MongoDB en local
docker compose up mongo -d

# Lancer l'API en hot-reload avec air
go install github.com/air-verse/air@latest
air

# Lancer les tests
go test ./... -v

# Lint
golangci-lint run
```

---

## Variables d'environnement

| Variable | Défaut | Description |
|---|---|---|
| `API_PORT` | `8080` | Port d'écoute |
| `GIN_MODE` | `debug` | `debug` ou `release` |
| `MONGO_URI` | `mongodb://localhost:27017` | URI de connexion MongoDB |
| `MONGO_DB_NAME` | `lambare` | Nom de la base |
| `JWT_SECRET` | *(requis)* | Clé de signature JWT |
| `JWT_EXPIRY_MINUTES` | `60` | Durée de vie access token |
| `JWT_REFRESH_DAYS` | `7` | Durée de vie refresh token |
