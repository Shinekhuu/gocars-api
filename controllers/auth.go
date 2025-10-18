package controllers

import (
	"gocars-api/config"
	"gocars-api/models"
	"gocars-api/services"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type SignUpInput struct {
	Email    string  `json:"email" binding:"required,email"`
	Password string  `json:"password" binding:"required,min=6"`
	Picture  *string `json:"picture"`
}

type SignInInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// SignUp registers a new user and sends OTP
func SignUp(c *gin.Context) {
	var input SignUpInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if email exists
	var existingUser models.User
	if err := config.DB.Where("email = ?", strings.ToLower(input.Email)).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email already registered"})
		return
	}

	// Hash password
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)

	// Create user with IsVerified = false
	user := models.User{
		Email:      strings.ToLower(input.Email),
		Password:   string(hashedPassword),
		Picture:    input.Picture,
		IsVerified: false,
	}
	config.DB.Create(&user)

	// Generate and send OTP via service
	if err := services.GenerateAndSendOtp(user.Email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send OTP"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User registered successfully. Please verify OTP sent to your email.",
		"user": gin.H{
			"id":      user.ID,
			"email":   user.Email,
			"picture": user.Picture,
		},
	})
}

// SignIn allows verified users to log in
func SignIn(c *gin.Context) {
	var input SignInInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var user models.User
	if err := config.DB.Where("email = ?", strings.ToLower(input.Email)).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	if !user.IsVerified {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Email not verified. Please verify your OTP first."})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	tokenString, _ := services.GenerateJWT(user.Email)

	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
		"user": gin.H{
			"id":      user.ID,
			"email":   user.Email,
			"picture": user.Picture,
		},
	})
}

func Profile(c *gin.Context) {
	email, _ := c.Get("email")

	var user models.User
	if err := config.DB.Where("email = ?", email).First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      user.ID,
		"email":   user.Email,
		"picture": user.Picture,
	})
}
