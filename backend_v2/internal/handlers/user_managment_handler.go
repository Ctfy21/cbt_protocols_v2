package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"

	"backend_v2/internal/models"
	"backend_v2/internal/services"
)

// UserManagementHandler handles user management HTTP requests (Admin only)
type UserManagementHandler struct {
	authService *services.AuthService
}

// NewUserManagementHandler creates a new user management handler
func NewUserManagementHandler(authService *services.AuthService) *UserManagementHandler {
	return &UserManagementHandler{
		authService: authService,
	}
}

// CreateUser handles POST /users (Admin only)
func (h *UserManagementHandler) CreateUser(c *gin.Context) {
	var req struct {
		models.RegisterRequest
		Role models.UserRole `json:"role"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(err.Error()))
		return
	}

	// If no role specified, default to user
	if req.Role == "" {
		req.Role = models.RoleUser
	}

	// Validate role
	if req.Role != models.RoleUser && req.Role != models.RoleAdmin {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Invalid role"))
		return
	}

	// Create user through auth service
	user, err := h.authService.CreateUser(&req.RegisterRequest, req.Role)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, models.SuccessResponse(user))
}

// GetUsers handles GET /users (Admin only)
func (h *UserManagementHandler) GetUsers(c *gin.Context) {
	users, err := h.authService.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(users))
}

// GetUser handles GET /users/:id (Admin only)
func (h *UserManagementHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")

	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(user))
}

// UpdateUser handles PUT /users/:id (Admin only)
func (h *UserManagementHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")

	var req struct {
		Name     string          `json:"name"`
		Username string          `json:"username"`
		Role     models.UserRole `json:"role"`
		IsActive *bool           `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(err.Error()))
		return
	}

	// Validate role if provided
	if req.Role != "" && req.Role != models.RoleUser && req.Role != models.RoleAdmin {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Invalid role"))
		return
	}

	// Get current user ID from context to prevent self-modification in critical ways
	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse("User not found in context"))
		return
	}

	// Prevent admin from deactivating themselves
	if req.IsActive != nil && !*req.IsActive && currentUserID.(string) == userID {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Cannot deactivate your own account"))
		return
	}

	// Prepare update
	update := bson.M{}
	if req.Name != "" {
		update["name"] = req.Name
	}
	if req.Username != "" {
		update["username"] = req.Username
	}
	if req.Role != "" {
		update["role"] = req.Role
	}
	if req.IsActive != nil {
		update["is_active"] = *req.IsActive
	}

	if len(update) == 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("No fields to update"))
		return
	}

	updatedUser, err := h.authService.UpdateUser(userID, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(updatedUser))
}

// DeactivateUser handles DELETE /users/:id (Admin only - soft delete)
func (h *UserManagementHandler) DeactivateUser(c *gin.Context) {
	userID := c.Param("id")

	// Don't allow admin to deactivate themselves
	currentUserID, exists := c.Get("user_id")
	if exists && currentUserID.(string) == userID {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Cannot deactivate your own account"))
		return
	}

	update := bson.M{"is_active": false}
	_, err := h.authService.UpdateUser(userID, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse("User deactivated successfully"))
}

// ActivateUser handles POST /users/:id/activate (Admin only)
func (h *UserManagementHandler) ActivateUser(c *gin.Context) {
	userID := c.Param("id")

	update := bson.M{"is_active": true}
	_, err := h.authService.UpdateUser(userID, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse("User activated successfully"))
}

// GetUserStats handles GET /users/stats (Admin only)
func (h *UserManagementHandler) GetUserStats(c *gin.Context) {
	users, err := h.authService.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
		return
	}

	stats := struct {
		Total  int `json:"total"`
		Active int `json:"active"`
		Admins int `json:"admins"`
	}{
		Total: len(users),
	}

	for _, user := range users {
		if user.IsActive {
			stats.Active++
		}
		if user.Role == models.RoleAdmin {
			stats.Admins++
		}
	}

	c.JSON(http.StatusOK, models.SuccessResponse(stats))
}

// ResetUserPassword handles POST /users/:id/reset-password (Admin only)
func (h *UserManagementHandler) ResetUserPassword(c *gin.Context) {
	userID := c.Param("id")

	var req struct {
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(err.Error()))
		return
	}

	// Update password directly (admin function)
	err := h.authService.UpdatePassword(userID, req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse("Password reset successfully"))
}

// SearchUsers handles GET /users/search (Admin only)
func (h *UserManagementHandler) SearchUsers(c *gin.Context) {
	query := c.Query("q")
	role := c.Query("role")
	status := c.Query("status")

	users, err := h.authService.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
		return
	}

	// Filter users based on query parameters
	var filteredUsers []models.User
	for _, user := range users {
		// Skip filtering if no criteria provided
		include := true

		// Text search filter
		if query != "" {
			include = include && (contains(user.Name, query) ||
				contains(user.Username, query))
		}

		// Role filter
		if role != "" && role != "all" {
			include = include && (string(user.Role) == role)
		}

		// Status filter
		if status != "" && status != "all" {
			if status == "active" {
				include = include && user.IsActive
			} else if status == "inactive" {
				include = include && !user.IsActive
			}
		}

		if include {
			filteredUsers = append(filteredUsers, user)
		}
	}

	c.JSON(http.StatusOK, models.SuccessResponse(filteredUsers))
}

// Helper function for case-insensitive string search
func contains(text, substr string) bool {
	// Simple case-insensitive search
	// In production, you might want to use a proper text search library
	return len(text) >= len(substr) &&
		findInString(strings.ToLower(text), strings.ToLower(substr)) >= 0
}

func findInString(text, substr string) int {
	for i := 0; i <= len(text)-len(substr); i++ {
		if text[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
