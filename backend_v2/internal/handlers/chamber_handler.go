package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"backend_v2/internal/models"
	"backend_v2/internal/services"
)

// ChamberHandler handles chamber-related HTTP requests
type ChamberHandler struct {
	chamberService *services.ChamberService
}

// NewChamberHandler creates a new chamber handler
func NewChamberHandler(chamberService *services.ChamberService) *ChamberHandler {
	return &ChamberHandler{
		chamberService: chamberService,
	}
}

// RegisterChamber handles POST /chambers
func (h *ChamberHandler) RegisterChamber(c *gin.Context) {
	var req services.RegisterChamberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(err.Error()))
		return
	}

	chamber, err := h.chamberService.RegisterChamber(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{"id": chamber.ID.Hex()}))
}

// Heartbeat handles POST /chambers/:id/heartbeat
func (h *ChamberHandler) Heartbeat(c *gin.Context) {
	chamberID := c.Param("id")

	err := h.chamberService.UpdateHeartbeat(chamberID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse("Heartbeat received"))
}

// GetChamber handles GET /chambers/:id
func (h *ChamberHandler) GetChamber(c *gin.Context) {
	chamberID := c.Param("id")

	chamber, err := h.chamberService.GetChamber(chamberID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(chamber))
}

// GetChambers handles GET /chambers
func (h *ChamberHandler) GetChambers(c *gin.Context) {
	chambers, err := h.chamberService.GetChambers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(chambers))
}

// GetChamberWateringZones handles GET /chambers/:id/watering-zones
func (h *ChamberHandler) GetChamberWateringZones(c *gin.Context) {
	chamberID := c.Param("id")

	chamber, err := h.chamberService.GetChamber(chamberID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(err.Error()))
		return
	}

	// Build response with watering zones and their associated input numbers
	type WateringZoneResponse struct {
		Zone         models.WateringZone            `json:"zone"`
		InputNumbers map[string]*models.InputNumber `json:"input_numbers"`
	}

	var response []WateringZoneResponse
	wateringInputs := chamber.GetWateringInputNumbers()

	for _, zone := range chamber.WateringZones {
		if inputs, ok := wateringInputs[zone.Name]; ok {
			response = append(response, WateringZoneResponse{
				Zone:         zone,
				InputNumbers: inputs,
			})
		}
	}

	c.JSON(http.StatusOK, models.SuccessResponse(response))
}
