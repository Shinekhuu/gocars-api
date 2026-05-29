package handler

import (
	"net/http"

	"gocars-api/internal/profile/repository/postgresql/model"
	profilesvc "gocars-api/internal/profile/service"

	"github.com/gin-gonic/gin"
)

type ProfileHandler struct {
	profileSvc *profilesvc.ProfileService
}

func NewProfileHandler(profileSvc *profilesvc.ProfileService) *ProfileHandler {
	return &ProfileHandler{profileSvc: profileSvc}
}

func (h *ProfileHandler) Profile(c *gin.Context) {
	userID, _ := c.Get("user_id")
	email, _ := c.Get("email")

	uid, ok := userID.(string)
	if !ok || uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	profile, err := h.profileSvc.GetProfile(uid)
	if err != nil {
		// Profile row not created yet — return basic info
		c.JSON(http.StatusOK, gin.H{"user_id": uid, "email": email})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": uid,
		"email":   email,
		"profile": profile,
	})
}

func (h *ProfileHandler) UpdateProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, ok := userID.(string)
	if !ok || uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var input struct {
		DisplayName string `json:"display_name"`
		AvatarURL   string `json:"avatar_url"`
		Phone       string `json:"phone"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	p := &model.Profile{
		ID:          uid,
		DisplayName: input.DisplayName,
		AvatarURL:   input.AvatarURL,
		Phone:       input.Phone,
	}
	if err := h.profileSvc.UpsertProfile(p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "profile updated", "profile": p})
}
