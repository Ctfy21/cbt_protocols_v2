package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"

	"backend_v2/internal/models"
	"backend_v2/internal/services"
)

// UserHandler handles user management HTTP requests
type UserHandler struct {
	authService *services.AuthService
}

// NewUserHandler creates a new user handler
func NewUserHandler(authService *services.AuthService) *UserHandler {
	return &UserHandler{
		authService: authService,
	}
}

// CreateUser handles POST /users (Admin only)
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(err.Error()))
		return
	}

	// Check role from request body (optional, defaults to user)
	type CreateUserRequest struct {
		models.RegisterRequest
		Role models.UserRole `json:"role"`
	}

	var createReq CreateUserRequest
	if err := c.ShouldBindJSON(&createReq); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(err.Error()))
		return
	}

	// If no role specified, default to user
	if createReq.Role == "" {
		createReq.Role = models.RoleUser
	}

	// Create user through auth service (we'll need to add this method)
	user, err := h.authService.CreateUser(&createReq.RegisterRequest, createReq.Role)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, models.SuccessResponse(user))
}

// GetUsers handles GET /users (Admin only)
func (h *UserHandler) GetUsers(c *gin.Context) {
	users, err := h.authService.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(users))
}

// GetUser handles GET /users/:id (Admin only)
func (h *UserHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")

	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(user))
}

// UpdateUser handles PUT /users/:id (Admin only)
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")

	var req struct {
		Name     string          `json:"name"`
		Email    string          `json:"email"`
		Role     models.UserRole `json:"role"`
		IsActive *bool           `json:"is_active"`
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
	if req.Email != "" {
		update["email"] = req.Email
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
func (h *UserHandler) DeactivateUser(c *gin.Context) {
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
func (h *UserHandler) ActivateUser(c *gin.Context) {
	userID := c.Param("id")

	update := bson.M{"is_active": true}
	_, err := h.authService.UpdateUser(userID, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse("User activated successfully"))
}
