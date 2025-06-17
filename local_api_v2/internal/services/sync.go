package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"local_api_v2/internal/config"
	"local_api_v2/internal/database"
	"local_api_v2/internal/models"
	"local_api_v2/pkg/ntp"
)

// SyncService handles synchronization with backend
type SyncService struct {
	config              *config.Config
	db                  *database.MongoDB
	ntpService          *ntp.TimeService
	httpClient          *http.Client
	chamberManager      *ChamberManager
	registrationService *RegistrationService
}

// NewSyncService creates a new sync service
func NewSyncService(cfg *config.Config, db *database.MongoDB, ntpService *ntp.TimeService) *SyncService {
	return &SyncService{
		config:     cfg,
		db:         db,
		ntpService: ntpService,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetChamberManager sets the chamber manager for config updates
func (s *SyncService) SetChamberManager(cm *ChamberManager) {
	s.chamberManager = cm
}

// SetRegistrationService sets the registration service for chamber ID mapping
func (s *SyncService) SetRegistrationService(rs *RegistrationService) {
	s.registrationService = rs
}

// StartSync starts the periodic synchronization
func (s *SyncService) StartSync(ctx context.Context) {
	// Initial sync
	if err := s.syncAll(); err != nil {
		log.Printf("❌ Initial sync failed: %v", err)
	}

	// Periodic sync every 60 seconds
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Sync service stopped")
			return
		case <-ticker.C:
			if err := s.syncAll(); err != nil {
				log.Printf("❌ Sync failed: %v", err)
			}
		}
	}
}

// syncAll performs all synchronization tasks
func (s *SyncService) syncAll() error {
	// Get registered chambers
	registeredChambers := s.chamberManager.GetRegisteredChambers()
	if len(registeredChambers) == 0 {
		log.Printf("⚠️  No chambers registered with backend yet")
		return nil
	}

	// Sync experiments for each chamber
	for _, chamber := range registeredChambers {
		if err := s.syncExperimentsForChamber(chamber); err != nil {
			log.Printf("Failed to sync experiments for chamber %s: %v", chamber.Name, err)
		}
	}

	// Sync chamber configurations
	if err := s.syncChamberConfigs(); err != nil {
		log.Printf("Failed to sync chamber configs: %v", err)
	}

	return nil
}

// syncChamberConfigs fetches and updates chamber configurations from backend
func (s *SyncService) syncChamberConfigs() error {
	if s.chamberManager == nil {
		return fmt.Errorf("chamber manager not set")
	}

	successCount := 0
	chambers := s.chamberManager.GetChambers()

	for _, chamber := range chambers {
		if chamber.BackendID.IsZero() {
			continue // Skip unregistered chambers
		}

		// Check if config needs update
		needsUpdate, err := s.checkConfigNeedsUpdate(chamber)
		if err != nil {
			log.Printf("Failed to check config update for chamber %s: %v", chamber.Name, err)
			continue
		}

		if !needsUpdate {
			continue
		}

		// Fetch updated config from backend
		config, err := s.fetchChamberConfig(chamber.BackendID)
		if err != nil {
			log.Printf("Failed to fetch config for chamber %s: %v", chamber.Name, err)
			continue
		}

		// Update local chamber config
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		if err := s.chamberManager.UpdateChamberConfig(ctx, chamber.ID, config); err != nil {
			cancel()
			log.Printf("Failed to update config for chamber %s: %v", chamber.Name, err)
			continue
		}
		cancel()

		successCount++
		log.Printf("📊 Synced configuration for chamber: %s", chamber.Name)
	}

	if successCount > 0 {
		log.Printf("✅ Successfully synced configs for %d chambers", successCount)
	}

	return nil
}

// checkConfigNeedsUpdate checks if chamber config needs to be updated
func (s *SyncService) checkConfigNeedsUpdate(chamber *models.Chamber) (bool, error) {
	url := fmt.Sprintf("%s/chambers/%s/config/check", s.config.BackendURL, chamber.BackendID.Hex())

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %v", err)
	}

	// Add last sync timestamp if available
	if chamber.Config.SyncedAt != nil {
		req.Header.Set("If-Modified-Since", chamber.Config.SyncedAt.Format(http.TimeFormat))
	}

	if s.config.BackendAPIKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.config.BackendAPIKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to check config: %v", err)
	}
	defer resp.Body.Close()

	// 304 Not Modified means config hasn't changed
	if resp.StatusCode == http.StatusNotModified {
		return false, nil
	}

	// 200 OK means config has been updated
	if resp.StatusCode == http.StatusOK {
		return true, nil
	}

	body, _ := io.ReadAll(resp.Body)
	return false, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
}

