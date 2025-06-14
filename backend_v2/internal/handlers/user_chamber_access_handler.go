package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"backend_v2/internal/models"
	"backend_v2/internal/services"
)

// UserChamberAccessHandler handles user chamber access-related HTTP requests
type UserChamberAccessHandler struct {
	userChamberAccessService *services.UserChamberAccessService
}

// NewUserChamberAccessHandler creates a new user chamber access handler
func NewUserChamberAccessHandler(userChamberAccessService *services.UserChamberAccessService) *UserChamberAccessHandler {
	return &UserChamberAccessHandler{
		userChamberAccessService: userChamberAccessService,
	}
}

// SetUserChamberAccess handles PUT /users/:id/chambers
// Sets chamber access for a user (replaces all existing access)
func (h *UserChamberAccessHandler) SetUserChamberAccess(c *gin.Context) {
	userID := c.Param("id")

	var req struct {
		ChamberIDs []string `json:"chamber_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(err.Error()))
		return
	}

	err := h.userChamberAccessService.SetUserChamberAccess(userID, req.ChamberIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse("Chamber access updated successfully"))
}

// GetUserChamberAccess handles GET /users/:id/chambers
// Gets all chambers a user has access to
func (h *UserChamberAccessHandler) GetUserChamberAccess(c *gin.Context) {
	userIDStr := c.Param("id")

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Invalid user ID"))
		return
	}

	chambers, err := h.userChamberAccessService.GetUserChamberAccess(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(chambers))
}

// GrantChamberAccess handles POST /users/:id/chambers/:chamber_id
// Grants chamber access to a user
func (h *UserChamberAccessHandler) GrantChamberAccess(c *gin.Context) {
	userIDStr := c.Param("id")
	chamberIDStr := c.Param("chamber_id")

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Invalid user ID"))
		return
	}

	chamberID, err := primitive.ObjectIDFromHex(chamberIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Invalid chamber ID"))
		return
	}

	err = h.userChamberAccessService.GrantChamberAccess(userID, chamberID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse("Chamber access granted successfully"))
}

// RevokeChamberAccess handles DELETE /users/:id/chambers/:chamber_id
// Revokes chamber access from a user
func (h *UserChamberAccessHandler) RevokeChamberAccess(c *gin.Context) {
	userIDStr := c.Param("id")
	chamberIDStr := c.Param("chamber_id")

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Invalid user ID"))
		return
	}

	chamberID, err := primitive.ObjectIDFromHex(chamberIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Invalid chamber ID"))
		return
	}

	err = h.userChamberAccessService.RevokeChamberAccess(userID, chamberID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse("Chamber access revoked successfully"))
}

// GetAllUsersWithChamberAccess handles GET /users/chambers
// Gets all users with their chamber access
func (h *UserChamberAccessHandler) GetAllUsersWithChamberAccess(c *gin.Context) {
	usersWithAccess, err := h.userChamberAccessService.GetAllUsersWithChamberAccess()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(usersWithAccess))
}

// HasChamberAccess handles GET /users/:id/chambers/:chamber_id/check
// Checks if user has access to specific chamber
func (h *UserChamberAccessHandler) HasChamberAccess(c *gin.Context) {
	userIDStr := c.Param("id")
	chamberIDStr := c.Param("chamber_id")

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Invalid user ID"))
		return
	}

	chamberID, err := primitive.ObjectIDFromHex(chamberIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Invalid chamber ID"))
		return
	}

	hasAccess, err := h.userChamberAccessService.HasChamberAccess(userID, chamberID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
		"has_access": hasAccess,
	}))
}
