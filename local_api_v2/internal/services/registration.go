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
	config     *config.Config
	db         *database.MongoDB
	ntpService *ntp.TimeService
	httpClient *http.Client
	chamberID  primitive.ObjectID
	backendID  primitive.ObjectID
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
	}
}

// Update the RegistrationRequest struct to support room chambers:
type RegistrationRequest struct {
	Name            string                `json:"name"`
	RoomSuffix      string                `json:"room_suffix,omitempty"`       // For room chambers
	ParentChamberID string                `json:"parent_chamber_id,omitempty"` // For room chambers
	Location        string                `json:"location"`
	HAUrl           string                `json:"ha_url"`
	AccessToken     string                `json:"access_token"`
	LocalIP         string                `json:"local_ip"`
	InputNumbers    []models.InputNumber  `json:"input_numbers"`
	Lamps           []models.Lamp         `json:"lamps"`
	WateringZones   []models.WateringZone `json:"watering_zones"`
	NTPEnabled      bool                  `json:"ntp_enabled"`
	NTPConnected    bool                  `json:"ntp_connected"`
	CurrentTime     string                `json:"current_time"`
}

// RegisterWithBackend registers the chamber with the backend
func (s *RegistrationService) RegisterWithBackend(chamber *models.Chamber) error {
	// Prepare registration request
	req := RegistrationRequest{
		Name:          chamber.Name,
		Location:      "Local API v2",
		HAUrl:         chamber.HomeAssistantURL,
		AccessToken:   s.config.HomeAssistantToken,
		LocalIP:       chamber.LocalIP,
		InputNumbers:  chamber.InputNumbers,
		Lamps:         chamber.Lamps,
		WateringZones: chamber.WateringZones,
		NTPEnabled:    s.ntpService.IsEnabled(),
		NTPConnected:  s.ntpService.IsConnected(),
		CurrentTime:   s.ntpService.NowInMoscow().Format("2006-01-02T15:04:05Z07:00"),
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
	s.backendID = backendID

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

	log.Printf("✅ Successfully registered with backend. Backend ID: %s", backendID.Hex())
	return nil
}

// StartHeartbeat starts the heartbeat service
func (s *RegistrationService) StartHeartbeat(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(s.config.HeartbeatInterval) * time.Second)
	defer ticker.Stop()

	// Send initial heartbeat
	if err := s.sendHeartbeat(); err != nil {
		log.Printf("❌ Failed to send initial heartbeat: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("Heartbeat service stopped")
			return
		case <-ticker.C:
			if err := s.sendHeartbeat(); err != nil {
				// Only log as warning if it's not the "no backend ID" error
				if s.backendID.IsZero() {
					log.Printf("⚠️  Heartbeat skipped: Chamber not registered with backend yet")
				} else {
					log.Printf("❌ Failed to send heartbeat: %v", err)
				}
			}
		}
	}
}

// sendHeartbeat sends a heartbeat to the backend
func (s *RegistrationService) sendHeartbeat() error {
	// Update local heartbeat timestamp first
	if !s.chamberID.IsZero() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		now := s.ntpService.NowInMoscow()
		_, err := s.db.ChambersCollection.UpdateOne(
			ctx,
			bson.M{"_id": s.chamberID},
			bson.M{"$set": bson.M{
				"last_heartbeat": now,
			}},
		)
		if err != nil {
			log.Printf("Failed to update local heartbeat: %v", err)
		}
	}

	// Skip backend heartbeat if not registered yet
	if s.backendID.IsZero() {
		return fmt.Errorf("no backend ID set - chamber not registered")
	}

	// Prepare heartbeat payload with NTP status
	heartbeatData := map[string]interface{}{
		"timestamp":     s.ntpService.NowInMoscow().Format("2006-01-02T15:04:05Z07:00"),
		"ntp_enabled":   s.ntpService.IsEnabled(),
		"ntp_connected": s.ntpService.IsConnected(),
		"ntp_offset":    s.ntpService.GetOffset().String(),
	}

	jsonData, err := json.Marshal(heartbeatData)
	if err != nil {
		log.Printf("Warning: Failed to marshal heartbeat data: %v", err)
		jsonData = []byte("{}")
	}

	// Send heartbeat to backend
	url := fmt.Sprintf("%s/chambers/%s/heartbeat", s.config.BackendURL, s.backendID.Hex())
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

	log.Println("💓 Heartbeat sent successfully")
	return nil
}

// SetChamberID sets the local chamber ID for the service
func (s *RegistrationService) SetChamberID(id primitive.ObjectID) {
	s.chamberID = id
	log.Printf("Registration service: Chamber ID set to %s", id.Hex())
}

