package services

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"backend_v2/internal/database"
	"backend_v2/internal/models"
)

// ExperimentService handles experiment-related business logic
type ExperimentService struct {
	db *database.MongoDB
}

// NewExperimentService creates a new experiment service
func NewExperimentService(db *database.MongoDB) *ExperimentService {
	return &ExperimentService{
		db: db,
	}
}

// CreateExperiment creates a new experiment
func (s *ExperimentService) CreateExperiment(req *CreateExperimentRequest) (*models.Experiment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Validate chamber exists
	chamberID, err := primitive.ObjectIDFromHex(req.ChamberID)
	if err != nil {
		return nil, fmt.Errorf("invalid chamber ID: %v", err)
	}

	count, err := s.db.ChambersCollection.CountDocuments(ctx, bson.M{"_id": chamberID})
	if err != nil {
		return nil, fmt.Errorf("failed to check chamber: %v", err)
	}
	if count == 0 {
		return nil, fmt.Errorf("chamber not found")
	}

	// Create experiment
	experiment := models.Experiment{
		ID:          primitive.NewObjectID(),
		Title:       req.Title,
		Description: req.Description,
		Status:      models.ExperimentStatusDraft,
		ChamberID:   chamberID,
		Phases:      req.Phases,
		Schedule:    req.Schedule,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	_, err = s.db.ExperimentsCollection.InsertOne(ctx, experiment)
	if err != nil {
		return nil, fmt.Errorf("failed to create experiment: %v", err)
	}

	return &experiment, nil
}

// GetExperiment retrieves an experiment by ID
func (s *ExperimentService) GetExperiment(experimentID string) (*models.Experiment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(experimentID)
	if err != nil {
		return nil, fmt.Errorf("invalid experiment ID: %v", err)
	}

	var experiment models.Experiment
	err = s.db.ExperimentsCollection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&experiment)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("experiment not found")
		}
		return nil, fmt.Errorf("failed to get experiment: %v", err)
	}

	return &experiment, nil
}

// GetExperiments retrieves all experiments, optionally filtered by chamber
func (s *ExperimentService) GetExperiments(chamberID string) ([]models.Experiment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{}
	if chamberID != "" {
		objectID, err := primitive.ObjectIDFromHex(chamberID)
		if err != nil {
			return nil, fmt.Errorf("invalid chamber ID: %v", err)
		}
		filter["chamber_id"] = objectID
	}

	opts := options.Find().SetSort(bson.D{primitive.E{Key: "created_at", Value: -1}})
	cursor, err := s.db.ExperimentsCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get experiments: %v", err)
	}
	defer cursor.Close(ctx)

	var experiments []models.Experiment
	if err = cursor.All(ctx, &experiments); err != nil {
		return nil, fmt.Errorf("failed to decode experiments: %v", err)
	}

	return experiments, nil
}

// UpdateExperiment updates an experiment
func (s *ExperimentService) UpdateExperiment(experimentID string, req *UpdateExperimentRequest) (*models.Experiment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(experimentID)
	if err != nil {
		return nil, fmt.Errorf("invalid experiment ID: %v", err)
	}

	update := bson.M{
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	// Add fields to update
	if req.Title != "" {
		update["$set"].(bson.M)["title"] = req.Title
	}
	if req.Description != "" {
		update["$set"].(bson.M)["description"] = req.Description
	}
	if req.Status != "" {
		update["$set"].(bson.M)["status"] = req.Status
	}
	if req.Phases != nil {
		update["$set"].(bson.M)["phases"] = req.Phases
	}
	if req.Schedule != nil {
		update["$set"].(bson.M)["schedule"] = req.Schedule
	}
	if req.ActivePhaseIndex != nil {
		update["$set"].(bson.M)["active_phase_index"] = req.ActivePhaseIndex
	}

	result, err := s.db.ExperimentsCollection.UpdateByID(ctx, objectID, update)
	if err != nil {
		return nil, fmt.Errorf("failed to update experiment: %v", err)
	}

	if result.MatchedCount == 0 {
		return nil, fmt.Errorf("experiment not found")
	}

	// Return updated experiment
	var experiment models.Experiment
	err = s.db.ExperimentsCollection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&experiment)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated experiment: %v", err)
	}

	return &experiment, nil
}

// DeleteExperiment deletes an experiment
func (s *ExperimentService) DeleteExperiment(experimentID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(experimentID)
	if err != nil {
		return fmt.Errorf("invalid experiment ID: %v", err)
	}

	result, err := s.db.ExperimentsCollection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return fmt.Errorf("failed to delete experiment: %v", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("experiment not found")
	}

	return nil
}

// CreateExperimentRequest represents the request to create an experiment
type CreateExperimentRequest struct {
	Title       string                `json:"title" binding:"required"`
	Description string                `json:"description"`
	ChamberID   string                `json:"chamber_id" binding:"required"`
	Phases      []models.Phase        `json:"phases"`
	Schedule    []models.ScheduleItem `json:"schedule"`
}

// UpdateExperimentRequest represents the request to update an experiment
type UpdateExperimentRequest struct {
	Title            string                  `json:"title"`
	Description      string                  `json:"description"`
	Status           models.ExperimentStatus `json:"status"`
	Phases           []models.Phase          `json:"phases"`
	Schedule         []models.ScheduleItem   `json:"schedule"`
	ActivePhaseIndex *int                    `json:"active_phase_index"`
}
