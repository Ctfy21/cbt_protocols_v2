package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"backend_v2/internal/models"
	"backend_v2/internal/services"
)

// UserRoomChamberAccessHandler handles user room chamber access-related HTTP requests
type UserRoomChamberAccessHandler struct {
	userRoomChamberAccessService *services.UserRoomChamberAccessService
}

// NewUserRoomChamberAccessHandler creates a new user room chamber access handler
func NewUserRoomChamberAccessHandler(userRoomChamberAccessService *services.UserRoomChamberAccessService) *UserRoomChamberAccessHandler {
	return &UserRoomChamberAccessHandler{
		userRoomChamberAccessService: userRoomChamberAccessService,
	}
}

// SetUserRoomChamberAccess handles PUT /users/:id/room-chambers
// Sets room chamber access for a user (replaces all existing access)
func (h *UserRoomChamberAccessHandler) SetUserRoomChamberAccess(c *gin.Context) {
	userID := c.Param("id")

	var req struct {
		RoomChamberIDs []string `json:"room_chamber_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(err.Error()))
		return
	}

	err := h.userRoomChamberAccessService.SetUserRoomChamberAccess(userID, req.RoomChamberIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse("Room chamber access updated successfully"))
}

// GetUserRoomChamberAccess handles GET /users/:id/room-chambers
// Gets all room chambers a user has access to
func (h *UserRoomChamberAccessHandler) GetUserRoomChamberAccess(c *gin.Context) {
	userIDStr := c.Param("id")

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Invalid user ID"))
		return
	}

	roomChambers, err := h.userRoomChamberAccessService.GetUserRoomChamberAccess(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(roomChambers))
}

// GrantRoomChamberAccess handles POST /users/:id/room-chambers/:room_chamber_id
// Grants room chamber access to a user
func (h *UserRoomChamberAccessHandler) GrantRoomChamberAccess(c *gin.Context) {
	userIDStr := c.Param("id")
	roomChamberIDStr := c.Param("room_chamber_id")

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Invalid user ID"))
		return
	}

	roomChamberID, err := primitive.ObjectIDFromHex(roomChamberIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Invalid room chamber ID"))
		return
	}

	err = h.userRoomChamberAccessService.GrantRoomChamberAccess(userID, roomChamberID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse("Room chamber access granted successfully"))
}

// RevokeRoomChamberAccess handles DELETE /users/:id/room-chambers/:room_chamber_id
// Revokes room chamber access from a user
func (h *UserRoomChamberAccessHandler) RevokeRoomChamberAccess(c *gin.Context) {
	userIDStr := c.Param("id")
	roomChamberIDStr := c.Param("room_chamber_id")

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Invalid user ID"))
		return
	}

	roomChamberID, err := primitive.ObjectIDFromHex(roomChamberIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Invalid room chamber ID"))
		return
	}

	err = h.userRoomChamberAccessService.RevokeRoomChamberAccess(userID, roomChamberID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse("Room chamber access revoked successfully"))
}

// GetAllUsersWithRoomChamberAccess handles GET /users/room-chambers
// Gets all users with their room chamber access
func (h *UserRoomChamberAccessHandler) GetAllUsersWithRoomChamberAccess(c *gin.Context) {
	usersWithAccess, err := h.userRoomChamberAccessService.GetAllUsersWithRoomChamberAccess()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(usersWithAccess))
}

// HasRoomChamberAccess handles GET /users/:id/room-chambers/:room_chamber_id/check
// Checks if user has access to specific room chamber
func (h *UserRoomChamberAccessHandler) HasRoomChamberAccess(c *gin.Context) {
	userIDStr := c.Param("id")
	roomChamberIDStr := c.Param("room_chamber_id")

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Invalid user ID"))
		return
	}

	roomChamberID, err := primitive.ObjectIDFromHex(roomChamberIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Invalid room chamber ID"))
		return
	}

	hasAccess, err := h.userRoomChamberAccessService.HasRoomChamberAccess(userID, roomChamberID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
		"has_access": hasAccess,
	}))
}
