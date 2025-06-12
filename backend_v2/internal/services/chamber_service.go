package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"backend_v2/internal/config"
	"backend_v2/internal/database"
	"backend_v2/internal/models"
)

// ChamberService handles chamber-related business logic
type ChamberService struct {
	db     *database.MongoDB
	config *config.Config
}

// NewChamberService creates a new chamber service
func NewChamberService(db *database.MongoDB, config *config.Config) *ChamberService {
	return &ChamberService{
		db:     db,
		config: config,
	}
}

// RegisterChamber registers a new chamber or updates existing one
func (s *ChamberService) RegisterChamber(req *RegisterChamberRequest) (*models.Chamber, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if chamber already exists by name and local IP
	var existingChamber models.Chamber
	err := s.db.ChambersCollection.FindOne(ctx, bson.M{
		"name":     req.Name,
		"local_ip": req.LocalIP,
	}).Decode(&existingChamber)

	now := time.Now()

	if err == mongo.ErrNoDocuments {
		// Create new chamber
		chamber := models.Chamber{
			ID:            primitive.NewObjectID(),
			Name:          req.Name,
			Location:      req.Location,
			HAUrl:         req.HAUrl,
			AccessToken:   req.AccessToken,
			LocalIP:       req.LocalIP,
			Status:        models.StatusOnline,
			LastHeartbeat: now,
			InputNumbers:  req.InputNumbers,
			Lamps:         req.Lamps,
			WateringZones: req.WateringZones,
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		_, err = s.db.ChambersCollection.InsertOne(ctx, chamber)
		if err != nil {
			return nil, fmt.Errorf("failed to create chamber: %v", err)
		}

		log.Printf("New chamber registered: %s (%s)", chamber.Name, chamber.ID.Hex())

		// Log discovered entities
		log.Printf("Chamber %s entities:", chamber.Name)
		log.Printf("  - %d input numbers", len(chamber.InputNumbers))
		log.Printf("  - %d lamps", len(chamber.Lamps))
		log.Printf("  - %d watering zones", len(chamber.WateringZones))

		// Log watering zones details
		if len(chamber.WateringZones) > 0 {
			log.Println("  Watering zones:")
			for _, zone := range chamber.WateringZones {
				log.Printf("    - %s:", zone.Name)
				log.Printf("      Start Time: %s", zone.StartTimeEntityID)
				log.Printf("      Period: %s", zone.PeriodEntityID)
				log.Printf("      Pause: %s", zone.PauseBetweenEntityID)
				log.Printf("      Duration: %s", zone.DurationEntityID)
			}
		}

		// Log watering-related input numbers
		wateringTypes := []string{
			models.InputNumberWateringStart,
			models.InputNumberWateringPeriod,
			models.InputNumberWateringPause,
			models.InputNumberWateringDuration,
		}

		log.Println("  Watering input numbers by type:")
		for _, wType := range wateringTypes {
			inputs := chamber.GetInputNumbersByType(wType)
			if len(inputs) > 0 {
				log.Printf("    - %s: %d entities", wType, len(inputs))
				for _, input := range inputs {
					log.Printf("      %s (current: %.1f)", input.EntityID, input.CurrentValue)
				}
			}
		}

		return &chamber, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to check existing chamber: %v", err)
	}

	// Update existing chamber
	update := bson.M{
		"$set": bson.M{
			"location":       req.Location,
			"ha_url":         req.HAUrl,
			"access_token":   req.AccessToken,
			"status":         models.StatusOnline,
			"last_heartbeat": now,
			"input_numbers":  req.InputNumbers,
			"lamps":          req.Lamps,
			"watering_zones": req.WateringZones,
			"updated_at":     now,
		},
	}

	_, err = s.db.ChambersCollection.UpdateByID(ctx, existingChamber.ID, update)
	if err != nil {
		return nil, fmt.Errorf("failed to update chamber: %v", err)
	}

	existingChamber.Status = models.StatusOnline
	existingChamber.LastHeartbeat = now
	existingChamber.UpdatedAt = now

	log.Printf("Chamber updated: %s (%s)", existingChamber.Name, existingChamber.ID.Hex())

	// Log discovered entities
	log.Printf("Chamber %s entities:", existingChamber.Name)
	log.Printf("  - %d input numbers", len(req.InputNumbers))
	log.Printf("  - %d lamps", len(req.Lamps))
	log.Printf("  - %d watering zones", len(req.WateringZones))

	// Log watering zones details
	if len(req.WateringZones) > 0 {
		log.Println("  Watering zones:")
		for _, zone := range req.WateringZones {
			log.Printf("    - %s:", zone.Name)
			log.Printf("      Start Time: %s", zone.StartTimeEntityID)
			log.Printf("      Period: %s", zone.PeriodEntityID)
			log.Printf("      Pause: %s", zone.PauseBetweenEntityID)
			log.Printf("      Duration: %s", zone.DurationEntityID)
		}
	}

	return &existingChamber, nil
}

// UpdateHeartbeat updates the chamber heartbeat
func (s *ChamberService) UpdateHeartbeat(chamberID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(chamberID)
	if err != nil {
		return fmt.Errorf("invalid chamber ID: %v", err)
	}

	update := bson.M{
		"$set": bson.M{
			"last_heartbeat": time.Now(),
			"status":         models.StatusOnline,
		},
	}

	result, err := s.db.ChambersCollection.UpdateByID(ctx, objectID, update)
	if err != nil {
		return fmt.Errorf("failed to update heartbeat: %v", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("chamber not found")
	}

	return nil
}

// GetChamber retrieves a chamber by ID
func (s *ChamberService) GetChamber(chamberID string) (*models.Chamber, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(chamberID)
	if err != nil {
		return nil, fmt.Errorf("invalid chamber ID: %v", err)
	}

	var chamber models.Chamber
	err = s.db.ChambersCollection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&chamber)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("chamber not found")
		}
		return nil, fmt.Errorf("failed to get chamber: %v", err)
	}

	return &chamber, nil
}

