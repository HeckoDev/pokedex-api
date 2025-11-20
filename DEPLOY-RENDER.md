# 🚀 Guide de déploiement Render - Étape par étape

## 📋 Prérequis

- [x] Compte GitHub
- [x] Repository avec le code
- [ ] Compte Render (gratuit) : https://render.com

---

## 1️⃣ Créer la base de données PostgreSQL

### Sur Render Dashboard

1. **Aller sur** : https://dashboard.render.com
2. **Cliquer sur** : `New +` → `PostgreSQL`
3. **Remplir** :
   - Name: `pokedex-db`
   - Database: `pokedex`
   - User: `pokedex_user`
   - Region: `Oregon (US West)` ou le plus proche
4. **Cliquer** : `Create Database`

### ⏳ Attendre 2-3 minutes

La base de données se crée automatiquement.

### 📝 Noter les informations de connexion

Une fois créée, aller dans l'onglet **Info** et noter :

```
Internal Database URL: postgresql://pokedex_user:xxx@dpg-xxx.oregon-postgres.render.com/pokedex
Hostname: dpg-xxx-a.oregon-postgres.render.com
Port: 5432
Database: pokedex
Username: pokedex_user
Password: <mot de passe auto-généré>
```

---

## 2️⃣ Créer le Web Service

### Sur Render Dashboard

1. **Cliquer sur** : `New +` → `Web Service`
2. **Connect repository** :
   - Si première fois : `Connect GitHub`
   - Autoriser Render à accéder à vos repos
   - Sélectionner : `pokedex-vue` (ou votre nom de repo)

### Configuration du service

**Build & Deploy**

```
Name: pokedex-api
Root Directory: pokedex-api
Environment: Go ⚠️ IMPORTANT: Sélectionner "Go", pas "Node"
Branch: main
```

**Build Command**

```bash
go build -tags netgo -ldflags '-s -w' -o app
```

ℹ️ Alternative simple : `go build -o main .`

**Start Command**

```bash
./app
```

ℹ️ Alternative simple : `./main`

**Instance Type**

