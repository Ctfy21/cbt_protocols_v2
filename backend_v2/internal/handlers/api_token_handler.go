package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"backend_v2/internal/models"
	"backend_v2/internal/services"
)

// APITokenHandler handles API token-related HTTP requests
type APITokenHandler struct {
	apiTokenService *services.APITokenService
}

// NewAPITokenHandler creates a new API token handler
func NewAPITokenHandler(apiTokenService *services.APITokenService) *APITokenHandler {
	return &APITokenHandler{
		apiTokenService: apiTokenService,
	}
}

// CreateAPIToken handles POST /api-tokens
func (h *APITokenHandler) CreateAPIToken(c *gin.Context) {
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

	var req models.CreateAPITokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(err.Error()))
		return
	}

	// Create API token
	tokenResp, err := h.apiTokenService.CreateAPIToken(user.ID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, models.SuccessResponse(tokenResp))
}

// GetAPITokens handles GET /api-tokens
func (h *APITokenHandler) GetAPITokens(c *gin.Context) {
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

	// Get API tokens
	tokens, err := h.apiTokenService.GetAPITokens(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(tokens))
}

// RevokeAPIToken handles DELETE /api-tokens/:id
func (h *APITokenHandler) RevokeAPIToken(c *gin.Context) {
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

	// Parse token ID
	tokenIDStr := c.Param("id")
	tokenID, err := primitive.ObjectIDFromHex(tokenIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Invalid token ID"))
		return
	}

	// Revoke token
	if err := h.apiTokenService.RevokeAPIToken(tokenID, user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse("API token revoked successfully"))
}
