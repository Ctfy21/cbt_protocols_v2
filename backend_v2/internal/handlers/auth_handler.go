package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"

	"backend_v2/internal/models"
	"backend_v2/internal/services"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authService *services.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// AuthService returns the auth service instance
func (h *AuthHandler) AuthService() *services.AuthService {
	return h.authService
}

// Register handles POST /auth/register
// func (h *AuthHandler) Register(c *gin.Context) {
// 	var req models.RegisterRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, models.ErrorResponse(err.Error()))
// 		return
// 	}

// 	authResp, err := h.authService.Register(&req)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, models.ErrorResponse(err.Error()))
// 		return
// 	}

// 	c.JSON(http.StatusCreated, models.SuccessResponse(authResp))
// }

// Login handles POST /auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(err.Error()))
		return
	}

	// Get user agent and IP for session tracking
	userAgent := c.GetHeader("User-Agent")
	ip := c.ClientIP()

	authResp, err := h.authService.Login(&req, userAgent, ip)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(authResp))
}

// RefreshToken handles POST /auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req models.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(err.Error()))
		return
	}

	authResp, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(authResp))
}

// Logout handles POST /auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// Get token from header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Missing authorization header"))
		return
	}

	// Extract token from "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Invalid authorization header format"))
		return
	}

	token := parts[1]
	err := h.authService.Logout(token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse("Successfully logged out"))
}

// Me handles GET /auth/me
func (h *AuthHandler) Me(c *gin.Context) {
	// Get user from context (set by auth middleware)
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse("User not found"))
		return
	}

	user, ok := userInterface.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Invalid user data"))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(user))
}

// UpdateProfile handles PUT /auth/profile
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	// Get user from context
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse("User not found"))
		return
	}

	user, ok := userInterface.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Invalid user data"))
		return
	}

	// Parse update request
	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(err.Error()))
		return
	}

	// Prepare update
	update := bson.M{}
	if req.Name != "" {
		update["name"] = req.Name
	}

	if len(update) == 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("No fields to update"))
		return
	}

	// Update user
	updatedUser, err := h.authService.UpdateUser(user.ID.Hex(), update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(updatedUser))
}

// ChangePassword handles POST /auth/change-password
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	// Get user from context
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse("User not found"))
		return
	}

	user, ok := userInterface.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Invalid user data"))
		return
	}

	// Parse request
	var req struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(err.Error()))
		return
	}

	// Verify current password
	err := h.authService.VerifyPassword(user.ID.Hex(), req.CurrentPassword)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse("Current password is incorrect"))
		return
	}

	// Update password
	err = h.authService.UpdatePassword(user.ID.Hex(), req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse("Password changed successfully"))
}
