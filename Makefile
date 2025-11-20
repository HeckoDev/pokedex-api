.PHONY: help dev build up down logs clean restart test

help: ## Affiche cette aide
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

dev: ## Lance le serveur en mode développement (local, SQLite)
	@echo "🚀 Starting development server..."
	go run main.go

build: ## Build les images Docker
	@echo "🏗️  Building Docker images..."
	docker-compose build

up: ## Lance les services avec Docker Compose
	@echo "🚀 Starting services..."
	docker-compose up -d
	@echo "✅ Services started!"
	@echo "📊 API: http://localhost:8080"
	@echo "🗄️  PostgreSQL: localhost:5432"

down: ## Arrête les services
	@echo "🛑 Stopping services..."
	docker-compose down

logs: ## Affiche les logs
	docker-compose logs -f

logs-api: ## Affiche les logs de l'API uniquement
	docker-compose logs -f api

logs-db: ## Affiche les logs de la base de données
	docker-compose logs -f postgres

restart: ## Redémarre les services
	@echo "🔄 Restarting services..."
	docker-compose restart

clean: ## Nettoie tout (containers, volumes, images)
	@echo "🧹 Cleaning up..."
	docker-compose down -v
	docker system prune -f

test: ## Lance les tests
	@echo "🧪 Running tests..."
	go test -v ./...

migrate: ## Lance les migrations manuellement
	@echo "🗄️  Running migrations..."
	docker-compose exec api ./main migrate

shell-db: ## Ouvre un shell PostgreSQL
	docker-compose exec postgres psql -U pokedex_user -d pokedex

shell-api: ## Ouvre un shell dans le container API
	docker-compose exec api sh

deps: ## Installe les dépendances Go
	@echo "📦 Installing dependencies..."
	go mod download
	go mod tidy

init: ## Initialise le projet (copie .env.example vers .env)
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "✅ .env file created from .env.example"; \
		echo "⚠️  Don't forget to update the values in .env"; \
	else \
		echo "ℹ️  .env file already exists"; \
	fi
