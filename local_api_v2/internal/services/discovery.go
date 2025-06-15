package services

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"local_api_v2/internal/models"
	"local_api_v2/pkg/homeassistant"
)

// DiscoveryService handles discovery of Home Assistant entities
type DiscoveryService struct {
	haClient *homeassistant.Client
}

// NewDiscoveryService creates a new discovery service
func NewDiscoveryService(haClient *homeassistant.Client) *DiscoveryService {
	return &DiscoveryService{
		haClient: haClient,
	}
}

// RoomEntities represents entities grouped by room suffix
type RoomEntities struct {
	RoomSuffix    string
	InputNumbers  []models.InputNumber
	Lamps         []models.Lamp
	WateringZones []models.WateringZone
}

// DiscoverInputNumbers discovers and categorizes input_number entities
func (s *DiscoveryService) DiscoverInputNumbers() ([]models.InputNumber, []models.Lamp, []models.WateringZone, error) {
	// Get all input numbers from Home Assistant
	haEntities, err := s.haClient.GetInputNumbers()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get input numbers: %v", err)
	}

	var (
		inputNumbers  []models.InputNumber
		lamps         []models.Lamp
		wateringZones []models.WateringZone
		lampMap       = make(map[string]*models.Lamp)
		wateringMap   = make(map[string]*models.WateringZone)
	)

	// Process each entity
	for _, entity := range haEntities {
		entityID := entity.EntityID
		lowerEntityID := strings.ToLower(entityID)
		friendlyName := entity.FriendlyName

		if strings.Contains(entityID, "prog") || strings.Contains(entityID, "test") {
			continue
		}

		log.Printf("Processing entity: %s (%s)", entityID, friendlyName)

		// Check if it's a lamp control
		if isLampEntity(lowerEntityID, friendlyName) {
			lampName := extractLampName(entityID, friendlyName)
			if lamp, exists := lampMap[lampName]; exists {
				// Update existing lamp
				lamp.EntityID = entityID
				lamp.IntensityMin = entity.Min
				lamp.IntensityMax = entity.Max
				lamp.CurrentValue = entity.Value
			} else {
				// Create new lamp
				lampMap[lampName] = &models.Lamp{
					Name:         lampName,
					EntityID:     entityID,
					IntensityMin: entity.Min,
					IntensityMax: entity.Max,
					CurrentValue: entity.Value,
				}
			}
			continue
		}

		// Check if it's a watering control
		wateringType, zoneName := getWateringType(lowerEntityID, friendlyName)
		if wateringType != "" {
			if zone, exists := wateringMap[zoneName]; exists {
				// Update existing zone
				switch wateringType {
				case models.InputNumberWateringStart:
					zone.StartTimeEntityID = entityID
				case models.InputNumberWateringPeriod:
					zone.PeriodEntityID = entityID
				case models.InputNumberWateringPause:
					zone.PauseBetweenEntityID = entityID
				case models.InputNumberWateringDuration:
					zone.DurationEntityID = entityID
				}
			} else {
				// Create new zone
				zone := &models.WateringZone{Name: zoneName}
				switch wateringType {
				case models.InputNumberWateringStart:
					zone.StartTimeEntityID = entityID
				case models.InputNumberWateringPeriod:
					zone.PeriodEntityID = entityID
				case models.InputNumberWateringPause:
					zone.PauseBetweenEntityID = entityID
				case models.InputNumberWateringDuration:
					zone.DurationEntityID = entityID
				}
				wateringMap[zoneName] = zone
			}
			continue
		}

		// Check if it's a regular input number (climate control)
		inputType := getInputNumberType(lowerEntityID, friendlyName)
		if inputType != "" {
			inputNumber := models.InputNumber{
				EntityID:     entityID,
				Name:         friendlyName,
				Type:         inputType,
				Min:          entity.Min,
				Max:          entity.Max,
				Step:         entity.Step,
				CurrentValue: entity.Value,
				Unit:         entity.Unit,
			}
			inputNumbers = append(inputNumbers, inputNumber)
		}
	}

	// Convert maps to slices
	for _, lamp := range lampMap {
		lamps = append(lamps, *lamp)
	}
	for _, zone := range wateringMap {
		wateringZones = append(wateringZones, *zone)
	}

	log.Printf("Discovered: %d input numbers, %d lamps, %d watering zones",
		len(inputNumbers), len(lamps), len(wateringZones))

	return inputNumbers, lamps, wateringZones, nil
}

