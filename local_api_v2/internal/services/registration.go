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
)

// RegistrationService handles chamber registration with backend
type RegistrationService struct {
	config     *config.Config
	db         *database.MongoDB
	httpClient *http.Client
	chamberID  primitive.ObjectID
	backendID  primitive.ObjectID
}

// NewRegistrationService creates a new registration service
func NewRegistrationService(cfg *config.Config, db *database.MongoDB) *RegistrationService {
	return &RegistrationService{
		config: cfg,
		db:     db,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// RegistrationRequest represents the data sent to backend for registration
type RegistrationRequest struct {
	Name          string                `json:"name"`
	Location      string                `json:"location"`
	HAUrl         string                `json:"ha_url"`
	AccessToken   string                `json:"access_token"`
	LocalIP       string                `json:"local_ip"`
	InputNumbers  []models.InputNumber  `json:"input_numbers"`
	Lamps         []models.Lamp         `json:"lamps"`
	WateringZones []models.WateringZone `json:"watering_zones"`
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

	_, err = s.db.ChambersCollection.UpdateOne(
		ctx,
		bson.M{"_id": chamber.ID},
		bson.M{"$set": bson.M{
			"backend_id": backendID,
			"updated_at": time.Now(),
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

		_, err := s.db.ChambersCollection.UpdateOne(
			ctx,
			bson.M{"_id": s.chamberID},
			bson.M{"$set": bson.M{
				"last_heartbeat": time.Now(),
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

	// Send heartbeat to backend
	url := fmt.Sprintf("%s/chambers/%s/heartbeat", s.config.BackendURL, s.backendID.Hex())
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create heartbeat request: %v", err)
	}

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

// RegisterRoomChamberWithBackend registers a room chamber with the backend
func (s *RegistrationService) RegisterRoomChamberWithBackend(roomChamber *models.RoomChamber) error {
	// Prepare registration request
	req := RegistrationRequest{
		Name:          roomChamber.Name,
		Location:      fmt.Sprintf("Room: %s", roomChamber.RoomSuffix),
		HAUrl:         roomChamber.HomeAssistantURL,
		AccessToken:   s.config.HomeAssistantToken,
		LocalIP:       roomChamber.LocalIP,
		InputNumbers:  roomChamber.InputNumbers,
		Lamps:         roomChamber.Lamps,
		WateringZones: roomChamber.WateringZones,
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

	// Update room chamber with backend ID
	roomChamber.BackendID = backendID

	// Update in database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = s.db.Database.Collection("room_chambers").UpdateOne(
		ctx,
		bson.M{"_id": roomChamber.ID},
		bson.M{"$set": bson.M{
			"backend_id": backendID,
			"updated_at": time.Now(),
		}},
	)
	if err != nil {
		return fmt.Errorf("failed to update room chamber with backend ID: %v", err)
	}

	log.Printf("✅ Successfully registered room chamber '%s' with backend. Backend ID: %s", roomChamber.Name, backendID.Hex())
	return nil
}
