package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/heckodev/pokedex-api/database"
	"github.com/heckodev/pokedex-api/handlers"
	"github.com/heckodev/pokedex-api/middleware"
)

func main() {
	// Sécurité: refuser de démarrer en production avec le secret JWT par défaut
	if os.Getenv("GIN_MODE") == "release" &&
		(os.Getenv("JWT_SECRET") == "" || os.Getenv("JWT_SECRET") == "your-secret-key-change-in-production") {
		log.Fatal("Refusing to start in release mode with the default JWT_SECRET. Set a strong JWT_SECRET.")
	}

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

	// Health check (public) — utile pour le monitoring et les plateformes cloud
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Routes publiques
	public := router.Group("/api")
	{
		public.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
	}

	// Routes d'authentification (rate-limitées pour limiter le brute-force)
	auth := router.Group("/api")
	auth.Use(middleware.RateLimit(10, time.Minute))
	{
		auth.POST("/register", handlers.Register)
		auth.POST("/login", handlers.Login)
		auth.POST("/verify-email", handlers.VerifyEmail)
		auth.POST("/resend-verification", handlers.ResendVerification)
	}

	// Routes protégées (authentification requise)
	protected := router.Group("/api")
	protected.Use(middleware.AuthMiddleware())
	{
		// Profil / compte utilisateur
		protected.GET("/profile", handlers.GetProfile)
		protected.PUT("/profile", handlers.UpdateProfile)
		protected.PUT("/password", handlers.ChangePassword)
		protected.DELETE("/account", handlers.DeleteAccount)

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

	// Démarrer le serveur — le port peut être fourni par la plateforme (Render: PORT)
	port := os.Getenv("PORT")
	if port == "" {
		port = os.Getenv("API_PORT")
	}
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
