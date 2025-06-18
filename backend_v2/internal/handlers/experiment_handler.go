package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"backend_v2/internal/models"
	"backend_v2/internal/services"
)

// ExperimentHandler handles experiment-related HTTP requests
type ExperimentHandler struct {
	experimentService *services.ExperimentService
}

// NewExperimentHandler creates a new experiment handler
func NewExperimentHandler(experimentService *services.ExperimentService) *ExperimentHandler {
	return &ExperimentHandler{
		experimentService: experimentService,
	}
}

// CreateExperiment handles POST /experiments
func (h *ExperimentHandler) CreateExperiment(c *gin.Context) {
	var req services.CreateExperimentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(err.Error()))
		return
	}

	experiment, err := h.experimentService.CreateExperiment(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, models.SuccessResponse(experiment))
}

// GetExperiment handles GET /experiments/:id
func (h *ExperimentHandler) GetExperiment(c *gin.Context) {
	experimentID := c.Param("id")

	experiment, err := h.experimentService.GetExperiment(experimentID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(experiment))
}

// GetExperiments handles GET /experiments
func (h *ExperimentHandler) GetExperiments(c *gin.Context) {
	chamberID := c.Query("chamber_id")

	experiments, err := h.experimentService.GetExperiments(chamberID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(experiments))
}

// UpdateExperiment handles PUT /experiments/:id
func (h *ExperimentHandler) UpdateExperiment(c *gin.Context) {
	experimentID := c.Param("id")

	var req services.UpdateExperimentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(err.Error()))
		return
	}

	experiment, err := h.experimentService.UpdateExperiment(experimentID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(experiment))
}

// DeleteExperiment handles DELETE /experiments/:id
func (h *ExperimentHandler) DeleteExperiment(c *gin.Context) {
	experimentID := c.Param("id")

	err := h.experimentService.DeleteExperiment(experimentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse("Experiment deleted successfully"))
}

// UpdateExperimentStatus handles PATCH /experiments/:id/status
func (h *ExperimentHandler) UpdateExperimentStatus(c *gin.Context) {
	experimentID := c.Param("id")

	var req struct {
		Status models.ExperimentStatus `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(err.Error()))
		return
	}

	experiment, err := h.experimentService.UpdateExperimentStatus(experimentID, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(experiment))
}
