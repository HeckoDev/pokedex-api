package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/heckodev/pokedex-api/database"
	"github.com/heckodev/pokedex-api/handlers"
	"github.com/heckodev/pokedex-api/middleware"
)

func main() {
	// Connexion à la base de données
	database.Connect()

	// Configuration Gin
	router := gin.Default()

	// Configuration CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{
			"http://localhost:5173",
			"http://localhost:3000",
			"https://heckodev.github.io",
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Routes publiques
	public := router.Group("/api")
	{
		public.POST("/register", handlers.Register)
		public.POST("/login", handlers.Login)
	}

	// Routes protégées (authentification requise)
	protected := router.Group("/api")
	protected.Use(middleware.AuthMiddleware())
	{
		// Profil utilisateur
		protected.GET("/profile", handlers.GetProfile)

		// Favoris
		protected.GET("/favorites", handlers.GetFavorites)
		protected.POST("/favorites", handlers.AddFavorite)
		protected.DELETE("/favorites/:pokemon_id", handlers.RemoveFavorite)

		// Équipes
		protected.GET("/teams", handlers.GetTeams)
		protected.GET("/teams/:id", handlers.GetTeam)
		protected.POST("/teams", handlers.CreateTeam)
		protected.PUT("/teams/:id", handlers.UpdateTeam)
		protected.DELETE("/teams/:id", handlers.DeleteTeam)

		// Pokémon dans les équipes
		protected.POST("/teams/:id/pokemons", handlers.AddPokemonToTeam)
		protected.DELETE("/teams/:id/pokemons/:pokemon_id", handlers.RemovePokemonFromTeam)
	}

	// Démarrer le serveur
	log.Println("Server starting on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
