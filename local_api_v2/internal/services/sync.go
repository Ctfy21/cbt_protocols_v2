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

// SyncService handles synchronization of experiments from backend
type SyncService struct {
	config          *config.Config
	db              *database.MongoDB
	ntpService      *ntp.TimeService
	httpClient      *http.Client
	backendServerID primitive.ObjectID
	chamberManager  *ChamberManager
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

// SetBackendID sets the backend chamber ID for syncing experiments
func (s *SyncService) SetBackendServerID(id primitive.ObjectID) {
	s.backendServerID = id
	log.Printf("Sync service: Backend Server ID set to %s", id.Hex())
}

// SetChamberManager sets the chamber manager for config updates
func (s *SyncService) SetChamberManager(cm *ChamberManager) {
	s.chamberManager = cm
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
				// Only log as warning if it's not the "no backend ID" error
				if s.backendServerID.IsZero() {
					log.Printf("⚠️  Sync skipped: Chamber not registered with backend yet")
				} else {
					log.Printf("❌ Sync failed: %v", err)
				}
			}
		}
	}
}

// syncAll performs all synchronization tasks
func (s *SyncService) syncAll() error {
	// Sync experiments
	if err := s.syncExperiments(); err != nil {
		log.Printf("Failed to sync experiments: %v", err)
	}

	// Sync chamber configuration if chamber manager is available
	if s.chamberManager != nil {
		if err := s.syncChamberConfig(); err != nil {
			log.Printf("Failed to sync chamber config: %v", err)
		}
	}

	return nil
}

// syncChamberConfig fetches chamber configuration updates from backend
func (s *SyncService) syncChamberConfig() error {
	if s.backendServerID.IsZero() {
		return fmt.Errorf("no backend ID set - chamber not registered")
	}

	// Fetch chamber config from backend
	url := fmt.Sprintf("%s/chambers/%s/config", s.config.BackendURL, s.backendServerID.Hex())
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	if s.config.BackendAPIKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.config.BackendAPIKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch chamber config: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified {
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("backend returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var response struct {
		Success bool                 `json:"success"`
		Data    models.ChamberConfig `json:"data"`
		Error   string               `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	if !response.Success {
		return fmt.Errorf("config sync failed: %s", response.Error)
	}

	// Update chamber configuration
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Also sync room chamber configs if they have backend IDs
	for _, chamber := range s.chamberManager.GetChambers() {
		// Fetch room chamber specific config
		roomUrl := fmt.Sprintf("%s/chambers/%s/config", s.config.BackendURL, chamber.BackendID.Hex())
		roomReq, err := http.NewRequest("GET", roomUrl, nil)
		if err != nil {
			log.Printf("Failed to create request for room chamber %s: %v", chamber.Name, err)
			continue
		}

		if s.config.BackendAPIKey != "" {
			roomReq.Header.Set("Authorization", "Bearer "+s.config.BackendAPIKey)
		}

		roomResp, err := s.httpClient.Do(roomReq)
		if err != nil {
			log.Printf("Failed to fetch config for room chamber %s: %v", chamber.Name, err)
			continue
		}
		defer roomResp.Body.Close()

		if roomResp.StatusCode == http.StatusOK {
			var roomResponse struct {
				Success bool                 `json:"success"`
				Data    models.ChamberConfig `json:"data"`
				Error   string               `json:"error"`
			}

			if err := json.NewDecoder(roomResp.Body).Decode(&roomResponse); err == nil && roomResponse.Success {
				if err := s.chamberManager.UpdateChamberConfig(ctx, chamber.ID, &roomResponse.Data); err != nil {
					log.Printf("Failed to update room chamber %s config: %v", chamber.Name, err)
				} else {
					log.Printf("📊 Room chamber %s configuration synced from backend", chamber.Name)
				}
			}
		}
	}

	return nil
}

// syncExperiments fetches experiments from backend and updates local database
func (s *SyncService) syncExperiments() error {
	if s.backendServerID.IsZero() {
		return fmt.Errorf("no backend ID set - chamber not registered")
	}

	// Fetch experiments from backend
	url := fmt.Sprintf("%s/experiments?chamber_id=%s", s.config.BackendURL, s.backendServerID.Hex())
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
		// Store backend ID
		backendID := experiment.ID
		experiment.BackendID = backendID
		experiment.ID = primitive.ObjectID{} // Clear ID for local storage
		experiment.SyncedAt = now

		// Check if experiment already exists
		var existingExperiment models.Experiment
		err := s.db.ExperimentsCollection.FindOne(
			ctx,
			bson.M{"backend_id": backendID},
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
			log.Printf("Inserting new experiment: %s", experiment.Title)
			_, err = s.db.ExperimentsCollection.InsertOne(ctx, experiment)
			if err != nil {
				log.Printf("Failed to insert experiment %s: %v", experiment.Title, err)
				continue
			}
		}

		syncedCount++

		// If experiment is active, ensure executor knows about it
		if experiment.Status == models.StatusActive {
			log.Printf("🔄 Active experiment detected: %s", experiment.Title)
			// The executor service will pick this up automatically
		}
	}

	log.Printf("📊 Synced %d experiments from backend (using %s time)",
		syncedCount,
		func() string {
			if s.ntpService.IsConnected() {
				return "NTP"
			}
			return "system"
		}())
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
	now := s.ntpService.NowInMoscow()

	return map[string]interface{}{
		"backend_connected":  !s.backendServerID.IsZero(),
		"backend_id":         s.backendServerID.Hex(),
		"active_experiments": len(activeExperiments),
		"last_sync_time":     now.Format("2006-01-02T15:04:05Z07:00"),
		"ntp_enabled":        s.ntpService.IsEnabled(),
		"ntp_connected":      s.ntpService.IsConnected(),
	}
}