// getInputNumberType determines the type of input number based on entity ID and friendly name
func getInputNumberType(entityID, friendlyName string) string {
	lowerID := strings.ToLower(entityID)
	lowerName := strings.ToLower(friendlyName)

	for inputType, substrings := range models.InputNumberSubstrings {
		// Skip watering types
		if strings.HasPrefix(inputType, "watering_") {
			continue
		}

		for _, substr := range substrings {
			if strings.Contains(lowerID, substr) || strings.Contains(lowerName, substr) {
				return inputType
			}
		}
	}

	return ""
}

// isLampEntity checks if the entity is a lamp control
func isLampEntity(entityID, friendlyName string) bool {
	lampKeywords := []string{"lamp", "light", "led", "лампа", "свет", "света", "лампы", "ppfd"}

	lowerID := strings.ToLower(entityID)
	lowerName := strings.ToLower(friendlyName)

	for _, keyword := range lampKeywords {
		if strings.Contains(lowerID, keyword) || strings.Contains(lowerName, keyword) {
			return true
		}
	}

	return false
}

// extractLampName extracts the lamp name from entity ID or friendly name
func extractLampName(entityID, friendlyName string) string {
	// Try to extract from friendly name first
	if friendlyName != "" && friendlyName != entityID {
		// Remove common prefixes/suffixes
		name := friendlyName
		name = strings.TrimPrefix(name, "Lamp ")
		name = strings.TrimPrefix(name, "Light ")
		name = strings.TrimSuffix(name, " Intensity")
		name = strings.TrimSuffix(name, " Brightness")
		name = strings.TrimSuffix(name, " Света")
		name = strings.TrimSuffix(name, " Лампа")
		return name
	}

	// Extract from entity ID
	parts := strings.Split(entityID, ".")
	if len(parts) > 1 {
		name := parts[1]
		name = strings.ReplaceAll(name, "_", " ")
		name = strings.Title(name)
		return name
	}

	return "Lamp"
}

// getWateringType determines if entity is a watering control and returns its type and zone name
func getWateringType(entityID, friendlyName string) (string, string) {
	lowerID := strings.ToLower(entityID)
	lowerName := strings.ToLower(friendlyName)

	// Check each watering type
	wateringTypes := []string{
		models.InputNumberWateringStart,
		models.InputNumberWateringPeriod,
		models.InputNumberWateringPause,
		models.InputNumberWateringDuration,
	}

	for _, wateringType := range wateringTypes {
		if substrings, ok := models.InputNumberSubstrings[wateringType]; ok {
			for _, substr := range substrings {
				if strings.Contains(lowerID, substr) || strings.Contains(lowerName, substr) {
					// Extract zone name
					zoneName := extractZoneName(entityID, friendlyName)
					return wateringType, zoneName
				}
			}
		}
	}

	return "", ""
}

// extractZoneName extracts the watering zone name
func extractZoneName(entityID, friendlyName string) string {
	// Common zone indicators
	zoneKeywords := []string{"zone", "зона", "area", "участок"}

	// Try to extract from friendly name
	if friendlyName != "" {
		for _, keyword := range zoneKeywords {
			if idx := strings.Index(strings.ToLower(friendlyName), keyword); idx != -1 {
				// Extract zone identifier after the keyword
				parts := strings.Fields(friendlyName[idx:])
				if len(parts) > 1 {
					return fmt.Sprintf("Zone %s", parts[1])
				}
			}
		}
	}

	// Try to extract from entity ID
	parts := strings.Split(entityID, "_")
	for i, part := range parts {
		for _, keyword := range zoneKeywords {
			if strings.Contains(strings.ToLower(part), keyword) && i+1 < len(parts) {
				return fmt.Sprintf("Zone %s", parts[i+1])
			}
		}
	}

	// Default zone name
	return "Zone 1"
}

// extractRoomSuffix extracts room suffix from entity ID (e.g., "room1", "room2", "midi_room1")
func extractRoomSuffix(entityID string) string {
	// Регулярное выражение для поиска суффиксов вида room1, room2, midi_room1 и т.д.
	re := regexp.MustCompile(`(room\d+)$`)
	matches := re.FindStringSubmatch(strings.ToLower(entityID))
	if len(matches) > 1 {
		return matches[1]
	}

	// Также ищем паттерны вида midi_room1, watering_room1
	re2 := regexp.MustCompile(`_?(room\d+)$`)
	matches2 := re2.FindStringSubmatch(strings.ToLower(entityID))
	if len(matches2) > 1 {
		return matches2[1]
	}

	return ""
}

