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

// ChamberManager управляет множественными chamber в зависимости от комнат
type ChamberManager struct {
	config     *config.Config
	db         *database.MongoDB
	discovery  *DiscoveryService
	ntpService *ntp.TimeService
	server     *models.Server
	chambers   map[string]*models.Chamber
}

// NewChamberManager создает новый менеджер chamber
func NewChamberManager(cfg *config.Config, db *database.MongoDB, discovery *DiscoveryService, ntpService *ntp.TimeService) *ChamberManager {
	// Устанавливаем chamber суффиксы в discovery service
	discovery.SetChamberSuffixes(cfg.ChamberSuffixes)

	return &ChamberManager{
		config:     cfg,
		db:         db,
		discovery:  discovery,
		ntpService: ntpService,
		chambers:   make(map[string]*models.Chamber),
	}
}

// InitializeChambers инициализирует parent chamber и создает room chambers
func (cm *ChamberManager) InitializeChambers(ctx context.Context) error {
	// Сначала получаем или создаем родительскую chamber
	server, err := cm.getOrCreateServer(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize parent chamber: %w", err)
	}
	cm.server = server

	log.Printf("Server initialized: %s", server.Name)

	// Обнаруживаем entities, сгруппированные по комнатам
	chamberEntities, err := cm.discovery.DiscoverChamberEntities()
	if err != nil {
		return fmt.Errorf("failed to discover chamber entities: %w", err)
	}

	// Создаем или обновляем room chambers
	for roomSuffix, entities := range chamberEntities {
		chamber, err := cm.createOrUpdateChamber(ctx, roomSuffix, entities)
		if err != nil {
			log.Printf("Warning: Failed to create/update room chamber for %s: %v", roomSuffix, err)
			continue
		}
		cm.chambers[roomSuffix] = chamber
		log.Printf("Chamber created/updated: %s (%s)", chamber.Name, roomSuffix)
	}

	return nil
}

