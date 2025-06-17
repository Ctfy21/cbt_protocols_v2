package handlers

import (
	"net/http"
	"time"

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

// UpdateChamberConfig handles PUT /chambers/:id/config
func (h *ChamberHandler) UpdateChamberConfig(c *gin.Context) {
	chamberID := c.Param("id")

	var req services.UpdateChamberConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(err.Error()))
		return
	}

	config, err := h.chamberService.UpdateChamberConfig(chamberID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(config))
}

// GetChamberConfig handles GET /chambers/:id/config
func (h *ChamberHandler) GetChamberConfig(c *gin.Context) {
	chamberID := c.Param("id")

	config, err := h.chamberService.GetChamberConfig(chamberID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(config))
}

func (h *ChamberHandler) CheckChamberConfigUpdate(c *gin.Context) {
	chamberID := c.Param("id")

	// Get If-Modified-Since header
	ifModifiedSince := c.GetHeader("If-Modified-Since")

	// Get chamber to check config timestamp
	chamber, err := h.chamberService.GetChamber(chamberID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(err.Error()))
		return
	}

	// Initialize config if it doesn't exist
	if chamber.Config == nil {
		chamber.InitializeConfig()
		// Return 200 to indicate config needs to be fetched
		c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
			"needs_update": true,
			"reason":       "no_config",
		}))
		return
	}

	// Parse If-Modified-Since header
	if ifModifiedSince != "" {
		lastSyncTime, err := time.Parse(http.TimeFormat, ifModifiedSince)
		if err == nil {
			// Compare with config update time
			if !chamber.Config.UpdatedAt.After(lastSyncTime) {
				// Config hasn't changed
				c.Status(http.StatusNotModified)
				return
			}
		}
	}

	// Config has been updated or no sync time provided
	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
		"needs_update": true,
		"updated_at":   chamber.Config.UpdatedAt,
	}))
}
