package services

import (
	"bytes"
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

// RegistrationService handles chamber registration with backend
type RegistrationService struct {
	config       *config.Config
	db           *database.MongoDB
	ntpService   *ntp.TimeService
	httpClient   *http.Client
	chamberIDMap map[primitive.ObjectID]primitive.ObjectID // local ID -> backend ID
}

// NewRegistrationService creates a new registration service
func NewRegistrationService(cfg *config.Config, db *database.MongoDB, ntpService *ntp.TimeService) *RegistrationService {
	return &RegistrationService{
		config:     cfg,
		db:         db,
		ntpService: ntpService,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		chamberIDMap: make(map[primitive.ObjectID]primitive.ObjectID),
	}
}

// Updated RegistrationRequest in registration.go
type RegistrationRequest struct {
	Name                 string                        `json:"name"`
	Suffix               string                        `json:"suffix"`
	Location             string                        `json:"location"`
	HAUrl                string                        `json:"ha_url"`
	AccessToken          string                        `json:"access_token"`
	LocalIP              string                        `json:"local_ip"`
	Lamps                []models.Lamp                 `json:"lamps"`
	WateringZones        []models.WateringZone         `json:"watering_zones"`
	UnrecognisedEntities []models.InputNumber          `json:"unrecognised_entities"`
	DayDuration          map[string]float64            `json:"day_duration"`
	DayStart             map[string]float64            `json:"day_start"`
	Temperature          map[string]map[string]float64 `json:"temperature"`
	Humidity             map[string]map[string]float64 `json:"humidity"`
	CO2                  map[string]map[string]float64 `json:"co2"`
	LightIntensity       map[string]float64            `json:"light_intensity"`
	WateringSettings     map[string]map[string]float64 `json:"watering_settings"`
}

// Updated RegisterChamberWithBackend method
func (s *RegistrationService) RegisterChamberWithBackend(chamber *models.Chamber) error {
	// Skip if already registered
	if !chamber.BackendID.IsZero() {
		s.chamberIDMap[chamber.ID] = chamber.BackendID
		log.Printf("Chamber %s already registered with backend ID: %s", chamber.Name, chamber.BackendID.Hex())
		return nil
	}

	// Prepare registration request
	req := RegistrationRequest{
		Name:                 chamber.Name,
		Suffix:               chamber.Suffix,
		Location:             fmt.Sprintf("Local API v2 - %s", chamber.Suffix),
		HAUrl:                chamber.HomeAssistantURL,
		AccessToken:          s.config.HomeAssistantToken,
		LocalIP:              chamber.LocalIP,
		Lamps:                chamber.Config.Lamps,
		WateringZones:        chamber.Config.WateringZones,
		UnrecognisedEntities: chamber.Config.UnrecognisedEntities,
		DayDuration:          chamber.Config.DayDuration,
		DayStart:             chamber.Config.DayStart,
		Temperature:          chamber.Config.Temperature,
		Humidity:             chamber.Config.Humidity,
		CO2:                  chamber.Config.CO2,
		LightIntensity:       chamber.Config.LightIntensity,
		WateringSettings:     chamber.Config.WateringSettings,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal registration request: %v", err)
	}

	// Send registration request to backend
	url := fmt.Sprintf("%s/chambers", s.config.BackendURL)
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if s.config.BackendAPIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+s.config.BackendAPIKey)
	}

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send registration request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("backend returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response to get backend chamber ID
	var response struct {
		Success bool `json:"success"`
		Data    struct {
			ID string `json:"id"`
		} `json:"data"`
		Error string `json:"error"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	if !response.Success {
		return fmt.Errorf("registration failed: %s", response.Error)
	}

	// Convert backend ID to ObjectID
	backendID, err := primitive.ObjectIDFromHex(response.Data.ID)
	if err != nil {
		return fmt.Errorf("invalid backend ID: %v", err)
	}

	// Update chamber with backend ID
	chamber.BackendID = backendID
	s.chamberIDMap[chamber.ID] = backendID

	// Update in database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := s.ntpService.NowInMoscow()
	_, err = s.db.ChambersCollection.UpdateOne(
		ctx,
		bson.M{"_id": chamber.ID},
		bson.M{"$set": bson.M{
			"backend_id": backendID,
			"updated_at": now,
		}},
	)
	if err != nil {
		return fmt.Errorf("failed to update chamber with backend ID: %v", err)
	}

	log.Printf("✅ Successfully registered chamber %s with backend. Backend ID: %s", chamber.Name, backendID.Hex())
	log.Printf("  - Sent %d lamps", len(req.Lamps))
	log.Printf("  - Sent %d watering zones", len(req.WateringZones))
	log.Printf("  - Sent %d unrecognised entities", len(req.UnrecognisedEntities))
	log.Printf("  - Climate mappings: %d day_start, %d day_duration",
		len(req.DayStart), len(req.DayDuration))
	log.Printf("  - Temperature: %d day, %d night",
		len(req.Temperature["day"]), len(req.Temperature["night"]))

	return nil
}

// StartHeartbeat starts the heartbeat service for all registered chambers
func (s *RegistrationService) StartHeartbeat(ctx context.Context, chamberManager *ChamberManager) {
	ticker := time.NewTicker(time.Duration(s.config.HeartbeatInterval) * time.Second)
	defer ticker.Stop()

	// Send initial heartbeat
	s.sendHeartbeats(chamberManager)

	for {
		select {
		case <-ctx.Done():
			log.Println("Heartbeat service stopped")
			return
		case <-ticker.C:
			s.sendHeartbeats(chamberManager)
		}
	}
}

// sendHeartbeats sends heartbeats for all registered chambers
func (s *RegistrationService) sendHeartbeats(chamberManager *ChamberManager) {
	// Update local heartbeats first
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := chamberManager.UpdateHeartbeat(ctx); err != nil {
		log.Printf("Failed to update local heartbeats: %v", err)
	}

	// Send heartbeats to backend for registered chambers
	for localID, backendID := range s.chamberIDMap {
		chamber := chamberManager.GetChamberByID(localID)
		if chamber == nil {
			continue
		}

		if err := s.sendHeartbeat(backendID); err != nil {
			log.Printf("Failed to send heartbeat for chamber %s: %v", chamber.Name, err)
		}
	}
}

// sendHeartbeat sends a heartbeat for a specific chamber
func (s *RegistrationService) sendHeartbeat(backendID primitive.ObjectID) error {
	// Prepare heartbeat payload with NTP status
	heartbeatData := map[string]interface{}{
		"timestamp":     s.ntpService.NowInMoscow().Format("2006-01-02T15:04:05Z07:00"),
		"ntp_enabled":   s.ntpService.IsEnabled(),
		"ntp_connected": s.ntpService.IsConnected(),
		"ntp_offset":    s.ntpService.GetOffset().String(),
	}

	jsonData, err := json.Marshal(heartbeatData)
	if err != nil {
		return fmt.Errorf("failed to marshal heartbeat data: %v", err)
	}

	// Send heartbeat to backend
	url := fmt.Sprintf("%s/chambers/%s/heartbeat", s.config.BackendURL, backendID.Hex())
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create heartbeat request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if s.config.BackendAPIKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.config.BackendAPIKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send heartbeat: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("heartbeat failed with status %d: %s", resp.StatusCode, string(body))
	}

	log.Printf("💓 Heartbeat sent for chamber %s", backendID.Hex())
	return nil
}

// GetBackendID returns the backend ID for a local chamber ID
func (s *RegistrationService) GetBackendID(localID primitive.ObjectID) (primitive.ObjectID, bool) {
	backendID, exists := s.chamberIDMap[localID]
	return backendID, exists
}
