package database

import (
	"fmt"
	"log"
	"os"

	"github.com/heckodev/pokedex-api/models"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect() {
	var err error
	var dialector gorm.Dialector

	// Vérifier si on utilise PostgreSQL (via variables d'environnement)
	dbHost := os.Getenv("DB_HOST")
	
	if dbHost != "" {
		// Configuration PostgreSQL
		dbUser := getEnv("DB_USER", "pokedex_user")
		dbPassword := getEnv("DB_PASSWORD", "pokedex_password")
		dbName := getEnv("DB_NAME", "pokedex")
		dbPort := getEnv("DB_PORT", "5432")

		dsn := fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
			dbHost, dbUser, dbPassword, dbName, dbPort,
		)
		
		dialector = postgres.Open(dsn)
		log.Println("Using PostgreSQL database")
	} else {
		// Fallback sur SQLite pour le développement local
		dialector = sqlite.Open("pokedex.db")
		log.Println("Using SQLite database (development mode)")
	}

	DB, err = gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto-migrate les modèles
	err = DB.AutoMigrate(
		&models.User{},
		&models.Favorite{},
		&models.Team{},
		&models.TeamPokemon{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	log.Println("Database connected and migrated successfully")
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
