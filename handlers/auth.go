package handlers

import (
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/heckodev/pokedex-api/database"
	"github.com/heckodev/pokedex-api/mailer"
	"github.com/heckodev/pokedex-api/middleware"
	"github.com/heckodev/pokedex-api/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	verificationCodeLength = 6
	verificationValidity   = 15 * time.Minute
	resendCooldown         = 60 * time.Second
)

var (
	hasUpper = regexp.MustCompile(`[A-Z]`)
	hasLower = regexp.MustCompile(`[a-z]`)
	hasDigit = regexp.MustCompile(`[0-9]`)
)

// validatePasswordStrength enforces the password policy on the server side,
// independently of any client-side checks.
func validatePasswordStrength(pw string) string {
	if len(pw) < 8 {
		return "Password must contain at least 8 characters"
	}
	if !hasUpper.MatchString(pw) {
		return "Password must contain at least one uppercase letter"
	}
	if !hasLower.MatchString(pw) {
		return "Password must contain at least one lowercase letter"
	}
	if !hasDigit.MatchString(pw) {
		return "Password must contain at least one digit"
	}
	return ""
}

// issueVerificationCode generates a fresh code, stores it on the user and emails it.
func issueVerificationCode(user *models.User) error {
	code := mailer.GenerateCode(verificationCodeLength)
	expires := time.Now().Add(verificationValidity)
	now := time.Now()

	user.VerificationCode = code
	user.VerificationExpiresAt = &expires
	user.VerificationSentAt = &now

	if err := database.DB.Save(user).Error; err != nil {
		return err
	}
	return mailer.SendVerificationCode(user.Email, code)
}

// Register crée un nouvel utilisateur (non vérifié) et envoie un code par email.
func Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if msg := validatePasswordStrength(req.Password); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	// Vérifier si l'email existe déjà
	var existingUser models.User
	if err := database.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
		return
	}

	// Vérifier si le username existe déjà
	if err := database.DB.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	}

	// Hasher le mot de passe
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := models.User{
		Username:      req.Username,
		Email:         req.Email,
		Password:      string(hashedPassword),
		EmailVerified: false,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	if err := issueVerificationCode(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send verification email"})
		return
	}

	// Pas de token: l'utilisateur doit d'abord vérifier son email.
	c.JSON(http.StatusCreated, models.MessageResponse{
		Message: "Account created. A verification code has been sent to your email.",
		Email:   user.Email,
	})
}

// VerifyEmail valide le code reçu par email et connecte l'utilisateur.
func VerifyEmail(c *gin.Context) {
	var req models.VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	var user models.User
	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid code or email"})
		return
	}

	if user.EmailVerified {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email already verified"})
		return
	}

	if user.VerificationExpiresAt == nil || time.Now().After(*user.VerificationExpiresAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Verification code has expired"})
		return
	}

	if user.VerificationCode == "" || user.VerificationCode != req.Code {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid code or email"})
		return
	}

	// Marquer comme vérifié et effacer le code
	user.EmailVerified = true
	user.VerificationCode = ""
	user.VerificationExpiresAt = nil
	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify email"})
		return
	}

	token := generateToken(user.ID)
	c.JSON(http.StatusOK, models.AuthResponse{Token: token, User: user})
}

// ResendVerification renvoie un nouveau code de vérification (avec anti-spam).
func ResendVerification(c *gin.Context) {
	var req models.ResendVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	var user models.User
	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		// Réponse générique pour ne pas révéler l'existence du compte
		c.JSON(http.StatusOK, models.MessageResponse{
			Message: "If an account exists, a new code has been sent.",
		})
		return
	}

	if user.EmailVerified {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email already verified"})
		return
	}

	// Anti-spam: respecter un délai minimal entre deux envois
	if user.VerificationSentAt != nil && time.Since(*user.VerificationSentAt) < resendCooldown {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Please wait before requesting a new code"})
		return
	}

	if err := issueVerificationCode(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send verification email"})
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{
		Message: "A new verification code has been sent.",
	})
}

// Login authentifie un utilisateur (bloqué si email non vérifié).
func Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	var user models.User
	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if !user.EmailVerified {
		c.JSON(http.StatusForbidden, gin.H{
			"error":          "Email not verified",
			"email_verified": false,
			"email":          user.Email,
		})
		return
	}

	token := generateToken(user.ID)
	c.JSON(http.StatusOK, models.AuthResponse{Token: token, User: user})
}

// GetProfile retourne le profil de l'utilisateur connecté
func GetProfile(c *gin.Context) {
	userID := c.GetUint("user_id")

	var user models.User
	if err := database.DB.Preload("Favorites").Preload("Teams.Pokemons").First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateProfile met à jour le nom d'utilisateur.
func UpdateProfile(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Vérifier l'unicité du username (hors utilisateur courant)
	var existing models.User
	if err := database.DB.Where("username = ? AND id <> ?", req.Username, userID).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	}

	user.Username = req.Username
	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// ChangePassword modifie le mot de passe (nécessite le mot de passe actuel).
func ChangePassword(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
		return
	}

	if msg := validatePasswordStrength(req.NewPassword); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user.Password = string(hashed)
	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{Message: "Password updated successfully"})
}

// DeleteAccount supprime définitivement le compte et toutes ses données.
func DeleteAccount(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req models.DeleteAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Password is incorrect"})
		return
	}

	// Supprimer définitivement l'utilisateur et ses données associées
	if err := deleteUserData(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete account"})
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{Message: "Account deleted successfully"})
}

// deleteUserData supprime définitivement (hard delete) l'utilisateur et toutes
// ses données liées (favoris, équipes, pokémons d'équipe) dans une transaction.
func deleteUserData(userID uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var teamIDs []uint
		if err := tx.Model(&models.Team{}).
			Where("user_id = ?", userID).
			Pluck("id", &teamIDs).Error; err != nil {
			return err
		}

		if len(teamIDs) > 0 {
			if err := tx.Where("team_id IN ?", teamIDs).
				Delete(&models.TeamPokemon{}).Error; err != nil {
				return err
			}
		}

		if err := tx.Unscoped().Where("user_id = ?", userID).
			Delete(&models.Team{}).Error; err != nil {
			return err
		}

		if err := tx.Where("user_id = ?", userID).
			Delete(&models.Favorite{}).Error; err != nil {
			return err
		}

		return tx.Unscoped().Delete(&models.User{}, userID).Error
	})
}

func generateToken(userID uint) string {
	claims := middleware.Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour * 7)), // 7 jours
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(middleware.GetJWTSecret())
	return tokenString
}
