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
	"local_api_v2/pkg/ntp"

	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ExecutorService handles the execution of experiment phases
type ExecutorService struct {
	db         *database.MongoDB
	haClient   *homeassistant.Client
	ntpService *ntp.TimeService
	cron       *cron.Cron
	chamberID  primitive.ObjectID // ID of the chamber this executor is responsible for
	mu         sync.RWMutex
	isRunning  bool
}

// NewExecutorService creates a new executor service for a specific chamber
func NewExecutorService(db *database.MongoDB, haClient *homeassistant.Client, chamberID primitive.ObjectID, ntpService *ntp.TimeService) *ExecutorService {
	if db == nil || haClient == nil || ntpService == nil {
		log.Printf("Error: Required dependencies are nil - cannot create executor service")
		return nil
	}

	return &ExecutorService{
		db:         db,
		haClient:   haClient,
		ntpService: ntpService,
		cron:       cron.New(cron.WithLocation(time.Local)),
		chamberID:  chamberID,
	}
}

// SetChamberID sets the chamber ID for this executor
func (s *ExecutorService) SetChamberID(chamberID primitive.ObjectID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.chamberID = chamberID
}

// Start begins the execution loop
func (s *ExecutorService) Start(ctx context.Context) error {
	if s == nil {
		return fmt.Errorf("executor service is nil")
	}

	s.mu.Lock()
	if s.isRunning {
		s.mu.Unlock()
		return fmt.Errorf("executor service already running")
	}
	s.isRunning = true
	s.mu.Unlock()

	// Validate that all required components are available
	if s.db == nil {
		return fmt.Errorf("database is not initialized")
	}
	if s.haClient == nil {
		return fmt.Errorf("home assistant client is not initialized")
	}
	if s.cron == nil {
		return fmt.Errorf("cron scheduler is not initialized")
	}
	if s.ntpService == nil {
		return fmt.Errorf("NTP service is not initialized")
	}

	// Add job to check and execute phases every minute
	_, err := s.cron.AddFunc("* * * * *", func() {
		if err := s.executeActivePhasesWrapper(ctx); err != nil {
			log.Printf("Error executing phases for chamber %s: %v", s.chamberID.Hex(), err)
		}
	})
	if err != nil {
		s.mu.Lock()
		s.isRunning = false
		s.mu.Unlock()
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	// Start the cron scheduler
	s.cron.Start()

	// Run immediately on start
	go func() {
		if err := s.executeActivePhasesWrapper(ctx); err != nil {
			log.Printf("Error executing phases on start for chamber %s: %v", s.chamberID.Hex(), err)
		}
	}()

	timeSource := "system"
	if s.ntpService.IsConnected() {
		timeSource = "NTP"
	}
	log.Printf("Executor service started for chamber %s (using %s time)", s.chamberID.Hex(), timeSource)
	return nil
}

// Stop halts the execution loop
func (s *ExecutorService) Stop() {
	if s == nil {
		log.Printf("Warning: Attempt to stop nil executor service")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return
	}

	if s.cron != nil {
		ctx := s.cron.Stop()
		<-ctx.Done()
	}

	s.isRunning = false
	log.Printf("Executor service stopped for chamber %s", s.chamberID.Hex())
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

// executeActivePhases finds and executes all active experiment phases for this chamber
func (s *ExecutorService) executeActivePhases(ctx context.Context) error {
	// Get all active experiments for this chamber
	experiments, err := s.getActiveExperimentsForChamber(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active experiments: %w", err)
	}

	if len(experiments) == 0 {
		return nil // No active experiments
	}

	log.Printf("Found %d active experiments for chamber %s", len(experiments), s.chamberID.Hex())

	// Process each experiment
	for _, exp := range experiments {
		if err := s.processExperiment(ctx, &exp); err != nil {
			log.Printf("Error processing experiment %s: %v", exp.Title, err)
			// Continue with other experiments even if one fails
		}
	}

	return nil
}

// getActiveExperimentsForChamber retrieves all active experiments for this chamber
func (s *ExecutorService) getActiveExperimentsForChamber(ctx context.Context) ([]models.Experiment, error) {
	filter := bson.M{
		"status":     "active",
		"chamber_id": s.chamberID,
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

// processExperiment processes a single experiment using NTP time
func (s *ExecutorService) processExperiment(ctx context.Context, exp *models.Experiment) error {
	// Determine current phase based on schedule using NTP time
	currentPhase, phaseIndex, currentDay := s.getCurrentPhaseWithDay(exp)
	if currentPhase == nil {
		log.Printf("No active phase found for experiment %s", exp.Title)
		return nil
	}

	timeSource := "system"
	if s.ntpService.IsConnected() {
		timeSource = "NTP"
	}
	log.Printf("Executing phase %d (%s) for experiment %s, day %d (using %s time)",
		phaseIndex, currentPhase.Title, exp.Title, currentDay, timeSource)

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

// getCurrentPhaseWithDay determines which phase should be active based on the schedule using NTP time
func (s *ExecutorService) getCurrentPhaseWithDay(exp *models.Experiment) (*models.Phase, int, int) {
	now := s.ntpService.NowInMoscow()

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

	// Apply start day configurations
	for _, startDayConfig := range phase.StartDay {
		if err := s.haClient.SetInputNumber(startDayConfig.EntityID, startDayConfig.Value); err != nil {
			log.Printf("Failed to set %s: %v", startDayConfig.EntityID, err)
			errors = append(errors, fmt.Errorf("start day config %s: %w", startDayConfig.EntityID, err))
		}
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

	// Update last executed time using NTP time
	now := s.ntpService.NowInMoscow()
	phase.LastExecuted = &now

	if len(errors) > 0 {
		return fmt.Errorf("multiple errors during phase execution: %v", errors)
	}

	return nil
}

// applyClimateControls applies temperature, humidity, and CO2 settings
func (s *ExecutorService) applyClimateControls(phase *models.Phase, currentDay int) error {
	var errors []error

	// Apply work day schedule
	for _, scheduleConfig := range phase.WorkDaySchedule {
		if value, exists := scheduleConfig.Schedule[currentDay]; exists {
			if err := s.haClient.SetInputNumber(scheduleConfig.EntityID, value); err != nil {
				log.Printf("Failed to set %s: %v", scheduleConfig.EntityID, err)
				errors = append(errors, err)
			}
		}
	}

	// Apply temperature day schedule
	for _, scheduleConfig := range phase.TemperatureDaySchedule {
		if value, exists := scheduleConfig.Schedule[currentDay]; exists {
			if err := s.haClient.SetInputNumber(scheduleConfig.EntityID, value); err != nil {
				log.Printf("Failed to set %s: %v", scheduleConfig.EntityID, err)
				errors = append(errors, err)
			}
		}
	}

	// Apply temperature night schedule
	for _, scheduleConfig := range phase.TemperatureNightSchedule {
		if value, exists := scheduleConfig.Schedule[currentDay]; exists {
			if err := s.haClient.SetInputNumber(scheduleConfig.EntityID, value); err != nil {
				log.Printf("Failed to set %s: %v", scheduleConfig.EntityID, err)
				errors = append(errors, err)
			}
		}
	}

	// Apply humidity day schedule
	for _, scheduleConfig := range phase.HumidityDaySchedule {
		if value, exists := scheduleConfig.Schedule[currentDay]; exists {
			if err := s.haClient.SetInputNumber(scheduleConfig.EntityID, value); err != nil {
				log.Printf("Failed to set %s: %v", scheduleConfig.EntityID, err)
				errors = append(errors, err)
			}
		}
	}

	// Apply humidity night schedule
	for _, scheduleConfig := range phase.HumidityNightSchedule {
		if value, exists := scheduleConfig.Schedule[currentDay]; exists {
			if err := s.haClient.SetInputNumber(scheduleConfig.EntityID, value); err != nil {
				log.Printf("Failed to set %s: %v", scheduleConfig.EntityID, err)
				errors = append(errors, err)
			}
		}
	}

	// Apply CO2 day schedule
	for _, scheduleConfig := range phase.CO2DaySchedule {
		if value, exists := scheduleConfig.Schedule[currentDay]; exists {
			if err := s.haClient.SetInputNumber(scheduleConfig.EntityID, value); err != nil {
				log.Printf("Failed to set %s: %v", scheduleConfig.EntityID, err)
				errors = append(errors, err)
			}
		}
	}

	// Apply CO2 night schedule
	for _, scheduleConfig := range phase.CO2NightSchedule {
		if value, exists := scheduleConfig.Schedule[currentDay]; exists {
			if err := s.haClient.SetInputNumber(scheduleConfig.EntityID, value); err != nil {
				log.Printf("Failed to set %s: %v", scheduleConfig.EntityID, err)
				errors = append(errors, err)
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("climate control errors: %v", errors)
	}

	return nil
}

// applyLightControls applies light intensity settings for each lamp
func (s *ExecutorService) applyLightControls(phase *models.Phase, currentDay int) error {
	var errors []error

	// Apply light intensity schedule
	for _, scheduleConfig := range phase.LightIntensitySchedule {
		if value, exists := scheduleConfig.Schedule[currentDay]; exists {
			if err := s.haClient.SetInputNumber(scheduleConfig.EntityID, value); err != nil {
				log.Printf("Failed to set %s: %v", scheduleConfig.EntityID, err)
				errors = append(errors, err)
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("light control errors: %v", errors)
	}

	return nil
}

// applyWateringControls handles watering zone controls
func (s *ExecutorService) applyWateringControls(phase *models.Phase, currentDay int) error {
	var errors []error

	// Apply watering zone schedules
	for _, scheduleConfig := range phase.WateringZones {
		// Start time schedule
		if value, exists := scheduleConfig.StartTimeSchedule[currentDay]; exists {
			if err := s.haClient.SetInputNumber(scheduleConfig.StartTimeEntityID, value); err != nil {
				log.Printf("Failed to set %s: %v", scheduleConfig.StartTimeEntityID, err)
				errors = append(errors, err)
			}
		}

		// Period schedule
		if value, exists := scheduleConfig.PeriodSchedule[currentDay]; exists {
			if err := s.haClient.SetInputNumber(scheduleConfig.PeriodEntityID, value); err != nil {
				log.Printf("Failed to set %s: %v", scheduleConfig.PeriodEntityID, err)
				errors = append(errors, err)
			}
		}

		// Pause between schedule
		if value, exists := scheduleConfig.PauseBetweenSchedule[currentDay]; exists {
			if err := s.haClient.SetInputNumber(scheduleConfig.PauseBetweenEntityID, value); err != nil {
				log.Printf("Failed to set %s: %v", scheduleConfig.PauseBetweenEntityID, err)
				errors = append(errors, err)
			}
		}

		// Duration schedule
		if value, exists := scheduleConfig.DurationSchedule[currentDay]; exists {
			if err := s.haClient.SetInputNumber(scheduleConfig.DurationEntityID, value); err != nil {
				log.Printf("Failed to set %s: %v", scheduleConfig.DurationEntityID, err)
				errors = append(errors, err)
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("watering control errors: %v", errors)
	}

	return nil
}

// updateExperimentActivePhase updates the active phase index in the database using NTP time
func (s *ExecutorService) updateExperimentActivePhase(ctx context.Context, exp *models.Experiment) error {
	now := s.ntpService.NowInMoscow()
	update := bson.M{
		"$set": bson.M{
			"active_phase_index": exp.ActivePhaseIndex,
			"updated_at":         now,
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

// GetStatus returns executor service status
func (s *ExecutorService) GetStatus() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := s.ntpService.NowInMoscow()
	return map[string]interface{}{
		"running":       s.isRunning,
		"chamber_id":    s.chamberID.Hex(),
		"ntp_enabled":   s.ntpService.IsEnabled(),
		"ntp_connected": s.ntpService.IsConnected(),
		"current_time":  now.Format("2006-01-02T15:04:05Z07:00"),
	}
}
