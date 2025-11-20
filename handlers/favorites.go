package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/heckodev/pokedex-api/database"
	"github.com/heckodev/pokedex-api/models"
)

// GetFavorites retourne tous les favoris de l'utilisateur
func GetFavorites(c *gin.Context) {
	userID := c.GetUint("user_id")

	var favorites []models.Favorite
	if err := database.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&favorites).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch favorites"})
		return
	}

	c.JSON(http.StatusOK, favorites)
}

// AddFavorite ajoute un Pokémon aux favoris
func AddFavorite(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req models.AddFavoriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Vérifier si déjà en favori
	var existingFav models.Favorite
	if err := database.DB.Where("user_id = ? AND pokemon_id = ?", userID, req.PokemonID).First(&existingFav).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Pokemon already in favorites"})
		return
	}

	favorite := models.Favorite{
		UserID:    userID,
		PokemonID: req.PokemonID,
	}

	if err := database.DB.Create(&favorite).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add favorite"})
		return
	}

	c.JSON(http.StatusCreated, favorite)
}

// RemoveFavorite retire un Pokémon des favoris
func RemoveFavorite(c *gin.Context) {
	userID := c.GetUint("user_id")
	pokemonID := c.Param("pokemon_id")

	result := database.DB.Where("user_id = ? AND pokemon_id = ?", userID, pokemonID).Delete(&models.Favorite{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove favorite"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Favorite not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Favorite removed successfully"})
}