// GetChambers retrieves all chambers
func (s *ChamberService) GetChambers() ([]models.Chamber, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{primitive.E{Key: "created_at", Value: -1}})
	cursor, err := s.db.ChambersCollection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get chambers: %v", err)
	}
	defer cursor.Close(ctx)

	var chambers []models.Chamber
	if err = cursor.All(ctx, &chambers); err != nil {
		return nil, fmt.Errorf("failed to decode chambers: %v", err)
	}

	return chambers, nil
}

// UpdateChamberStatus updates the status of chambers based on heartbeat timeout
func (s *ChamberService) UpdateChamberStatus() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find chambers that haven't sent heartbeat within timeout period
	cutoffTime := time.Now().Add(-s.config.HeartbeatTimeout)

	filter := bson.M{
		"status":         models.StatusOnline,
		"last_heartbeat": bson.M{"$lt": cutoffTime},
	}

	update := bson.M{
		"$set": bson.M{
			"status": models.StatusOffline,
		},
	}

	result, err := s.db.ChambersCollection.UpdateMany(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update chamber status: %v", err)
	}

	if result.ModifiedCount > 0 {
		log.Printf("Marked %d chambers as offline", result.ModifiedCount)
	}

	return nil
}

// StartStatusMonitor starts the background service to monitor chamber status
func (s *ChamberService) StartStatusMonitor(ctx context.Context) {
	ticker := time.NewTicker(s.config.CleanupInterval)
	defer ticker.Stop()

	// Initial cleanup
	if err := s.UpdateChamberStatus(); err != nil {
		log.Printf("Error updating chamber status: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("Chamber status monitor stopped")
			return
		case <-ticker.C:
			if err := s.UpdateChamberStatus(); err != nil {
				log.Printf("Error updating chamber status: %v", err)
			}
		}
	}
}

// RegisterChamberRequest represents the request to register a chamber
type RegisterChamberRequest struct {
	Name          string                `json:"name" binding:"required"`
	Location      string                `json:"location"`
	HAUrl         string                `json:"ha_url" binding:"required"`
	AccessToken   string                `json:"access_token" binding:"required"`
	LocalIP       string                `json:"local_ip" binding:"required"`
	InputNumbers  []models.InputNumber  `json:"input_numbers"`
	Lamps         []models.Lamp         `json:"lamps"`
	WateringZones []models.WateringZone `json:"watering_zones"`
}
