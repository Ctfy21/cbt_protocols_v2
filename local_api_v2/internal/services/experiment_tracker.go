package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"local_api_v2/internal/config"
	"local_api_v2/internal/database"
	"local_api_v2/internal/models"
	"local_api_v2/pkg/ntp"
)

// ExperimentTracker handles experiment lifecycle tracking
type ExperimentTracker struct {
	config     *config.Config
	db         *database.MongoDB
	ntpService *ntp.TimeService
	httpClient *http.Client
}

// NewExperimentTracker creates a new experiment tracker
func NewExperimentTracker(cfg *config.Config, db *database.MongoDB, ntpService *ntp.TimeService) *ExperimentTracker {
	return &ExperimentTracker{
		config:     cfg,
		db:         db,
		ntpService: ntpService,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// StartTracking starts the experiment tracking service
func (et *ExperimentTracker) StartTracking(ctx context.Context) {
	log.Println("🔍 Starting experiment tracking service...")

	// Initial check
	if err := et.checkAndUpdateExperiments(); err != nil {
		log.Printf("❌ Initial experiment check failed: %v", err)
	}

	// Periodic check every 2 minutes (more frequent than frontend)
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("🛑 Experiment tracking service stopped")
			return
		case <-ticker.C:
			if err := et.checkAndUpdateExperiments(); err != nil {
				log.Printf("❌ Experiment check failed: %v", err)
			}
		}
	}
}

// checkAndUpdateExperiments checks for completed experiments and updates their status
func (et *ExperimentTracker) checkAndUpdateExperiments() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get all active experiments
	cursor, err := et.db.ExperimentsCollection.Find(
		ctx,
		bson.M{"status": models.StatusActive},
	)
	if err != nil {
		return fmt.Errorf("failed to find active experiments: %v", err)
	}
	defer cursor.Close(ctx)

	var experiments []models.Experiment
	if err := cursor.All(ctx, &experiments); err != nil {
		return fmt.Errorf("failed to decode experiments: %v", err)
	}

	if len(experiments) == 0 {
		return nil
	}

	log.Printf("🔍 Checking %d active experiments...", len(experiments))

	completedCount := 0
	now := et.ntpService.NowInMoscow()

	for _, experiment := range experiments {
		isCompleted := et.isExperimentCompleted(experiment, now)

		if isCompleted {
			log.Printf("✅ Experiment completed: %s", experiment.Title)

			if err := et.completeExperiment(ctx, &experiment); err != nil {
				log.Printf("❌ Failed to complete experiment %s: %v", experiment.Title, err)
				continue
			}

			completedCount++
		}
	}

	if completedCount > 0 {
		log.Printf("🎯 Completed %d experiments", completedCount)
	}

	return nil
}

// isExperimentCompleted checks if an experiment should be marked as completed
func (et *ExperimentTracker) isExperimentCompleted(experiment models.Experiment, now time.Time) bool {
	if len(experiment.Schedule) == 0 {
		return false
	}

	// Find the latest end timestamp
	var latestEndTime int64 = 0
	for _, scheduleItem := range experiment.Schedule {
		if scheduleItem.EndTimestamp > latestEndTime {
			latestEndTime = scheduleItem.EndTimestamp
		}
	}

	if latestEndTime == 0 {
		return false
	}

	// Convert to time and add grace period (5 minutes)
	endTime := time.Unix(latestEndTime, 0)
	gracePeriod := 5 * time.Minute
	completionTime := endTime.Add(gracePeriod)

	// Check if current time has passed the completion time
	return now.After(completionTime)
}

// completeExperiment marks an experiment as completed and syncs with backend
func (et *ExperimentTracker) completeExperiment(ctx context.Context, experiment *models.Experiment) error {
	// Update local status
	experiment.Status = models.StatusCompleted
	experiment.UpdatedAt = et.ntpService.NowInMoscow()

	// Sync status with backend
	if err := et.syncExperimentStatusToBackend(experiment); err != nil {
		log.Printf("⚠️  Failed to sync experiment status to backend: %v", err)
		// Don't return error - local update succeeded
	}

	// Update in local database
	_, err := et.db.ExperimentsCollection.UpdateByID(
		ctx,
		experiment.ID,
		bson.M{
			"$set": bson.M{
				"status":     models.StatusCompleted,
				"updated_at": experiment.UpdatedAt,
			},
		},
	)

	if err != nil {
		return fmt.Errorf("failed to update local experiment: %v", err)
	}

	log.Printf("✅ Successfully completed experiment: %s", experiment.Title)
	return nil
}