// getRoomBaseName extracts base name by removing room suffix
func getRoomBaseName(entityID, roomSuffix string) string {
	if roomSuffix == "" {
		return entityID
	}

	// Удаляем суффикс комнаты из entity ID
	lowerID := strings.ToLower(entityID)
	lowerSuffix := strings.ToLower(roomSuffix)

	// Удаляем суффикс с возможным префиксом "_"
	if strings.HasSuffix(lowerID, "_"+lowerSuffix) {
		return entityID[:len(entityID)-len("_"+lowerSuffix)]
	} else if strings.HasSuffix(lowerID, lowerSuffix) {
		return entityID[:len(entityID)-len(lowerSuffix)]
	}

	return entityID
}

// DiscoverRoomEntities discovers entities grouped by room suffixes
func (s *DiscoveryService) DiscoverRoomEntities() (map[string]*RoomEntities, error) {
	// Get all input numbers from Home Assistant
	haEntities, err := s.haClient.GetInputNumbers()
	if err != nil {
		return nil, fmt.Errorf("failed to get input numbers: %v", err)
	}

	roomMap := make(map[string]*RoomEntities)

	// Сначала собираем все entity по комнатам
	for _, entity := range haEntities {
		entityID := entity.EntityID
		lowerEntityID := strings.ToLower(entityID)
		friendlyName := entity.FriendlyName

		if strings.Contains(entityID, "prog") || strings.Contains(entityID, "test") {
			continue
		}

		roomSuffix := extractRoomSuffix(entityID)
		if roomSuffix == "" {
			roomSuffix = "default" // Для entity без суффикса комнаты
		}

		// Создаем или получаем комнату
		if _, exists := roomMap[roomSuffix]; !exists {
			roomMap[roomSuffix] = &RoomEntities{
				RoomSuffix:    roomSuffix,
				InputNumbers:  []models.InputNumber{},
				Lamps:         []models.Lamp{},
				WateringZones: []models.WateringZone{},
			}
		}

		room := roomMap[roomSuffix]

		log.Printf("Processing entity for room '%s': %s (%s)", roomSuffix, entityID, friendlyName)

		// Обработка ламп
		if isLampEntity(lowerEntityID, friendlyName) {
			lampName := extractLampName(entityID, friendlyName)
			lamp := models.Lamp{
				Name:         lampName,
				EntityID:     entityID,
				IntensityMin: entity.Min,
				IntensityMax: entity.Max,
				CurrentValue: entity.Value,
			}
			room.Lamps = append(room.Lamps, lamp)
			continue
		}

		// Обработка полива
		wateringType, zoneName := getWateringType(lowerEntityID, friendlyName)
		if wateringType != "" {
			// Ищем существующую зону полива для этой комнаты
			var targetZone *models.WateringZone
			for i := range room.WateringZones {
				if room.WateringZones[i].Name == zoneName {
					targetZone = &room.WateringZones[i]
					break
				}
			}

			if targetZone == nil {
				// Создаем новую зону
				newZone := models.WateringZone{Name: zoneName}
				room.WateringZones = append(room.WateringZones, newZone)
				targetZone = &room.WateringZones[len(room.WateringZones)-1]
			}

			// Устанавливаем соответствующий entity ID
			switch wateringType {
			case models.InputNumberWateringStart:
				targetZone.StartTimeEntityID = entityID
			case models.InputNumberWateringPeriod:
				targetZone.PeriodEntityID = entityID
			case models.InputNumberWateringPause:
				targetZone.PauseBetweenEntityID = entityID
			case models.InputNumberWateringDuration:
				targetZone.DurationEntityID = entityID
			}
			continue
		}

		// Обработка обычных input numbers (климат контроль)
		inputType := getInputNumberType(lowerEntityID, friendlyName)
		if inputType != "" {
			inputNumber := models.InputNumber{
				EntityID:     entityID,
				Name:         friendlyName,
				Type:         inputType,
				Min:          entity.Min,
				Max:          entity.Max,
				Step:         entity.Step,
				CurrentValue: entity.Value,
				Unit:         entity.Unit,
			}
			room.InputNumbers = append(room.InputNumbers, inputNumber)
		}
	}

	log.Printf("Discovered entities grouped by rooms:")
	for roomSuffix, room := range roomMap {
		log.Printf("  Room '%s': %d input numbers, %d lamps, %d watering zones",
			roomSuffix, len(room.InputNumbers), len(room.Lamps), len(room.WateringZones))
	}

	return roomMap, nil
}
