# Pokédex API - Backend Go

API REST pour gérer les utilisateurs, favoris et équipes de Pokémon.

## 🚀 Technologies

- **Go 1.21+**
- **Gin** - Framework web
- **GORM** - ORM
- **SQLite** - Base de données
- **JWT** - Authentification
- **bcrypt** - Hachage des mots de passe

## 📦 Installation

```bash
cd pokedex-api

# Initialiser les dépendances
go mod download

# Lancer le serveur
go run main.go
```

Le serveur démarre sur `http://localhost:8080`

## 📚 Endpoints API

### Authentification

#### Inscription

```http
POST /api/register
Content-Type: application/json

{
  "username": "johndoe",
  "email": "john@example.com",
  "password": "password123"
}
```

#### Connexion

```http
POST /api/login
Content-Type: application/json

{
  "email": "john@example.com",
  "password": "password123"
}
```

Réponse:

```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": 1,
    "username": "johndoe",
    "email": "john@example.com"
  }
}
```

### Profil

#### Obtenir le profil

```http
GET /api/profile
Authorization: Bearer <token>
```

### Favoris

#### Lister les favoris

```http
GET /api/favorites
Authorization: Bearer <token>
```

#### Ajouter un favori

```http
POST /api/favorites
Authorization: Bearer <token>
Content-Type: application/json

{
  "pokemon_id": 25
}
```

#### Supprimer un favori

```http
DELETE /api/favorites/:pokemon_id
Authorization: Bearer <token>
```

### Équipes

#### Lister les équipes

```http
GET /api/teams
Authorization: Bearer <token>
```

#### Obtenir une équipe

```http
GET /api/teams/:id
Authorization: Bearer <token>
```

#### Créer une équipe

```http
POST /api/teams
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Mon Équipe Élite"
}
```

**Limite: 3 équipes maximum par utilisateur**

#### Modifier une équipe

```http
PUT /api/teams/:id
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Nouveau nom"
}
```

#### Supprimer une équipe

```http
DELETE /api/teams/:id
Authorization: Bearer <token>
```

#### Ajouter un Pokémon à une équipe

```http
POST /api/teams/:id/pokemons
Authorization: Bearer <token>
Content-Type: application/json

{
  "pokemon_id": 25,
  "position": 1,
  "nickname": "Pika",
  "is_shiny": true
}
```

**Limite: 6 Pokémon maximum par équipe (positions 1-6)**

#### Retirer un Pokémon d'une équipe

```http
DELETE /api/teams/:id/pokemons/:pokemon_id
Authorization: Bearer <token>
```

## 🗄️ Modèles de données

### User

- `id` - ID unique
- `username` - Nom d'utilisateur (unique)
- `email` - Email (unique)
- `password` - Mot de passe (hashé)
- `favorites` - Liste des favoris
- `teams` - Liste des équipes

### Favorite

- `id` - ID unique
- `user_id` - ID de l'utilisateur
- `pokemon_id` - ID du Pokédex

### Team

- `id` - ID unique
- `user_id` - ID de l'utilisateur
- `name` - Nom de l'équipe
- `pokemons` - Liste des Pokémon (max 6)

### TeamPokemon

- `id` - ID unique
- `team_id` - ID de l'équipe
- `pokemon_id` - ID du Pokédex
- `position` - Position (1-6)
- `nickname` - Surnom optionnel
- `is_shiny` - Shiny ou non

## 🔒 Sécurité

- Mots de passe hashés avec **bcrypt**
- Authentification par **JWT** (expire après 7 jours)
- Validation des données avec **Gin binding**
- Protection CORS configurée

## 🛠️ Configuration

Variables d'environnement (optionnelles):

```bash
export JWT_SECRET="votre-secret-super-securise"
```

## 📝 Notes

- La base de données SQLite (`pokedex.db`) est créée automatiquement
- Les migrations sont exécutées au démarrage
- CORS autorise localhost:5173 (Vite) et localhost:3000

## 🧪 Test avec cURL

```bash
# Inscription
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","email":"test@test.com","password":"test123"}'

# Connexion
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"test123"}'

# Créer une équipe (avec le token reçu)
curl -X POST http://localhost:8080/api/teams \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <votre-token>" \
  -d '{"name":"Équipe de Feu"}'
```
