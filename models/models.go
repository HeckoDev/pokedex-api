package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Username  string         `gorm:"unique;not null" json:"username"`
	Email     string         `gorm:"unique;not null" json:"email"`
	Password  string         `gorm:"not null" json:"-"`
	// Email verification
	EmailVerified         bool       `gorm:"default:false" json:"email_verified"`
	VerificationCode      string     `gorm:"default:''" json:"-"`
	VerificationExpiresAt *time.Time `json:"-"`
	VerificationSentAt    *time.Time `json:"-"`
	Favorites             []Favorite `json:"favorites,omitempty"`
	Teams                 []Team     `json:"teams,omitempty"`
}

type Favorite struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	PokemonID int       `gorm:"not null" json:"pokemon_id"` // ID du Pokédex
	User      User      `gorm:"foreignKey:UserID" json:"-"`
}

type Team struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	Name      string         `gorm:"not null" json:"name"`
	Pokemons  []TeamPokemon  `json:"pokemons,omitempty"`
	User      User           `gorm:"foreignKey:UserID" json:"-"`
}

type TeamPokemon struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	TeamID    uint      `gorm:"not null;index" json:"team_id"`
	PokemonID int       `gorm:"not null" json:"pokemon_id"` // ID du Pokédex
	Position  int       `gorm:"not null" json:"position"`   // Position dans l'équipe (1-6)
	Nickname  string    `json:"nickname,omitempty"`         // Surnom optionnel
	IsShiny   bool      `gorm:"default:false" json:"is_shiny"`
	Team      Team      `gorm:"foreignKey:TeamID" json:"-"`
}

// DTOs pour les requêtes/réponses
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type VerifyEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required,len=6"`
}

type ResendVerificationRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type UpdateProfileRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

type DeleteAccountRequest struct {
	Password string `json:"password" binding:"required"`
}

type MessageResponse struct {
	Message string `json:"message"`
	Email   string `json:"email,omitempty"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type AddFavoriteRequest struct {
	PokemonID int `json:"pokemon_id" binding:"required,min=1"`
}

type CreateTeamRequest struct {
	Name string `json:"name" binding:"required,min=1,max=50"`
}

type UpdateTeamRequest struct {
	Name string `json:"name" binding:"required,min=1,max=50"`
}

type AddPokemonToTeamRequest struct {
	PokemonID int    `json:"pokemon_id" binding:"required,min=1"`
	Position  int    `json:"position" binding:"required,min=1,max=6"`
	Nickname  string `json:"nickname,omitempty"`
	IsShiny   bool   `json:"is_shiny"`
}