- Free (dort après 15min d'inactivité) ✅
- Starter - $7/month (toujours actif)

### ⚠️ Ne pas encore déployer !

Cliquer sur **Advanced** pour configurer les variables d'environnement d'abord.

---

## 3️⃣ Configurer les variables d'environnement

Dans la section **Environment Variables**, ajouter :

### Variables obligatoires

| Key                 | Value               | Exemple                                |
| ------------------- | ------------------- | -------------------------------------- |
| `DB_HOST`           | Hostname PostgreSQL | `dpg-xxx-a.oregon-postgres.render.com` |
| `POSTGRES_DB`       | `pokedex`           | `pokedex`                              |
| `POSTGRES_USER`     | `pokedex_user`      | `pokedex_user`                         |
| `POSTGRES_PASSWORD` | Password PostgreSQL | Copier depuis Info PostgreSQL          |
| `POSTGRES_PORT`     | `5432`              | `5432`                                 |
| `GIN_MODE`          | `release`           | `release`                              |
| `API_PORT`          | `8080`              | `8080`                                 |

### Variable JWT_SECRET

**Générer un secret fort** :

```bash
openssl rand -base64 32
```

Résultat exemple : `kX9mP2nQ5rS8tU1vW3xY6zA0bC4dE7fH9jK2lM5nO8=`

Ajouter :

```
JWT_SECRET=<votre secret généré>
```

### ✅ Vérifier

Vous devez avoir **8 variables** au total.

---

## 4️⃣ Déployer

1. **Cliquer** : `Create Web Service`
2. **Attendre** : 3-5 minutes pour le premier déploiement

### 👀 Suivre les logs

Les logs s'affichent en temps réel :

```
Building...
Step 1/10 : FROM golang:1.21-alpine AS builder
...
Starting service with './main'
Database connected and migrated successfully
Server starting on :8080
```

### ✅ Déploiement réussi

Quand vous voyez :

```
Your service is live 🎉
```

---

## 5️⃣ Tester l'API

### Récupérer l'URL

En haut de la page : `https://pokedex-api-xxxx.onrender.com`

### Test d'inscription

```bash
curl -X POST https://votre-api.onrender.com/api/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","email":"test@test.com","password":"password123"}'
```

Réponse attendue :

```json
{
  "token": "eyJhbGc...",
  "user": {
    "id": 1,
    "username": "test",
    "email": "test@test.com"
  }
}
```

### ✅ API fonctionnelle !

---

## 6️⃣ Mettre à jour le frontend

### Modifier l'URL de l'API

```bash
cd /home/hecko/projects/Sandbox/Pokedex/pokedex-vue
```

**Éditer `.env.production`** :

```env
VITE_API_URL=https://votre-api.onrender.com
```

**Éditer `.env`** :

```env
VITE_API_URL=https://votre-api.onrender.com
```

### Commit et push

```bash
git add .
git commit -m "Update API URL for production"
git push origin main
```

Le frontend se redéploiera automatiquement via GitHub Actions.

---

## 7️⃣ Configurer le domaine personnalisé (Optionnel)

### Dans Render

1. **Service Settings** → `Custom Domains`
2. **Add Custom Domain** : `api.votredomaine.com`
3. **Configurer DNS** chez votre registrar :
   ```
   Type: CNAME
   Name: api
   Value: pokedex-api-xxxx.onrender.com
   ```

### HTTPS

Render configure automatiquement le certificat SSL (Let's Encrypt).

---

## 🔧 Maintenance

### Voir les logs

Dashboard → Service → **Logs** (temps réel)

### Redéployer

Dashboard → Service → **Manual Deploy** → Dernière version

### Mettre à jour les variables

Dashboard → Service → **Environment** → Modifier et sauvegarder

### Monitorer

Dashboard → Service → **Metrics**

- CPU Usage
- Memory Usage
- Request Count

---

## 🐛 Troubleshooting

### Service ne démarre pas

**Vérifier les logs** pour :

```
Error: failed to connect to database
```

**Solution** : Vérifier `DB_HOST` et credentials PostgreSQL

### Build échoue

```
go: module requires Go 1.21
```

**Solution** : Render utilise Go 1.21, c'est normal. Vérifier `go.mod`

### 502 Bad Gateway

**Cause** : Service crash au démarrage

**Solution** :

1. Vérifier les variables d'environnement
2. Vérifier les logs
3. Tester en local avec Docker

### Database connection timeout

**Solution** : Utiliser le **Internal Database URL** (pas External)

---

## 💰 Coûts

### Free Tier

- **Web Service** : 750h/mois
  - ⚠️ Dort après 15min d'inactivité
  - Premier démarrage : ~30 secondes
- **PostgreSQL** : 90 jours gratuits
  - Puis $7/mois

### Starter Plan (recommandé)

- **Web Service** : $7/mois
  - ✅ Toujours actif
  - Démarrage instantané
- **PostgreSQL** : $7/mois
  - Backups automatiques
  - Point-in-time recovery

**Total** : $14/mois

---

## 📊 Résumé

✅ PostgreSQL créée  
✅ Web Service configuré  
✅ Variables d'environnement ajoutées  
✅ API déployée et testée  
✅ Frontend mis à jour  
✅ HTTPS automatique

---

## 🎉 C'est fait !

Votre API est maintenant accessible publiquement :

**API** : `https://votre-api.onrender.com`  
**Frontend** : `https://heckodev.github.io/pokedex-vue/`

### Prochaines étapes

1. Tester l'inscription sur le frontend
2. Créer des favoris
3. Créer une équipe
4. Partager votre Pokédex !

---

## 📞 Support

**Documentation** :

- [DEPLOYMENT.md](../DEPLOYMENT.md) - Guide complet
- [Render Docs](https://render.com/docs) - Documentation officielle

**Problème** :

- GitHub Issues : https://github.com/HeckoDev/pokedex-vue/issues
- Render Support : support@render.com
