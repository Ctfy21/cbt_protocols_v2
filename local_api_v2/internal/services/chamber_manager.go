package services

import (
	"context"
	"fmt"
	"log"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"local_api_v2/internal/config"
	"local_api_v2/internal/database"
	"local_api_v2/internal/models"
	"local_api_v2/pkg/ntp"
)

// ChamberManager manages multiple chambers based on room suffixes
type ChamberManager struct {
	config     *config.Config
	db         *database.MongoDB
	discovery  *DiscoveryService
	ntpService *ntp.TimeService
	chambers   map[string]*models.Chamber // key is suffix
}

// NewChamberManager creates a new chamber manager
func NewChamberManager(cfg *config.Config, db *database.MongoDB, discovery *DiscoveryService, ntpService *ntp.TimeService) *ChamberManager {
	// Set chamber suffixes in discovery service
	discovery.SetChamberSuffixes(cfg.ChamberSuffixes)

	return &ChamberManager{
		config:     cfg,
		db:         db,
		discovery:  discovery,
		ntpService: ntpService,
		chambers:   make(map[string]*models.Chamber),
	}
}

// InitializeChambers discovers and initializes all chambers
func (cm *ChamberManager) InitializeChambers(ctx context.Context) error {
	log.Printf("Initializing chambers with suffixes: %v", cm.config.ChamberSuffixes)

	// Discover entities grouped by rooms
	chamberEntities, err := cm.discovery.DiscoverChamberEntities()
	if err != nil {
		return fmt.Errorf("failed to discover chamber entities: %w", err)
	}

	// Create or update chambers
	for suffix, entities := range chamberEntities {
		chamber, err := cm.createOrUpdateChamber(ctx, suffix, entities)
		if err != nil {
			log.Printf("Warning: Failed to create/update chamber for %s: %v", suffix, err)
			continue
		}
		cm.chambers[suffix] = chamber
		log.Printf("Chamber initialized: %s (suffix: %s)", chamber.Name, suffix)
	}

	log.Printf("Initialized %d chambers", len(cm.chambers))
	return nil
}

// Updated createOrUpdateChamber method in chamber_manager.go
func (cm *ChamberManager) createOrUpdateChamber(ctx context.Context, suffix string, entities *ChamberEntities) (*models.Chamber, error) {
	// Generate chamber name
	chamberName := cm.generateChamberName(suffix)

	// Try to find existing chamber
	var chamber models.Chamber
	err := cm.db.ChambersCollection.FindOne(ctx, bson.M{
		"suffix": suffix,
	}).Decode(&chamber)

	now := cm.ntpService.Now()

	if err == mongo.ErrNoDocuments {
		// Create new chamber
		chamber = models.Chamber{
			ID:               primitive.NewObjectID(),
			Name:             chamberName,
			Suffix:           suffix,
			LocalIP:          cm.config.LocalIP,
			LocalAPIversion:  cm.config.LocalAPIversion,
			TimeOffset:       cm.ntpService.TimeOffset(),
			HomeAssistantURL: cm.config.HomeAssistantURL,
			Status:           "online",
			LastHeartbeat:    now,
			Config: models.ChamberConfig{
				Lamps:                entities.Config.Lamps,
				WateringZones:        entities.Config.WateringZones,
				UnrecognisedEntities: entities.Config.UnrecognisedEntities,
				DayDuration:          entities.Config.DayDuration,
				DayStart:             entities.Config.DayStart,
				Temperature:          entities.Config.Temperature,
				Humidity:             entities.Config.Humidity,
				CO2:                  entities.Config.CO2,
				UpdatedAt:            now,
			},
			DiscoveryCompleted: true,
			CreatedAt:          now,
			UpdatedAt:          now,
		}

		_, err = cm.db.ChambersCollection.InsertOne(ctx, chamber)
		if err != nil {
			return nil, fmt.Errorf("failed to create chamber: %w", err)
		}

		log.Printf("Created new chamber: %s", chamber.Name)
		cm.logChamberSummary(&chamber)
	} else if err != nil {
		return nil, fmt.Errorf("failed to query chamber: %w", err)
	} else {
		// Update existing chamber
		update := bson.M{
			"$set": bson.M{
				"name":                chamberName,
				"local_ip":            cm.config.LocalIP,
				"ha_url":              cm.config.HomeAssistantURL,
				"status":              "online",
				"last_heartbeat":      now,
				"discovery_completed": true,
				"config": bson.M{
					"lamps":                 entities.Config.Lamps,
					"watering_zones":        entities.Config.WateringZones,
					"unrecognised_entities": entities.Config.UnrecognisedEntities,
					"day_duration":          entities.Config.DayDuration,
					"day_start":             entities.Config.DayStart,
					"temperature":           entities.Config.Temperature,
					"humidity":              entities.Config.Humidity,
					"co2":                   entities.Config.CO2,
					"updated_at":            now,
				},
				"updated_at": now,
			},
		}

		_, err = cm.db.ChambersCollection.UpdateByID(ctx, chamber.ID, update)
		if err != nil {
			return nil, fmt.Errorf("failed to update chamber: %w", err)
		}

		// Update local data
		chamber.Name = chamberName
		chamber.LocalIP = cm.config.LocalIP
		chamber.HomeAssistantURL = cm.config.HomeAssistantURL
		chamber.Status = "online"
		chamber.LastHeartbeat = now
		chamber.Config = models.ChamberConfig{
			Lamps:                entities.Config.Lamps,
			WateringZones:        entities.Config.WateringZones,
			UnrecognisedEntities: entities.Config.UnrecognisedEntities,
			DayDuration:          entities.Config.DayDuration,
			DayStart:             entities.Config.DayStart,
			Temperature:          entities.Config.Temperature,
			Humidity:             entities.Config.Humidity,
			CO2:                  entities.Config.CO2,
			UpdatedAt:            now,
		}
		chamber.DiscoveryCompleted = true
		chamber.UpdatedAt = now

		log.Printf("Updated chamber: %s", chamber.Name)
		cm.logChamberSummary(&chamber)
	}

	return &chamber, nil
}