// SetBackendID sets the backend chamber ID for the service
func (s *RegistrationService) SetBackendID(id primitive.ObjectID) {
	s.backendID = id
	log.Printf("Registration service: Backend ID set to %s", id.Hex())
}

func (s *RegistrationService) RegisterRoomChamberWithBackend(roomChamber *models.RoomChamber) error {
	// Prepare registration request for room chamber
	req := RegistrationRequest{
		Name:          roomChamber.Name,
		Location:      fmt.Sprintf("Room: %s", roomChamber.RoomSuffix),
		HAUrl:         roomChamber.HomeAssistantURL,
		AccessToken:   s.config.HomeAssistantToken,
		LocalIP:       roomChamber.LocalIP,
		InputNumbers:  roomChamber.InputNumbers,
		Lamps:         roomChamber.Lamps,
		WateringZones: roomChamber.WateringZones,
		CurrentTime:   s.ntpService.NowInMoscow().Format("2006-01-02T15:04:05Z07:00"),
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal registration request: %v", err)
	}

	// Send registration request to backend room-chambers endpoint
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

	// Parse response to get backend room chamber ID
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

	// Update room chamber with backend ID
	roomChamber.BackendID = backendID

	// Update in database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := s.ntpService.NowInMoscow()
	_, err = s.db.Database.Collection("room_chambers").UpdateOne(
		ctx,
		bson.M{"_id": roomChamber.ID},
		bson.M{"$set": bson.M{
			"backend_id": backendID,
			"updated_at": now,
		}},
	)
	if err != nil {
		return fmt.Errorf("failed to update room chamber with backend ID: %v", err)
	}

	log.Printf("✅ Successfully registered room chamber '%s' with backend. Backend ID: %s", roomChamber.Name, backendID.Hex())
	return nil
}

// SendRoomChamberHeartbeat sends a heartbeat for a specific room chamber
func (s *RegistrationService) SendRoomChamberHeartbeat(roomChamber *models.RoomChamber) error {
	// Skip backend heartbeat if not registered yet
	if roomChamber.BackendID.IsZero() {
		return fmt.Errorf("no backend ID set for room chamber %s - not registered", roomChamber.Name)
	}

	// Prepare heartbeat payload with NTP status
	heartbeatData := map[string]interface{}{
		"timestamp":     s.ntpService.NowInMoscow().Format("2006-01-02T15:04:05Z07:00"),
		"ntp_enabled":   s.ntpService.IsEnabled(),
		"ntp_connected": s.ntpService.IsConnected(),
		"ntp_offset":    s.ntpService.GetOffset().String(),
		"room_suffix":   roomChamber.RoomSuffix,
	}

	jsonData, err := json.Marshal(heartbeatData)
	if err != nil {
		log.Printf("Warning: Failed to marshal room chamber heartbeat data: %v", err)
		jsonData = []byte("{}")
	}

	// Send heartbeat to backend
	url := fmt.Sprintf("%s/room-chambers/%s/heartbeat", s.config.BackendURL, roomChamber.BackendID.Hex())
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

	log.Printf("💓 Room chamber '%s' heartbeat sent successfully", roomChamber.Name)
	return nil
}

// Enhanced heartbeat service to handle room chambers
func (s *RegistrationService) StartEnhancedHeartbeat(ctx context.Context, chamberManager *ChamberManager) {
	ticker := time.NewTicker(time.Duration(s.config.HeartbeatInterval) * time.Second)
	defer ticker.Stop()

	// Send initial heartbeat
	s.sendEnhancedHeartbeat(chamberManager)

	for {
		select {
		case <-ctx.Done():
			log.Println("Enhanced heartbeat service stopped")
			return
		case <-ticker.C:
			s.sendEnhancedHeartbeat(chamberManager)
		}
	}
}

func (s *RegistrationService) sendEnhancedHeartbeat(chamberManager *ChamberManager) {
	// Send heartbeat for parent chamber
	if err := s.sendHeartbeat(); err != nil {
		if s.backendID.IsZero() {
			log.Printf("⚠️  Parent chamber heartbeat skipped: Chamber not registered with backend yet")
		} else {
			log.Printf("❌ Failed to send parent chamber heartbeat: %v", err)
		}
	}

	// Send heartbeats for all room chambers
	roomChambers := chamberManager.GetRoomChambers()
	for roomSuffix, roomChamber := range roomChambers {
		if err := s.SendRoomChamberHeartbeat(roomChamber); err != nil {
			if roomChamber.BackendID.IsZero() {
				log.Printf("⚠️  Room chamber '%s' heartbeat skipped: Not registered with backend yet", roomSuffix)
			} else {
				log.Printf("❌ Failed to send room chamber '%s' heartbeat: %v", roomSuffix, err)
			}
		}
	}
}
