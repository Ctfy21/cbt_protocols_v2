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
	config          *config.Config
	db              *database.MongoDB
	ntpService      *ntp.TimeService
	httpClient      *http.Client
	backendServerID primitive.ObjectID
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
	Name         string               `json:"name"`
	RoomSuffix   string               `json:"room_suffix,omitempty"`
	Location     string               `json:"location"`
	ServerID     string               `json:"server_id"`
	AccessToken  string               `json:"access_token"`
	Config       models.ChamberConfig `json:"config"`
	NTPEnabled   bool                 `json:"ntp_enabled"`
	NTPConnected bool                 `json:"ntp_connected"`
	CurrentTime  string               `json:"current_time"`
}

// RegisterWithBackend registers the chamber with the backend
func (s *RegistrationService) RegisterWithBackend(server *models.Server) error {
	// Prepare registration request
	req := RegistrationRequest{
		Name:         server.Name,
		Location:     "Local API v2",
		ServerID:     server.ID.Hex(),
		AccessToken:  s.config.HomeAssistantToken,
		NTPEnabled:   s.ntpService.IsEnabled(),
		NTPConnected: s.ntpService.IsConnected(),
		CurrentTime:  s.ntpService.NowInMoscow().Format("2006-01-02T15:04:05"),
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal registration request: %v", err)
	}

	// Send registration request to backend
	url := fmt.Sprintf("%s/servers", s.config.BackendURL)
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

	// Update server with backend ID
	server.BackendServerID = backendID
	s.backendServerID = backendID

	// Update in database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := s.ntpService.NowInMoscow()
	_, err = s.db.ServersCollection.UpdateOne(
		ctx,
		bson.M{"_id": server.ID},
		bson.M{"$set": bson.M{
			"backend_server_id": backendID,
			"updated_at":        now,
		}},
	)
	if err != nil {
		return fmt.Errorf("failed to update server with backend ID: %v", err)
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
				if s.backendServerID.IsZero() {
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
	if !s.backendServerID.IsZero() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		now := s.ntpService.NowInMoscow()
		_, err := s.db.ServersCollection.UpdateOne(
			ctx,
			bson.M{"_id": s.backendServerID},
			bson.M{"$set": bson.M{
				"last_heartbeat": now,
			}},
		)
		if err != nil {
			log.Printf("Failed to update local heartbeat: %v", err)
		}
	}

	// Skip backend heartbeat if not registered yet
	if s.backendServerID.IsZero() {
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
	url := fmt.Sprintf("%s/servers/%s/heartbeat", s.config.BackendURL, s.backendServerID.Hex())
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

// SetBackendServerID sets the backend server ID for the service
func (s *RegistrationService) SetBackendServerID(id primitive.ObjectID) {
	s.backendServerID = id
	log.Printf("Registration service: Backend ID set to %s", id.Hex())
}

func (s *RegistrationService) RegisterChamberWithBackend(roomChamber *models.Chamber) error {
	// Prepare registration request for room chamber
	req := RegistrationRequest{
		Name:        roomChamber.Name,
		Location:    fmt.Sprintf("Room: %s", roomChamber.RoomSuffix),
		ServerID:    roomChamber.ServerID.Hex(),
		AccessToken: s.config.HomeAssistantToken,
		Config:      roomChamber.Config,
		CurrentTime: s.ntpService.NowInMoscow().Format("2006-01-02T15:04:05Z07:00"),
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
	_, err = s.db.Database.Collection("chambers").UpdateOne(
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