// Helper method to log chamber summary
func (cm *ChamberManager) logChamberSummary(chamber *models.Chamber) {
	log.Printf("Chamber %s (suffix: %s) summary:", chamber.Name, chamber.Suffix)
	log.Printf("  - %d lamps", len(chamber.Config.Lamps))
	log.Printf("  - %d watering zones", len(chamber.Config.WateringZones))
	log.Printf("  - %d unrecognised entities", len(chamber.Config.UnrecognisedEntities))
	log.Printf("  - Climate mappings: %d day_start, %d day_duration",
		len(chamber.Config.DayStart), len(chamber.Config.DayDuration))
	log.Printf("  - Temperature: %d day, %d night",
		len(chamber.Config.Temperature["day"]), len(chamber.Config.Temperature["night"]))
	log.Printf("  - Humidity: %d day, %d night",
		len(chamber.Config.Humidity["day"]), len(chamber.Config.Humidity["night"]))
	log.Printf("  - CO2: %d day, %d night",
		len(chamber.Config.CO2["day"]), len(chamber.Config.CO2["night"]))

}

// generateChamberName generates a descriptive name for the chamber
func (cm *ChamberManager) generateChamberName(suffix string) string {
	baseName := cm.config.ChamberName

	if suffix == "default" {
		return baseName
	}

	return fmt.Sprintf("%s_%s", baseName, strings.ToUpper(suffix))

}

// GetChambers returns all chambers
func (cm *ChamberManager) GetChambers() map[string]*models.Chamber {
	return cm.chambers
}

// GetChamber returns a chamber by suffix
func (cm *ChamberManager) GetChamber(suffix string) *models.Chamber {
	return cm.chambers[suffix]
}

// GetChamberByID returns a chamber by its MongoDB ID
func (cm *ChamberManager) GetChamberByID(id primitive.ObjectID) *models.Chamber {
	for _, chamber := range cm.chambers {
		if chamber.ID == id {
			return chamber
		}
	}
	return nil
}

// UpdateChamberConfig updates chamber configuration
func (cm *ChamberManager) UpdateChamberConfig(ctx context.Context, chamberID primitive.ObjectID, config *models.ChamberConfig) error {
	now := cm.ntpService.Now()
	config.UpdatedAt = now
	config.SyncedAt = &now

	update := bson.M{
		"$set": bson.M{
			"config":     config,
			"updated_at": now,
		},
	}

	_, err := cm.db.ChambersCollection.UpdateByID(ctx, chamberID, update)
	if err != nil {
		return fmt.Errorf("failed to update chamber config: %w", err)
	}

	// Update local copy
	for _, chamber := range cm.chambers {
		if chamber.ID == chamberID {
			chamber.Config = *config
			chamber.UpdatedAt = now
			log.Printf("Updated configuration for chamber: %s", chamber.Name)
			break
		}
	}

	return nil
}

// UpdateHeartbeat updates heartbeat for all chambers
func (cm *ChamberManager) UpdateHeartbeat(ctx context.Context) error {
	now := cm.ntpService.Now()

	for suffix, chamber := range cm.chambers {
		update := bson.M{
			"$set": bson.M{
				"last_heartbeat": now,
				"status":         "online",
				"updated_at":     now,
			},
		}

		_, err := cm.db.ChambersCollection.UpdateByID(ctx, chamber.ID, update)
		if err != nil {
			log.Printf("Failed to update heartbeat for chamber %s: %v", suffix, err)
		} else {
			chamber.LastHeartbeat = now
			chamber.Status = "online"
			chamber.UpdatedAt = now
		}
	}

	return nil
}

// RegisterChambersWithBackend registers all chambers with the backend
func (cm *ChamberManager) RegisterChambersWithBackend(registrationService *RegistrationService) error {
	successCount := 0

	for suffix, chamber := range cm.chambers {
		log.Printf("Registering chamber '%s' with backend...", suffix)
		if err := registrationService.RegisterChamberWithBackend(chamber); err != nil {
			log.Printf("Warning: Failed to register chamber '%s': %v", suffix, err)
			continue
		}
		successCount++
	}

	if successCount == 0 {
		return fmt.Errorf("failed to register any chambers")
	}

	log.Printf("Successfully registered %d/%d chambers", successCount, len(cm.chambers))
	return nil
}

// GetRegisteredChambers returns only chambers that are registered with backend
func (cm *ChamberManager) GetRegisteredChambers() []*models.Chamber {
	var registered []*models.Chamber

	for _, chamber := range cm.chambers {
		if !chamber.BackendID.IsZero() {
			registered = append(registered, chamber)
		}
	}

	return registered
}