// syncExperimentStatusToBackend sends status update to the backend
func (et *ExperimentTracker) syncExperimentStatusToBackend(experiment *models.Experiment) error {
	if experiment.BackendID.IsZero() {
		return fmt.Errorf("experiment has no backend ID")
	}

	url := fmt.Sprintf("%s/experiments/%s/status", et.config.BackendURL, experiment.BackendID.Hex())

	payload := map[string]interface{}{
		"status": experiment.Status,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %v", err)
	}

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if et.config.BackendAPIKey != "" {
		req.Header.Set("Authorization", "Bearer "+et.config.BackendAPIKey)
	}

	// Add local chamber information
	req.Header.Set("X-Local-Time", et.ntpService.NowInMoscow().Format("2006-01-02T15:04:05Z07:00"))
	req.Header.Set("X-Chamber-Name", experiment.ChamberName)

	resp, err := et.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("backend returned status %d", resp.StatusCode)
	}

	log.Printf("🔄 Synced experiment status to backend: %s -> %s", experiment.Title, experiment.Status)
	return nil
}

// GetExperimentProgress returns progress information for an experiment
func (et *ExperimentTracker) GetExperimentProgress(experiment models.Experiment) ExperimentProgress {
	now := et.ntpService.NowInMoscow()

	if len(experiment.Schedule) == 0 {
		return ExperimentProgress{
			CurrentPhase:    -1,
			ProgressPercent: 0,
			TimeRemaining:   0,
			IsCompleted:     false,
		}
	}

	// Find current phase
	currentPhase := -1
	for _, scheduleItem := range experiment.Schedule {
		startTime := time.Unix(scheduleItem.StartTimestamp, 0)
		endTime := time.Unix(scheduleItem.EndTimestamp, 0)

		if now.After(startTime) && now.Before(endTime) {
			currentPhase = scheduleItem.PhaseIndex
			break
		}
	}

	// Calculate total experiment duration
	var startTime, endTime time.Time
	if len(experiment.Schedule) > 0 {
		startTime = time.Unix(experiment.Schedule[0].StartTimestamp, 0)

		var latestEnd int64 = 0
		for _, item := range experiment.Schedule {
			if item.EndTimestamp > latestEnd {
				latestEnd = item.EndTimestamp
			}
		}
		endTime = time.Unix(latestEnd, 0)
	}

	// Calculate progress
	var progressPercent float64 = 0
	var timeRemaining time.Duration = 0
	isCompleted := false

	if !startTime.IsZero() && !endTime.IsZero() {
		totalDuration := endTime.Sub(startTime)
		elapsed := now.Sub(startTime)

		if elapsed < 0 {
			// Experiment hasn't started yet
			progressPercent = 0
			timeRemaining = totalDuration
		} else if elapsed >= totalDuration {
			// Experiment is completed or overdue
			progressPercent = 100
			timeRemaining = 0
			isCompleted = true
		} else {
			// Experiment is in progress
			progressPercent = (elapsed.Seconds() / totalDuration.Seconds()) * 100
			timeRemaining = totalDuration - elapsed
		}
	}

	return ExperimentProgress{
		CurrentPhase:    currentPhase,
		ProgressPercent: progressPercent,
		TimeRemaining:   timeRemaining,
		IsCompleted:     isCompleted,
	}
}

// ExperimentProgress represents the progress of an experiment
type ExperimentProgress struct {
	CurrentPhase    int           `json:"current_phase"`
	ProgressPercent float64       `json:"progress_percent"`
	TimeRemaining   time.Duration `json:"time_remaining"`
	IsCompleted     bool          `json:"is_completed"`
}

// GetActiveExperimentsWithProgress returns all active experiments with their progress
func (et *ExperimentTracker) GetActiveExperimentsWithProgress() ([]ExperimentWithProgress, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := et.db.ExperimentsCollection.Find(
		ctx,
		bson.M{"status": models.StatusActive},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find active experiments: %v", err)
	}
	defer cursor.Close(ctx)

	var experiments []models.Experiment
	if err := cursor.All(ctx, &experiments); err != nil {
		return nil, fmt.Errorf("failed to decode experiments: %v", err)
	}

	result := make([]ExperimentWithProgress, len(experiments))
	for i, experiment := range experiments {
		progress := et.GetExperimentProgress(experiment)
		result[i] = ExperimentWithProgress{
			Experiment: experiment,
			Progress:   progress,
		}
	}

	return result, nil
}

// ExperimentWithProgress combines experiment data with progress information
type ExperimentWithProgress struct {
	Experiment models.Experiment  `json:"experiment"`
	Progress   ExperimentProgress `json:"progress"`
}

// GetTrackingStatus returns the current status of the tracking service
func (et *ExperimentTracker) GetTrackingStatus() map[string]interface{} {
	activeExperiments, _ := et.GetActiveExperimentsWithProgress()

	completingSoon := 0
	for _, exp := range activeExperiments {
		if exp.Progress.TimeRemaining < 24*time.Hour && exp.Progress.TimeRemaining > 0 {
			completingSoon++
		}
	}

	return map[string]interface{}{
		"active_experiments":   len(activeExperiments),
		"completing_soon_24h":  completingSoon,
		"last_check_time":      et.ntpService.NowInMoscow().Format("2006-01-02T15:04:05Z07:00"),
		"tracking_enabled":     true,
		"backend_sync_enabled": et.config.BackendURL != "",
	}
}
