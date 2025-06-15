package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"backend_v2/internal/models"
	"backend_v2/internal/services"
)

// RoomChamberHandler handles room chamber-related HTTP requests
type RoomChamberHandler struct {
	roomChamberService *services.RoomChamberService
}

// NewRoomChamberHandler creates a new room chamber handler
func NewRoomChamberHandler(roomChamberService *services.RoomChamberService) *RoomChamberHandler {
	return &RoomChamberHandler{
		roomChamberService: roomChamberService,
	}
}

// RegisterRoomChamber handles POST /room-chambers
func (h *RoomChamberHandler) RegisterRoomChamber(c *gin.Context) {
	var req services.RegisterRoomChamberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(err.Error()))
		return
	}

	roomChamber, err := h.roomChamberService.RegisterRoomChamber(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{"id": roomChamber.ID.Hex()}))
}

// Heartbeat handles POST /room-chambers/:id/heartbeat
func (h *RoomChamberHandler) Heartbeat(c *gin.Context) {
	roomChamberID := c.Param("id")

	err := h.roomChamberService.UpdateRoomChamberHeartbeat(roomChamberID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse("Heartbeat received"))
}

// GetRoomChamber handles GET /room-chambers/:id
func (h *RoomChamberHandler) GetRoomChamber(c *gin.Context) {
	roomChamberID := c.Param("id")

	roomChamber, err := h.roomChamberService.GetRoomChamber(roomChamberID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(roomChamber))
}

// GetRoomChambers handles GET /room-chambers
func (h *RoomChamberHandler) GetRoomChambers(c *gin.Context) {
	parentChamberID := c.Query("parent_chamber_id")

	var roomChambers []models.RoomChamberResponse
	var err error

	if parentChamberID != "" {
		roomChambers, err = h.roomChamberService.GetRoomChambersByParent(parentChamberID)
	} else {
		roomChambers, err = h.roomChamberService.GetRoomChambers()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(roomChambers))
}

// GetRoomChamberWateringZones handles GET /room-chambers/:id/watering-zones
func (h *RoomChamberHandler) GetRoomChamberWateringZones(c *gin.Context) {
	roomChamberID := c.Param("id")

	response, err := h.roomChamberService.GetRoomChamberWateringZones(roomChamberID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(response))
}