// getOrCreateServer получает существующую или создает новую родительскую chamber
func (cm *ChamberManager) getOrCreateServer(ctx context.Context) (*models.Server, error) {
	// Пытаемся найти существующую chamber
	var server models.Server
	err := cm.db.ServersCollection.FindOne(ctx, bson.M{}).Decode(&server)

	if err == nil {
		// Chamber существует, возвращаем её
		log.Printf("Found existing server: %s", server.Name)
		return &server, nil
	}

	if err != mongo.ErrNoDocuments {
		return nil, fmt.Errorf("failed to query chambers: %w", err)
	}

	// Создаем новую chamber
	log.Println("Creating new server...")
	now := cm.ntpService.NowInMoscow()
	server = models.Server{
		ID:               primitive.NewObjectID(),
		Name:             cm.config.ChamberName,
		LocalIP:          cm.config.LocalIP,
		HomeAssistantURL: cm.config.HomeAssistantURL,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	// Вставляем chamber в базу
	_, err = cm.db.ServersCollection.InsertOne(ctx, server)
	if err != nil {
		return nil, fmt.Errorf("failed to create parent chamber: %w", err)
	}

	log.Printf("Created new parent chamber: %s", server.Name)
	return &server, nil
}

// createOrUpdateRoomChamber создает или обновляет room chamber
func (cm *ChamberManager) createOrUpdateChamber(ctx context.Context, roomSuffix string, entities *ChamberEntities) (*models.Chamber, error) {
	// Определяем имя для room chamber
	roomChamberName := cm.config.ChamberName
	if roomSuffix != "default" {
		// Для известных суффиксов создаем более описательные имена
		switch roomSuffix {
		case "galo":
			roomChamberName = fmt.Sprintf("%s_Galo", cm.config.ChamberName)
		case "sb4":
			roomChamberName = fmt.Sprintf("%s_SB4", cm.config.ChamberName)
		case "oreol":
			roomChamberName = fmt.Sprintf("%s_Oreol", cm.config.ChamberName)
		case "sb1":
			roomChamberName = fmt.Sprintf("%s_SB1", cm.config.ChamberName)
		case "midi":
			roomChamberName = fmt.Sprintf("%s_MIDI", cm.config.ChamberName)
		default:
			roomChamberName = fmt.Sprintf("%s_%s", cm.config.ChamberName, strings.ToUpper(roomSuffix))
		}
	}

	// Пытаемся найти существующую chamber
	var chamber models.Chamber
	err := cm.db.Database.Collection("chambers").FindOne(ctx, bson.M{
		"server_id":   cm.server.ID,
		"room_suffix": roomSuffix,
	}).Decode(&chamber)

	now := cm.ntpService.NowInMoscow()

	if err == mongo.ErrNoDocuments {
		// Создаем новую room chamber
		chamber = models.Chamber{
			ID:         primitive.NewObjectID(),
			Name:       roomChamberName,
			RoomSuffix: roomSuffix,
			ServerID:   cm.server.ID,
			Config: models.ChamberConfig{
				InputNumbers:  entities.Config.InputNumbers,
				Lamps:         entities.Config.Lamps,
				WateringZones: entities.Config.WateringZones,
			},
			CreatedAt: now,
			UpdatedAt: now,
		}

		_, err = cm.db.Database.Collection("chambers").InsertOne(ctx, chamber)
		if err != nil {
			return nil, fmt.Errorf("failed to create room chamber: %w", err)
		}

		log.Printf("New chamber created: %s (%s)", chamber.Name, roomSuffix)
	} else if err != nil {
		return nil, fmt.Errorf("failed to query chamber: %w", err)
	} else {
		// Обновляем существующую chamber
		update := bson.M{
			"$set": bson.M{
				"name":           roomChamberName,
				"last_heartbeat": now,
				"config": bson.M{
					"input_numbers":  entities.Config.InputNumbers,
					"lamps":          entities.Config.Lamps,
					"watering_zones": entities.Config.WateringZones,
				},
				"updated_at": now,
			},
		}

		_, err = cm.db.Database.Collection("chambers").UpdateByID(ctx, chamber.ID, update)
		if err != nil {
			return nil, fmt.Errorf("failed to update chamber: %w", err)
		}

		// Обновляем локальные данные
		chamber.Name = roomChamberName
		chamber.Config = entities.Config
		chamber.UpdatedAt = now

		log.Printf("Chamber updated: %s (%s)", chamber.Name, roomSuffix)
	}

	// Логируем обнаруженные entities
	log.Printf("Chamber %s entities:", chamber.Name)
	log.Printf("  - %d input numbers", len(chamber.Config.InputNumbers))
	log.Printf("  - %d lamps", len(chamber.Config.Lamps))
	log.Printf("  - %d watering zones", len(chamber.Config.WateringZones))

	return &chamber, nil
}

// GetServer возвращает родительскую chamber
func (cm *ChamberManager) GetServer() *models.Server {
	return cm.server
}

// GetRoomChambers возвращает все room chambers
func (cm *ChamberManager) GetChambers() map[string]*models.Chamber {
	return cm.chambers
}

// GetRoomChamber возвращает room chamber по суффиксу
func (cm *ChamberManager) GetChamber(roomSuffix string) *models.Chamber {
	return cm.chambers[roomSuffix]
}

// UpdateHeartbeat обновляет heartbeat для всех chambers
func (cm *ChamberManager) UpdateHeartbeat(ctx context.Context) error {
	now := cm.ntpService.NowInMoscow()

	// Обновляем parent chamber
	if cm.server != nil {
		update := bson.M{
			"$set": bson.M{
				"last_heartbeat": now,
				"updated_at":     now,
			},
		}
		_, err := cm.db.ServersCollection.UpdateByID(ctx, cm.server.ID, update)
		if err != nil {
			log.Printf("Failed to update parent chamber heartbeat: %v", err)
		} else {
			cm.server.LastHeartbeat = now
			cm.server.UpdatedAt = now
		}
	}

	// Обновляем room chambers
	for roomSuffix, chamber := range cm.chambers {
		update := bson.M{
			"$set": bson.M{
				"last_heartbeat": now,
				"updated_at":     now,
			},
		}
		_, err := cm.db.Database.Collection("chambers").UpdateByID(ctx, chamber.ID, update)
		if err != nil {
			log.Printf("Failed to update room chamber heartbeat for %s: %v", roomSuffix, err)
		} else {
			chamber.UpdatedAt = now
		}
	}

	return nil
}

// RegisterRoomChambersWithBackend регистрирует все room chambers с бэкендом
func (cm *ChamberManager) RegisterChambersWithBackend(registrationService *RegistrationService) error {
	for roomSuffix, chamber := range cm.chambers {
		log.Printf("Registering chamber '%s' with backend...", roomSuffix)
		if err := registrationService.RegisterChamberWithBackend(chamber); err != nil {
			log.Printf("Warning: Failed to register chamber '%s': %v", roomSuffix, err)
			continue
		}
	}
	return nil
}
