package controllers

import (
	"gocars-api/config"
	"gocars-api/models"
	"gocars-api/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// VerifyOtp checks OTP and activates user account
func VerifyOtp(c *gin.Context) {
	var input struct {
		Email            string `json:"email" binding:"required,email"`
		VerificationCode string `json:"verification_code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate OTP via service
	if err := services.VerifyOtp(input.Email, input.VerificationCode); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Mark user as verified
	config.DB.Model(&models.User{}).Where("email = ?", input.Email).Update("is_verified", true)

	// Generate JWT after verification
	tokenString, _ := services.GenerateJWT(input.Email)

	c.JSON(http.StatusOK, gin.H{
		"message": "Email verified successfully",
		"token":   tokenString,
	})
}

// ResendOtp endpoint
func ResendOtp(c *gin.Context) {
	var input struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call OTP service
	if err := services.ResendOtp(input.Email); err != nil {
		status := http.StatusBadRequest
		if err.Error() == "please wait before requesting a new OTP" {
			status = http.StatusTooManyRequests
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "A new OTP has been sent to your email"})
}
