package controllers

import (
	"gocars-api/database"
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

	// Fetch user
	var user models.User
	if err := database.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}

	// Mark user as verified
	user.IsVerified = true
	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user verification"})
		return
	}

	// Generate JWT after verification
	tokenString, _ := services.GenerateJWT(input.Email)

	c.JSON(http.StatusOK, gin.H{
		"message": "Email verified successfully",
		"token":   tokenString,
		"user": gin.H{
			"id":          user.ID,
			"name":        user.Name,
			"email":       user.Email,
			"created_at":  user.CreatedAt,
			"is_verified": user.IsVerified,
		},
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
