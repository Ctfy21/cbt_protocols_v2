package services

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"local_api_v2/internal/config"
	"local_api_v2/internal/database"
	"local_api_v2/internal/models"
)

// ChamberManager управляет множественными chamber в зависимости от комнат
type ChamberManager struct {
	config        *config.Config
	db            *database.MongoDB
	discovery     *DiscoveryService
	parentChamber *models.Chamber
	roomChambers  map[string]*models.RoomChamber
}

// NewChamberManager создает новый менеджер chamber
func NewChamberManager(cfg *config.Config, db *database.MongoDB, discovery *DiscoveryService) *ChamberManager {
	// Устанавливаем chamber суффиксы в discovery service
	discovery.SetChamberSuffixes(cfg.ChamberSuffixes)

	return &ChamberManager{
		config:       cfg,
		db:           db,
		discovery:    discovery,
		roomChambers: make(map[string]*models.RoomChamber),
	}
}

// InitializeChambers инициализирует parent chamber и создает room chambers
func (cm *ChamberManager) InitializeChambers(ctx context.Context) error {
	// Сначала получаем или создаем родительскую chamber
	parentChamber, err := cm.getOrCreateParentChamber(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize parent chamber: %w", err)
	}
	cm.parentChamber = parentChamber

	log.Printf("Parent chamber initialized: %s", parentChamber.Name)

	// Обнаруживаем entities, сгруппированные по комнатам
	roomEntities, err := cm.discovery.DiscoverRoomEntities()
	if err != nil {
		return fmt.Errorf("failed to discover room entities: %w", err)
	}

	// Создаем или обновляем room chambers
	for roomSuffix, entities := range roomEntities {
		roomChamber, err := cm.createOrUpdateRoomChamber(ctx, roomSuffix, entities)
		if err != nil {
			log.Printf("Warning: Failed to create/update room chamber for %s: %v", roomSuffix, err)
			continue
		}
		cm.roomChambers[roomSuffix] = roomChamber
		log.Printf("Room chamber created/updated: %s (%s)", roomChamber.Name, roomSuffix)
	}

	return nil
}

// getOrCreateParentChamber получает существующую или создает новую родительскую chamber
func (cm *ChamberManager) getOrCreateParentChamber(ctx context.Context) (*models.Chamber, error) {
	// Пытаемся найти существующую chamber
	var chamber models.Chamber
	err := cm.db.ChambersCollection.FindOne(ctx, bson.M{}).Decode(&chamber)

	if err == nil {
		// Chamber существует, возвращаем её
		log.Printf("Found existing parent chamber: %s", chamber.Name)
		return &chamber, nil
	}

	if err != mongo.ErrNoDocuments {
		return nil, fmt.Errorf("failed to query chambers: %w", err)
	}

	// Создаем новую chamber
	log.Println("Creating new parent chamber...")
	chamber = models.Chamber{
		ID:               primitive.NewObjectID(),
		Name:             cm.config.ChamberName,
		LocalIP:          cm.config.LocalIP,
		HomeAssistantURL: cm.config.HomeAssistantURL,
		InputNumbers:     []models.InputNumber{},
		Lamps:            []models.Lamp{},
		WateringZones:    []models.WateringZone{},
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Вставляем chamber в базу
	_, err = cm.db.ChambersCollection.InsertOne(ctx, chamber)
	if err != nil {
		return nil, fmt.Errorf("failed to create parent chamber: %w", err)
	}

	log.Printf("Created new parent chamber: %s", chamber.Name)
	return &chamber, nil
}

// createOrUpdateRoomChamber создает или обновляет room chamber
func (cm *ChamberManager) createOrUpdateRoomChamber(ctx context.Context, roomSuffix string, entities *RoomEntities) (*models.RoomChamber, error) {
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
		default:
			roomChamberName = fmt.Sprintf("%s_%s", cm.config.ChamberName, strings.ToUpper(roomSuffix))
		}
	}

	// Пытаемся найти существующую room chamber
	var roomChamber models.RoomChamber
	err := cm.db.Database.Collection("room_chambers").FindOne(ctx, bson.M{
		"parent_chamber_id": cm.parentChamber.ID,
		"room_suffix":       roomSuffix,
	}).Decode(&roomChamber)

	now := time.Now()

	if err == mongo.ErrNoDocuments {
		// Создаем новую room chamber
		roomChamber = models.RoomChamber{
			ID:               primitive.NewObjectID(),
			Name:             roomChamberName,
			RoomSuffix:       roomSuffix,
			ParentChamberID:  cm.parentChamber.ID,
			LocalIP:          cm.config.LocalIP,
			HomeAssistantURL: cm.config.HomeAssistantURL,
			LastHeartbeat:    now,
			InputNumbers:     entities.InputNumbers,
			Lamps:            entities.Lamps,
			WateringZones:    entities.WateringZones,
			CreatedAt:        now,
			UpdatedAt:        now,
		}

		_, err = cm.db.Database.Collection("room_chambers").InsertOne(ctx, roomChamber)
		if err != nil {
			return nil, fmt.Errorf("failed to create room chamber: %w", err)
		}

		log.Printf("New room chamber created: %s (%s)", roomChamber.Name, roomSuffix)
	} else if err != nil {
		return nil, fmt.Errorf("failed to query room chamber: %w", err)
	} else {
		// Обновляем существующую room chamber
		update := bson.M{
			"$set": bson.M{
				"name":           roomChamberName,
				"last_heartbeat": now,
				"input_numbers":  entities.InputNumbers,
				"lamps":          entities.Lamps,
				"watering_zones": entities.WateringZones,
				"updated_at":     now,
			},
		}

		_, err = cm.db.Database.Collection("room_chambers").UpdateByID(ctx, roomChamber.ID, update)
		if err != nil {
			return nil, fmt.Errorf("failed to update room chamber: %w", err)
		}

		// Обновляем локальные данные
		roomChamber.Name = roomChamberName
		roomChamber.LastHeartbeat = now
		roomChamber.InputNumbers = entities.InputNumbers
		roomChamber.Lamps = entities.Lamps
		roomChamber.WateringZones = entities.WateringZones
		roomChamber.UpdatedAt = now

		log.Printf("Room chamber updated: %s (%s)", roomChamber.Name, roomSuffix)
	}

	// Логируем обнаруженные entities
	log.Printf("Room chamber %s entities:", roomChamber.Name)
	log.Printf("  - %d input numbers", len(roomChamber.InputNumbers))
	log.Printf("  - %d lamps", len(roomChamber.Lamps))
	log.Printf("  - %d watering zones", len(roomChamber.WateringZones))

	return &roomChamber, nil
}

