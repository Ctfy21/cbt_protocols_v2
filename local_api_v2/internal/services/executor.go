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
	currentPhase, phaseIndex, currentDay := s.getCurrentPhaseWithDay(exp)
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
	return s.applyPhaseSettings(currentPhase, currentDay)
}

// getCurrentPhase determines which phase should be active based on the schedule
func (s *ExecutorService) getCurrentPhaseWithDay(exp *models.Experiment) (*models.Phase, int, int) {
	now := time.Now()

	for _, scheduleItem := range exp.Schedule {
		// Parse schedule timestamps
		startTime := time.Unix(scheduleItem.StartTimestamp, 0)
		endTime := time.Unix(scheduleItem.EndTimestamp, 0)
		currentDay := 0

		if now.After(startTime) && now.Before(endTime) {
			// Find the corresponding phase
			if scheduleItem.PhaseIndex < len(exp.Phases) {
				intervalDays := getDaysAndTimestamps(scheduleItem.StartTimestamp, scheduleItem.EndTimestamp)
				for i := 0; i < len(intervalDays)-1; i++ {
					if intervalDays[i].Timestamp < now.Unix() && intervalDays[i+1].Timestamp > now.Unix() {
						currentDay = intervalDays[i+1].Day
					}
				}
				return &exp.Phases[scheduleItem.PhaseIndex], scheduleItem.PhaseIndex, currentDay
			}
		}
	}

	return nil, -1, -1
}

// applyPhaseSettings applies the phase settings to Home Assistant
func (s *ExecutorService) applyPhaseSettings(phase *models.Phase, currentDay int) error {
	var errors []error

	log.Printf("Applying phase settings for phase %s, day %d", phase.Title, currentDay)

	if err := s.haClient.SetInputNumber(, phase.StartDay); err != nil {
		log.Printf("Failed to set %s: %v", scheduleConfig.EntityID, err)
	}

	// Apply climate controls based on day/night
	if err := s.applyClimateControls(phase, currentDay); err != nil {
		errors = append(errors, fmt.Errorf("climate controls: %w", err))
	}

	// Apply light controls
	if err := s.applyLightControls(phase, currentDay); err != nil {
		errors = append(errors, fmt.Errorf("light controls: %w", err))
	}

	// Apply watering controls
	if err := s.applyWateringControls(phase, currentDay); err != nil {
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
func (s *ExecutorService) applyClimateControls(phase *models.Phase, currentDay int) error {

	// Find matching entities from chamber's InputNumbers
	for _, scheduleConfig := range phase.WorkDaySchedule {
		if err := s.haClient.SetInputNumber(scheduleConfig.EntityID, scheduleConfig.Schedule[currentDay]); err != nil {
			log.Printf("Failed to set %s: %v", scheduleConfig.EntityID, err)
		}
	}

	// Find matching entities from chamber's InputNumbers
	for _, scheduleConfig := range phase.TemperatureDaySchedule {
		if err := s.haClient.SetInputNumber(scheduleConfig.EntityID, scheduleConfig.Schedule[currentDay]); err != nil {
			log.Printf("Failed to set %s: %v", scheduleConfig.EntityID, err)
		}
	}

	// Find matching entities from chamber's InputNumbers
	for _, scheduleConfig := range phase.TemperatureNightSchedule {
		if err := s.haClient.SetInputNumber(scheduleConfig.EntityID, scheduleConfig.Schedule[currentDay]); err != nil {
			log.Printf("Failed to set %s: %v", scheduleConfig.EntityID, err)
		}
	}

	for _, scheduleConfig := range phase.HumidityDaySchedule {
		if err := s.haClient.SetInputNumber(scheduleConfig.EntityID, scheduleConfig.Schedule[currentDay]); err != nil {
			log.Printf("Failed to set %s: %v", scheduleConfig.EntityID, err)
		}
	}

	for _, scheduleConfig := range phase.HumidityNightSchedule {
		if err := s.haClient.SetInputNumber(scheduleConfig.EntityID, scheduleConfig.Schedule[currentDay]); err != nil {
			log.Printf("Failed to set %s: %v", scheduleConfig.EntityID, err)
		}
	}

	for _, scheduleConfig := range phase.CO2DaySchedule {
		if err := s.haClient.SetInputNumber(scheduleConfig.EntityID, scheduleConfig.Schedule[currentDay]); err != nil {
			log.Printf("Failed to set %s: %v", scheduleConfig.EntityID, err)
		}
	}

	for _, scheduleConfig := range phase.CO2NightSchedule {
		if err := s.haClient.SetInputNumber(scheduleConfig.EntityID, scheduleConfig.Schedule[currentDay]); err != nil {
			log.Printf("Failed to set %s: %v", scheduleConfig.EntityID, err)
		}
	}

	return nil
}

// applyLightControls applies light intensity settings for each lamp
func (s *ExecutorService) applyLightControls(phase *models.Phase, currentDay int) error {

	// Find matching entities from chamber's InputNumbers
	for _, scheduleConfig := range phase.LightIntensitySchedule {
		if err := s.haClient.SetInputNumber(scheduleConfig.EntityID, scheduleConfig.Schedule[currentDay]); err != nil {
			log.Printf("Failed to set %s: %v", scheduleConfig.EntityID, err)
		}
	}

	return nil
}

// applyWateringControls handles watering zone controls
func (s *ExecutorService) applyWateringControls(phase *models.Phase, currentDay int) error {

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

// getDaysAndTimestamps generates an array of days and timestamps between start and end timestamps
func getDaysAndTimestamps(startTimestamp, endTimestamp int64) []models.DayAndTimestamp {
	var result []models.DayAndTimestamp

	// Calculate number of days
	startDay := startTimestamp / 86400
	endDay := endTimestamp / 86400

	// Generate array of days and timestamps
	for day := startDay; day <= endDay; day++ {
		timestamp := day * 86400 // Convert day back to timestamp
		if timestamp >= startTimestamp && timestamp <= endTimestamp {
			result = append(result, models.DayAndTimestamp{
				Day:       int(day - startDay),
				Timestamp: timestamp,
			})
		}
	}

	return result
}
