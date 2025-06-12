package services

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"local_api_v2/internal/database"
	"local_api_v2/internal/models"
	"local_api_v2/pkg/homeassistant"

	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/bson"
)

// ExecutorService handles the execution of experiment phases
type ExecutorService struct {
	db        *database.MongoDB
	haClient  *homeassistant.Client
	cron      *cron.Cron
	chamber   *models.Chamber
	mu        sync.RWMutex
	isRunning bool
}

// NewExecutorService creates a new executor service
func NewExecutorService(db *database.MongoDB, haClient *homeassistant.Client, chamber *models.Chamber) *ExecutorService {
	return &ExecutorService{
		db:       db,
		haClient: haClient,
		cron:     cron.New(cron.WithLocation(time.Local)),
		chamber:  chamber,
	}
}

// Start begins the execution loop
func (s *ExecutorService) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.isRunning {
		s.mu.Unlock()
		return fmt.Errorf("executor service already running")
	}
	s.isRunning = true
	s.mu.Unlock()

	// Add job to check and execute phases every minute
	_, err := s.cron.AddFunc("* * * * *", func() {
		if err := s.executeActivePhasesWrapper(ctx); err != nil {
			log.Printf("Error executing phases: %v", err)
		}
	})
	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	// Start the cron scheduler
	s.cron.Start()

	// Run immediately on start
	go func() {
		if err := s.executeActivePhasesWrapper(ctx); err != nil {
			log.Printf("Error executing phases on start: %v", err)
		}
	}()

	log.Println("Executor service started")
	return nil
}

// Stop halts the execution loop
func (s *ExecutorService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return
	}

	ctx := s.cron.Stop()
	<-ctx.Done()
	s.isRunning = false
	log.Println("Executor service stopped")
}

// executeActivePhasesWrapper wraps executeActivePhases with context checking
func (s *ExecutorService) executeActivePhasesWrapper(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return s.executeActivePhases(ctx)
	}
}

// executeActivePhases finds and executes all active experiment phases
func (s *ExecutorService) executeActivePhases(ctx context.Context) error {
	// Get all active experiments
	experiments, err := s.getActiveExperiments(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active experiments: %w", err)
	}

	if len(experiments) == 0 {
		return nil // No active experiments
	}

	log.Printf("Found %d active experiments to execute", len(experiments))

	// Process each experiment
	for _, exp := range experiments {
		if err := s.processExperiment(ctx, &exp); err != nil {
			log.Printf("Error processing experiment %s: %v", exp.Title, err)
			// Continue with other experiments even if one fails
		}
	}

	return nil
}

// getActiveExperiments retrieves all experiments with status "active"
func (s *ExecutorService) getActiveExperiments(ctx context.Context) ([]models.Experiment, error) {
	filter := bson.M{
		"status":     "active",
		"chamber_id": s.chamber.BackendID,
	}

	cursor, err := s.db.ExperimentsCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var experiments []models.Experiment
	if err := cursor.All(ctx, &experiments); err != nil {
		return nil, err
	}

	return experiments, nil
}

// processExperiment processes a single experiment
func (s *ExecutorService) processExperiment(ctx context.Context, exp *models.Experiment) error {
	// Determine current phase based on schedule
	currentPhase, phaseIndex := s.getCurrentPhase(exp)
	if currentPhase == nil {
		log.Printf("No active phase found for experiment %s", exp.Title)
		return nil
	}

	log.Printf("Executing phase %d (%s) for experiment %s", phaseIndex, currentPhase.Title, exp.Title)

	// Update the active phase index if changed
	if exp.ActivePhaseIndex == nil || *exp.ActivePhaseIndex != phaseIndex {
		exp.ActivePhaseIndex = &phaseIndex
		if err := s.updateExperimentActivePhase(ctx, exp); err != nil {
			log.Printf("Failed to update active phase index: %v", err)
		}
	}

	// Apply phase settings to Home Assistant
	return s.applyPhaseSettings(currentPhase)
}

// getCurrentPhase determines which phase should be active based on the schedule
func (s *ExecutorService) getCurrentPhase(exp *models.Experiment) (*models.Phase, int) {
	now := time.Now()

	for _, scheduleItem := range exp.Schedule {
		// Parse schedule timestamps
		startTime := time.Unix(scheduleItem.StartTimestamp, 0)
		endTime := time.Unix(scheduleItem.EndTimestamp, 0)

		if now.After(startTime) && now.Before(endTime) {
			// Find the corresponding phase
			if scheduleItem.PhaseIndex < len(exp.Phases) {
				return &exp.Phases[scheduleItem.PhaseIndex], scheduleItem.PhaseIndex
			}
		}
	}

	return nil, -1
}

// applyPhaseSettings applies the phase settings to Home Assistant
func (s *ExecutorService) applyPhaseSettings(phase *models.Phase) error {
	var errors []error

	// Apply climate controls based on day/night
	if err := s.applyClimateControls(phase); err != nil {
		errors = append(errors, fmt.Errorf("climate controls: %w", err))
	}

	// Apply light controls
	if err := s.applyLightControls(phase); err != nil {
		errors = append(errors, fmt.Errorf("light controls: %w", err))
	}

	// Apply watering controls
	if err := s.applyWateringControls(phase); err != nil {
		errors = append(errors, fmt.Errorf("watering controls: %w", err))
	}

	// Update last executed time
	now := time.Now()
	phase.LastExecuted = &now

	if len(errors) > 0 {
		return fmt.Errorf("multiple errors during phase execution: %v", errors)
	}

	return nil
}

// applyClimateControls applies temperature, humidity, and CO2 settings
func (s *ExecutorService) applyClimateControls(phase *models.Phase) error {

	// Find matching entities from chamber's InputNumbers
	for _, inputNumber := range s.chamber.InputNumbers {

		if phase.InputNumbers[inputNumber.EntityID] != nil {
			if err := s.haClient.SetInputNumber(inputNumber.EntityID, phase.InputNumbers[inputNumber.EntityID].Value); err != nil {
				log.Printf("Failed to set %s: %v", inputNumber.EntityID, err)
			} else {
				log.Printf("Set %s to %.1f", inputNumber.EntityID, phase.InputNumbers[inputNumber.EntityID].Value)
			}
		}
	}

	return nil
}

// applyLightControls applies light intensity settings for each lamp
func (s *ExecutorService) applyLightControls(phase *models.Phase) error {

	for _, lamp := range s.chamber.Lamps {
		if phase.LightIntensity[lamp.EntityID] != nil {
			if err := s.haClient.SetInputNumber(lamp.EntityID, phase.LightIntensity[lamp.EntityID].Intensity); err != nil {
				log.Printf("Failed to set %s: %v", lamp.EntityID, err)
			}
		}
	}

	return nil
}

// applyWateringControls handles watering zone controls
func (s *ExecutorService) applyWateringControls(phase *models.Phase) error {

	return nil
}

// updateExperimentActivePhase updates the active phase index in the database
func (s *ExecutorService) updateExperimentActivePhase(ctx context.Context, exp *models.Experiment) error {
	update := bson.M{
		"$set": bson.M{
			"active_phase_index": exp.ActivePhaseIndex,
			"updated_at":         time.Now(),
		},
	}

	_, err := s.db.ExperimentsCollection.UpdateOne(
		ctx,
		bson.M{"_id": exp.ID},
		update,
	)
	return err
}
