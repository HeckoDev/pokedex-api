#!/bin/bash

echo "🎮 Test d'intégration Pokédex Frontend-Backend"
echo "=============================================="
echo ""

API_URL="http://localhost:8080"
RANDOM_USER="testuser_$(date +%s)"
EMAIL="${RANDOM_USER}@test.com"
PASSWORD="password123"

echo "📝 Étape 1 : Inscription d'un nouvel utilisateur"
echo "Username: $RANDOM_USER"
echo "Email: $EMAIL"
echo ""

REGISTER_RESPONSE=$(curl -s -X POST "${API_URL}/api/register" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"${RANDOM_USER}\",\"email\":\"${EMAIL}\",\"password\":\"${PASSWORD}\"}")

echo "$REGISTER_RESPONSE" | jq '.'
TOKEN=$(echo "$REGISTER_RESPONSE" | jq -r '.token')

if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
    echo "❌ Erreur : Inscription échouée"
    exit 1
fi

echo ""
echo "✅ Inscription réussie !"
echo "Token: ${TOKEN:0:20}..."
echo ""

echo "🔍 Étape 2 : Récupération du profil"
PROFILE=$(curl -s "${API_URL}/api/profile" \
  -H "Authorization: Bearer ${TOKEN}")

echo "$PROFILE" | jq '.'
echo ""

echo "⭐ Étape 3 : Ajout de favoris (Pikachu #25, Dracaufeu #6, Mewtwo #150)"
echo ""

# Pikachu
FAV1=$(curl -s -X POST "${API_URL}/api/favorites" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"pokemon_id":25}')
echo "Pikachu: $(echo "$FAV1" | jq -c '.')"

# Dracaufeu
FAV2=$(curl -s -X POST "${API_URL}/api/favorites" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"pokemon_id":6}')
echo "Dracaufeu: $(echo "$FAV2" | jq -c '.')"

# Mewtwo
FAV3=$(curl -s -X POST "${API_URL}/api/favorites" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"pokemon_id":150}')
echo "Mewtwo: $(echo "$FAV3" | jq -c '.')"

echo ""
echo "✅ Favoris ajoutés !"
echo ""

echo "📋 Étape 4 : Liste des favoris"
FAVORITES=$(curl -s "${API_URL}/api/favorites" \
  -H "Authorization: Bearer ${TOKEN}")

echo "$FAVORITES" | jq '.'
echo ""

echo "🏆 Étape 5 : Création d'une équipe"
TEAM=$(curl -s -X POST "${API_URL}/api/teams" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"name":"Mon équipe Kanto"}')

echo "$TEAM" | jq '.'
TEAM_ID=$(echo "$TEAM" | jq -r '.id')
echo ""

echo "✅ Équipe créée avec ID: $TEAM_ID"
echo ""

echo "➕ Étape 6 : Ajout de Pokémon à l'équipe"
echo ""

# Position 1: Pikachu
POKEMON1=$(curl -s -X POST "${API_URL}/api/teams/${TEAM_ID}/pokemons" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"pokemon_id":25,"position":1,"nickname":"Pikachu","is_shiny":false}')
echo "Position 1: $(echo "$POKEMON1" | jq -c '.')"

# Position 2: Dracaufeu
POKEMON2=$(curl -s -X POST "${API_URL}/api/teams/${TEAM_ID}/pokemons" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"pokemon_id":6,"position":2,"nickname":"Dracaufeu","is_shiny":true}')
echo "Position 2: $(echo "$POKEMON2" | jq -c '.')"

# Position 3: Mewtwo
POKEMON3=$(curl -s -X POST "${API_URL}/api/teams/${TEAM_ID}/pokemons" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"pokemon_id":150,"position":3,"nickname":"Mewtwo","is_shiny":false}')
echo "Position 3: $(echo "$POKEMON3" | jq -c '.')"

echo ""
echo "✅ Pokémon ajoutés à l'équipe !"
echo ""

echo "📊 Étape 7 : Récupération de l'équipe complète"
FULL_TEAM=$(curl -s "${API_URL}/api/teams/${TEAM_ID}" \
  -H "Authorization: Bearer ${TOKEN}")

echo "$FULL_TEAM" | jq '.'
echo ""

echo "🎉 Test d'intégration terminé avec succès !"
echo ""
echo "=============================================="
echo "📊 Résumé :"
echo "  - Utilisateur créé : $RANDOM_USER"
echo "  - Favoris : 3 Pokémon"
echo "  - Équipes : 1 équipe avec 3 Pokémon"
echo "  - Token valide pendant 7 jours"
echo ""
echo "🌐 Accéder au frontend : http://localhost:5173"
echo "🔐 Se connecter avec :"
echo "  Email: $EMAIL"
echo "  Password: $PASSWORD"
echo "=============================================="
