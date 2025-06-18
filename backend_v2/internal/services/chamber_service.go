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
			ID:                 primitive.NewObjectID(),
			Name:               req.Name,
			Suffix:             req.Suffix,
			Location:           req.Location,
			HAUrl:              req.HAUrl,
			AccessToken:        req.AccessToken,
			LocalIP:            req.LocalIP,
			Status:             models.StatusOnline,
			LastHeartbeat:      now,
			DiscoveryCompleted: true,
			CreatedAt:          now,
			UpdatedAt:          now,
		}

		// Initialize config
		chamber.InitializeConfig()
		chamber.Config.Lamps = req.Lamps
		chamber.Config.WateringZones = req.WateringZones
		chamber.Config.UnrecognisedEntities = req.UnrecognisedEntities
		chamber.Config.DayDuration = req.DayDuration
		chamber.Config.DayStart = req.DayStart
		chamber.Config.Temperature = req.Temperature
		chamber.Config.Humidity = req.Humidity
		chamber.Config.CO2 = req.CO2
		chamber.Config.UpdatedAt = now

		_, err = s.db.ChambersCollection.InsertOne(ctx, chamber)
		if err != nil {
			return nil, fmt.Errorf("failed to create chamber: %v", err)
		}

		log.Printf("New chamber registered: %s (%s)", chamber.Name, chamber.ID.Hex())
		s.logChamberEntities(&chamber)

		return &chamber, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to check existing chamber: %v", err)
	}

	// Update existing chamber
	if existingChamber.Config == nil {
		existingChamber.InitializeConfig()
	}

	// Update config with new entities
	existingChamber.Config.Lamps = req.Lamps
	existingChamber.Config.WateringZones = req.WateringZones
	existingChamber.Config.UnrecognisedEntities = req.UnrecognisedEntities
	existingChamber.Config.DayDuration = req.DayDuration
	existingChamber.Config.DayStart = req.DayStart
	existingChamber.Config.Temperature = req.Temperature
	existingChamber.Config.Humidity = req.Humidity
	existingChamber.Config.CO2 = req.CO2
	existingChamber.Config.UpdatedAt = now

	update := bson.M{
		"$set": bson.M{
			"suffix":              req.Suffix,
			"location":            req.Location,
			"ha_url":              req.HAUrl,
			"access_token":        req.AccessToken,
			"status":              models.StatusOnline,
			"last_heartbeat":      now,
			"discovery_completed": true,
			"config":              existingChamber.Config,
			"updated_at":          now,
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
	s.logChamberEntities(&existingChamber)

	return &existingChamber, nil
}

// Helper method to log chamber entities
func (s *ChamberService) logChamberEntities(chamber *models.Chamber) {
	if chamber.Config == nil {
		return
	}

	log.Printf("Chamber %s entities:", chamber.Name)
	log.Printf("  - %d lamps", len(chamber.Config.Lamps))
	log.Printf("  - %d watering zones", len(chamber.Config.WateringZones))
	log.Printf("  - %d unrecognised entities", len(chamber.Config.UnrecognisedEntities))

	// Log climate control mappings
	log.Printf("  Climate control mappings:")
	log.Printf("    - Day duration: %d entities", len(chamber.Config.DayDuration))
	log.Printf("    - Day start: %d entities", len(chamber.Config.DayStart))
	log.Printf("    - Temperature: %d day, %d night", len(chamber.Config.Temperature["day"]), len(chamber.Config.Temperature["night"]))
	log.Printf("    - Humidity: %d day, %d night", len(chamber.Config.Humidity["day"]), len(chamber.Config.Humidity["night"]))
	log.Printf("    - CO2: %d day, %d night", len(chamber.Config.CO2["day"]), len(chamber.Config.CO2["night"]))

	// Log watering zones details
	if len(chamber.Config.WateringZones) > 0 {
		log.Println("  Watering zones:")
		for _, zone := range chamber.Config.WateringZones {
			log.Printf("    - %s:", zone.Name)
			log.Printf("      Start Time: %s", zone.StartTimeEntityID)
			log.Printf("      Period: %s", zone.PeriodEntityID)
			log.Printf("      Pause: %s", zone.PauseBetweenEntityID)
			log.Printf("      Duration: %s", zone.DurationEntityID)
		}
	}

	// Log unrecognised entities
	if len(chamber.Config.UnrecognisedEntities) > 0 {
		log.Printf("  Unrecognised entities:")
		for _, entity := range chamber.Config.UnrecognisedEntities {
			log.Printf("    - %s (%s)", entity.EntityID, entity.FriendlyName)
		}
	}
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
	Name                 string                                   `json:"name" binding:"required"`
	Suffix               string                                   `json:"suffix"`
	Location             string                                   `json:"location"`
	HAUrl                string                                   `json:"ha_url" binding:"required"`
	AccessToken          string                                   `json:"access_token" binding:"required"`
	LocalIP              string                                   `json:"local_ip" binding:"required"`
	Lamps                map[string]models.InputNumber            `json:"lamps"`
	WateringZones        []models.WateringZone                    `json:"watering_zones"`
	UnrecognisedEntities map[string]models.InputNumber            `json:"unrecognised_entities"`
	DayDuration          map[string]models.InputNumber            `json:"day_duration"`
	DayStart             map[string]models.InputNumber            `json:"day_start"`
	Temperature          map[string]map[string]models.InputNumber `json:"temperature"`
	Humidity             map[string]map[string]models.InputNumber `json:"humidity"`
	CO2                  map[string]map[string]models.InputNumber `json:"co2"`
}

// UpdateChamberConfigRequest represents the request to update chamber configuration
type UpdateChamberConfigRequest struct {
	Lamps                map[string]models.InputNumber            `json:"lamps"`
	WateringZones        []models.WateringZone                    `json:"watering_zones"`
	UnrecognisedEntities map[string]models.InputNumber            `json:"unrecognised_entities"`
	DayDuration          map[string]models.InputNumber            `json:"day_duration"`
	DayStart             map[string]models.InputNumber            `json:"day_start"`
	Temperature          map[string]map[string]models.InputNumber `json:"temperature"`
	Humidity             map[string]map[string]models.InputNumber `json:"humidity"`
	CO2                  map[string]map[string]models.InputNumber `json:"co2"`
}

func (s *ChamberService) UpdateChamberConfig(chamberID string, req *UpdateChamberConfigRequest) (*models.ChamberConfig, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(chamberID)
	if err != nil {
		return nil, fmt.Errorf("invalid chamber ID: %v", err)
	}

	// Get the chamber to ensure it exists
	var chamber models.Chamber
	err = s.db.ChambersCollection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&chamber)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("chamber not found")
		}
		return nil, fmt.Errorf("failed to get chamber: %v", err)
	}

	// Initialize config if it doesn't exist
	chamber.InitializeConfig()

	// Update configuration fields
	now := time.Now()
	if req.Lamps != nil {
		chamber.Config.Lamps = req.Lamps
	}
	if req.WateringZones != nil {
		chamber.Config.WateringZones = req.WateringZones
	}
	if req.UnrecognisedEntities != nil {
		chamber.Config.UnrecognisedEntities = req.UnrecognisedEntities
	}
	if req.DayDuration != nil {
		chamber.Config.DayDuration = req.DayDuration
	}
	if req.DayStart != nil {
		chamber.Config.DayStart = req.DayStart
	}
	if req.Temperature != nil {
		chamber.Config.Temperature = req.Temperature
	}
	if req.Humidity != nil {
		chamber.Config.Humidity = req.Humidity
	}
	if req.CO2 != nil {
		chamber.Config.CO2 = req.CO2
	}
	chamber.Config.UpdatedAt = now

	// Update the chamber in database
	update := bson.M{
		"$set": bson.M{
			"config":     chamber.Config,
			"updated_at": now,
		},
	}

	_, err = s.db.ChambersCollection.UpdateByID(ctx, objectID, update)
	if err != nil {
		return nil, fmt.Errorf("failed to update chamber config: %v", err)
	}

	log.Printf("Chamber config updated: %s (%s)", chamber.Name, chamber.ID.Hex())

	return chamber.Config, nil
}