// fetchChamberConfig fetches chamber configuration from backend
func (s *SyncService) fetchChamberConfig(backendID primitive.ObjectID) (*models.ChamberConfig, error) {
	url := fmt.Sprintf("%s/chambers/%s/config", s.config.BackendURL, backendID.Hex())

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	if s.config.BackendAPIKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.config.BackendAPIKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch config: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("backend returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var response struct {
		Success bool                 `json:"success"`
		Data    models.ChamberConfig `json:"data"`
		Error   string               `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("config fetch failed: %s", response.Error)
	}

	return &response.Data, nil
}

// syncExperimentsForChamber fetches experiments for a specific chamber
func (s *SyncService) syncExperimentsForChamber(chamber *models.Chamber) error {
	// Fetch experiments from backend
	url := fmt.Sprintf("%s/experiments?chamber_id=%s", s.config.BackendURL, chamber.BackendID.Hex())
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// Add NTP timing information to request headers
	req.Header.Set("X-Local-Time", s.ntpService.NowInMoscow().Format("2006-01-02T15:04:05Z07:00"))
	req.Header.Set("X-NTP-Enabled", fmt.Sprintf("%t", s.ntpService.IsEnabled()))
	req.Header.Set("X-NTP-Connected", fmt.Sprintf("%t", s.ntpService.IsConnected()))

	if s.config.BackendAPIKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.config.BackendAPIKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch experiments: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("backend returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var response struct {
		Success bool                `json:"success"`
		Data    []models.Experiment `json:"data"`
		Error   string              `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	if !response.Success {
		return fmt.Errorf("sync failed: %s", response.Error)
	}

	// Update local database
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	syncedCount := 0
	now := s.ntpService.NowInMoscow()

	for _, experiment := range response.Data {
		// Store backend ID and chamber info
		backendID := experiment.ID
		experiment.BackendID = backendID
		experiment.ID = primitive.ObjectID{} // Clear ID for local storage
		experiment.ChamberID = chamber.ID
		experiment.ChamberName = chamber.Name
		experiment.SyncedAt = now

		// Check if experiment already exists
		var existingExperiment models.Experiment
		err := s.db.ExperimentsCollection.FindOne(
			ctx,
			bson.M{
				"backend_id": backendID,
				"chamber_id": chamber.ID,
			},
		).Decode(&existingExperiment)

		if err == nil {
			// Update existing experiment
			experiment.ID = existingExperiment.ID
			experiment.UpdatedAt = now
			_, err = s.db.ExperimentsCollection.ReplaceOne(
				ctx,
				bson.M{"_id": existingExperiment.ID},
				experiment,
			)
			if err != nil {
				log.Printf("Failed to update experiment %s: %v", experiment.Title, err)
				continue
			}
		} else {
			// Insert new experiment
			experiment.ID = primitive.NewObjectID()
			experiment.CreatedAt = now
			experiment.UpdatedAt = now
			_, err = s.db.ExperimentsCollection.InsertOne(ctx, experiment)
			if err != nil {
				log.Printf("Failed to insert experiment %s: %v", experiment.Title, err)
				continue
			}
		}

		syncedCount++

		// Log active experiments
		if experiment.Status == models.StatusActive {
			log.Printf("🔄 Active experiment for chamber %s: %s", chamber.Name, experiment.Title)
		}
	}

	log.Printf("📊 Synced %d experiments for chamber %s", syncedCount, chamber.Name)
	return nil
}

// GetActiveExperiments returns all active experiments
func (s *SyncService) GetActiveExperiments() ([]models.Experiment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := s.db.ExperimentsCollection.Find(
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

	return experiments, nil
}

// GetSyncStatus returns sync service status information
func (s *SyncService) GetSyncStatus() map[string]interface{} {
	activeExperiments, _ := s.GetActiveExperiments()
	registeredChambers := s.chamberManager.GetRegisteredChambers()
	now := s.ntpService.NowInMoscow()

	return map[string]interface{}{
		"registered_chambers": len(registeredChambers),
		"active_experiments":  len(activeExperiments),
		"last_sync_time":      now.Format("2006-01-02T15:04:05Z07:00"),
		"ntp_enabled":         s.ntpService.IsEnabled(),
		"ntp_connected":       s.ntpService.IsConnected(),
	}
}