// GetParentChamber возвращает родительскую chamber
func (cm *ChamberManager) GetParentChamber() *models.Chamber {
	return cm.parentChamber
}

// GetRoomChambers возвращает все room chambers
func (cm *ChamberManager) GetRoomChambers() map[string]*models.RoomChamber {
	return cm.roomChambers
}

// GetRoomChamber возвращает room chamber по суффиксу
func (cm *ChamberManager) GetRoomChamber(roomSuffix string) *models.RoomChamber {
	return cm.roomChambers[roomSuffix]
}

// UpdateHeartbeat обновляет heartbeat для всех chambers
func (cm *ChamberManager) UpdateHeartbeat(ctx context.Context) error {
	now := time.Now()

	// Обновляем parent chamber
	if cm.parentChamber != nil {
		update := bson.M{
			"$set": bson.M{
				"last_heartbeat": now,
				"updated_at":     now,
			},
		}
		_, err := cm.db.ChambersCollection.UpdateByID(ctx, cm.parentChamber.ID, update)
		if err != nil {
			log.Printf("Failed to update parent chamber heartbeat: %v", err)
		} else {
			cm.parentChamber.LastHeartbeat = now
			cm.parentChamber.UpdatedAt = now
		}
	}

	// Обновляем room chambers
	for roomSuffix, roomChamber := range cm.roomChambers {
		update := bson.M{
			"$set": bson.M{
				"last_heartbeat": now,
				"updated_at":     now,
			},
		}
		_, err := cm.db.Database.Collection("room_chambers").UpdateByID(ctx, roomChamber.ID, update)
		if err != nil {
			log.Printf("Failed to update room chamber heartbeat for %s: %v", roomSuffix, err)
		} else {
			roomChamber.LastHeartbeat = now
			roomChamber.UpdatedAt = now
		}
	}

	return nil
}

// RegisterRoomChambersWithBackend регистрирует все room chambers с бэкендом
func (cm *ChamberManager) RegisterRoomChambersWithBackend(registrationService *RegistrationService) error {
	for roomSuffix, roomChamber := range cm.roomChambers {
		if roomChamber.BackendID.IsZero() {
			log.Printf("Registering room chamber '%s' with backend...", roomSuffix)
			if err := registrationService.RegisterRoomChamberWithBackend(roomChamber); err != nil {
				log.Printf("Warning: Failed to register room chamber '%s': %v", roomSuffix, err)
				continue
			}
		}
	}
	return nil
}