// GetChamberConfig retrieves the configuration for a chamber
func (s *ChamberService) GetChamberConfig(chamberID string) (*models.ChamberConfig, error) {
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

	// Initialize config if it doesn't exist
	if chamber.Config == nil {
		chamber.InitializeConfig()

		// Save the initialized config
		update := bson.M{
			"$set": bson.M{
				"config":     chamber.Config,
				"updated_at": time.Now(),
			},
		}
		_, err = s.db.ChambersCollection.UpdateByID(ctx, objectID, update)
		if err != nil {
			log.Printf("Failed to save initialized config: %v", err)
		}
	}

	return chamber.Config, nil
}

// GetChambersWithUpdatedConfigs retrieves all chambers that have updated configurations since the given timestamp
func (s *ChamberService) GetChambersWithUpdatedConfigs(since time.Time) ([]models.Chamber, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"config.updated_at": bson.M{"$gt": since},
	}

	opts := options.Find().SetSort(bson.D{primitive.E{Key: "config.updated_at", Value: -1}})
	cursor, err := s.db.ChambersCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get chambers with updated configs: %v", err)
	}
	defer cursor.Close(ctx)

	var chambers []models.Chamber
	if err = cursor.All(ctx, &chambers); err != nil {
		return nil, fmt.Errorf("failed to decode chambers: %v", err)
	}

	return chambers, nil
}

// MarkConfigSynced marks the configuration as synced for a chamber
func (s *ChamberService) MarkConfigSynced(chamberID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"config.synced_at": now,
		},
	}

	_, err := s.db.ChambersCollection.UpdateByID(ctx, chamberID, update)
	if err != nil {
		return fmt.Errorf("failed to mark config as synced: %v", err)
	}

	return nil
}
