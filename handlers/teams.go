package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/heckodev/pokedex-api/database"
	"github.com/heckodev/pokedex-api/models"
)

// GetTeams retourne toutes les équipes de l'utilisateur
func GetTeams(c *gin.Context) {
	userID := c.GetUint("user_id")

	var teams []models.Team
	if err := database.DB.Where("user_id = ?", userID).Preload("Pokemons").Find(&teams).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch teams"})
		return
	}

	c.JSON(http.StatusOK, teams)
}

// GetTeam retourne une équipe spécifique
func GetTeam(c *gin.Context) {
	userID := c.GetUint("user_id")
	teamID := c.Param("id")

	var team models.Team
	if err := database.DB.Where("id = ? AND user_id = ?", teamID, userID).Preload("Pokemons").First(&team).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	c.JSON(http.StatusOK, team)
}

// CreateTeam crée une nouvelle équipe
func CreateTeam(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req models.CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Vérifier le nombre d'équipes (max 3)
	var count int64
	database.DB.Model(&models.Team{}).Where("user_id = ?", userID).Count(&count)
	if count >= 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Maximum 3 teams allowed per user"})
		return
	}

	team := models.Team{
		UserID: userID,
		Name:   req.Name,
	}

	if err := database.DB.Create(&team).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create team"})
		return
	}

	c.JSON(http.StatusCreated, team)
}

// UpdateTeam met à jour le nom d'une équipe
func UpdateTeam(c *gin.Context) {
	userID := c.GetUint("user_id")
	teamID := c.Param("id")

	var req models.UpdateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var team models.Team
	if err := database.DB.Where("id = ? AND user_id = ?", teamID, userID).First(&team).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	team.Name = req.Name
	if err := database.DB.Save(&team).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update team"})
		return
	}

	c.JSON(http.StatusOK, team)
}

// DeleteTeam supprime une équipe
func DeleteTeam(c *gin.Context) {
	userID := c.GetUint("user_id")
	teamID := c.Param("id")

	result := database.DB.Where("id = ? AND user_id = ?", teamID, userID).Delete(&models.Team{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete team"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Team deleted successfully"})
}

// AddPokemonToTeam ajoute un Pokémon à une équipe
func AddPokemonToTeam(c *gin.Context) {
	userID := c.GetUint("user_id")
	teamID := c.Param("id")

	var req models.AddPokemonToTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Vérifier que l'équipe appartient à l'utilisateur
	var team models.Team
	if err := database.DB.Where("id = ? AND user_id = ?", teamID, userID).First(&team).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	// Vérifier le nombre de Pokémon dans l'équipe (max 6)
	var count int64
	database.DB.Model(&models.TeamPokemon{}).Where("team_id = ?", teamID).Count(&count)
	if count >= 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Team is full (max 6 Pokemon)"})
		return
	}

	// Vérifier si la position est déjà occupée
	var existing models.TeamPokemon
	if err := database.DB.Where("team_id = ? AND position = ?", teamID, req.Position).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Position already occupied"})
		return
	}

	teamPokemon := models.TeamPokemon{
		TeamID:    team.ID,
		PokemonID: req.PokemonID,
		Position:  req.Position,
		Nickname:  req.Nickname,
		IsShiny:   req.IsShiny,
	}

	if err := database.DB.Create(&teamPokemon).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add Pokemon to team"})
		return
	}

	c.JSON(http.StatusCreated, teamPokemon)
}

// RemovePokemonFromTeam retire un Pokémon d'une équipe
func RemovePokemonFromTeam(c *gin.Context) {
	userID := c.GetUint("user_id")
	teamID := c.Param("id")
	pokemonID := c.Param("pokemon_id")

	// Vérifier que l'équipe appartient à l'utilisateur
	var team models.Team
	if err := database.DB.Where("id = ? AND user_id = ?", teamID, userID).First(&team).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	result := database.DB.Where("team_id = ? AND id = ?", teamID, pokemonID).Delete(&models.TeamPokemon{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove Pokemon from team"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pokemon not found in team"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pokemon removed from team successfully"})
}
