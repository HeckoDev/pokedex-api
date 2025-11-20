# Pokédex API - Déploiement avec Docker

Guide de déploiement de l'API Pokédex avec PostgreSQL.

## 🚀 Démarrage rapide

### 1. Configuration initiale

```bash
# Copier le fichier d'environnement
make init

# Ou manuellement
cp .env.example .env
```

**Important:** Éditez `.env` et changez les valeurs par défaut, surtout en production !

### 2. Démarrer les services

```bash
# Build et démarrer
make build
make up

# Ou en une seule commande
docker-compose up -d --build
```

L'API sera disponible sur **http://localhost:8080**

## 🛠️ Commandes utiles

```bash
make help          # Affiche toutes les commandes disponibles
make up            # Démarre les services
make down          # Arrête les services
make logs          # Affiche les logs
make logs-api      # Logs de l'API uniquement
make logs-db       # Logs PostgreSQL uniquement
make restart       # Redémarre les services
make clean         # Nettoie tout (⚠️ supprime les volumes)
make shell-db      # Ouvre un shell PostgreSQL
make shell-api     # Ouvre un shell dans le container API
```

## 📦 Services

### API

- **Port:** 8080
- **Container:** pokedex-api
- **Image:** Build depuis le Dockerfile local

### PostgreSQL

- **Port:** 5432
- **Container:** pokedex-db
- **Image:** postgres:16-alpine
- **Volumes:** Les données sont persistées dans un volume Docker

## 🔧 Configuration

### Variables d'environnement (.env)

```env
# Database
POSTGRES_DB=pokedex
POSTGRES_USER=pokedex_user
POSTGRES_PASSWORD=votre_mot_de_passe_securise
POSTGRES_PORT=5432

# API
API_PORT=8080
JWT_SECRET=votre_secret_jwt_super_securise
GIN_MODE=release
```

### Mode développement local (sans Docker)

Pour développer localement sans Docker, l'API utilise SQLite automatiquement :

```bash
make dev
# ou
go run main.go
```

## 🗄️ Base de données

### Accéder à PostgreSQL

```bash
# Via make
make shell-db

# Ou directement
docker-compose exec postgres psql -U pokedex_user -d pokedex
```

### Commandes PostgreSQL utiles

```sql
-- Lister les tables
\dt

-- Décrire une table
\d users

-- Voir les utilisateurs
SELECT * FROM users;

-- Compter les favoris
SELECT COUNT(*) FROM favorites;

-- Quitter
\q
```

## 🚢 Déploiement en production

### 1. Sur un serveur VPS

```bash
# Cloner le repo
git clone https://github.com/HeckoDev/pokedex-api.git
cd pokedex-api

# Configurer l'environnement
cp .env.example .env
nano .env  # Modifier avec des valeurs sécurisées

# Démarrer en production
GIN_MODE=release docker-compose up -d --build
```

### 2. Configuration SSL/TLS

Ajoutez un reverse proxy (Nginx ou Traefik) devant l'API :

**docker-compose.prod.yml** (exemple avec Traefik) :

```yaml
version: "3.8"

services:
  api:
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.pokedex-api.rule=Host(`api.votre-domaine.com`)"
      - "traefik.http.routers.pokedex-api.entrypoints=websecure"
      - "traefik.http.routers.pokedex-api.tls.certresolver=letsencrypt"
```

### 3. Backup automatique

Script de backup PostgreSQL :

```bash
#!/bin/bash
# backup.sh
docker-compose exec -T postgres pg_dump -U pokedex_user pokedex > backup_$(date +%Y%m%d_%H%M%S).sql
```

Ajoutez dans crontab :

```bash
0 2 * * * /path/to/backup.sh
```

## 📊 Monitoring

### Logs en temps réel

```bash
# Tous les services
docker-compose logs -f

# API uniquement
docker-compose logs -f api

# PostgreSQL uniquement
docker-compose logs -f postgres
```

### Statistiques des containers

```bash
docker stats
```

## 🔒 Sécurité

**⚠️ En production, n'oubliez pas de :**

1. Changer `JWT_SECRET` avec une valeur forte
2. Changer les mots de passe PostgreSQL
3. Utiliser `GIN_MODE=release`
4. Configurer un pare-feu
5. Utiliser HTTPS (reverse proxy)
6. Limiter l'accès direct à PostgreSQL (port 5432)
7. Backups réguliers

### Générer un secret sécurisé

```bash
openssl rand -base64 32
```

## 🧹 Maintenance

### Nettoyer les ressources Docker

```bash
# Arrêter et supprimer tout
make clean

# Supprimer les images inutilisées
docker image prune -a

# Voir l'espace disque utilisé
docker system df
```

### Reconstruire après modifications

```bash
docker-compose down
docker-compose up -d --build
```

## 🐛 Troubleshooting

### L'API ne démarre pas

```bash
# Vérifier les logs
make logs-api

# Vérifier que PostgreSQL est prêt
make logs-db
```

### Erreur de connexion à la base de données

```bash
# Vérifier que PostgreSQL est up
docker-compose ps

# Tester la connexion
make shell-db
```

### Reset complet

```bash
make clean
rm -rf .env
make init
# Éditer .env
make up
```

## 📝 Tests

### Tester l'API

```bash
# Inscription
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","email":"test@test.com","password":"test123"}'

# Health check (à implémenter)
curl http://localhost:8080/health
```

## 🔄 Mises à jour

```bash
# Récupérer les dernières modifications
git pull

# Reconstruire et redémarrer
docker-compose down
docker-compose up -d --build
```
