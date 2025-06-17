package services

import (
	"fmt"
	"log"
	"strings"

	"local_api_v2/internal/models"
	"local_api_v2/pkg/homeassistant"
)

// DiscoveryService handles discovery of Home Assistant entities
type DiscoveryService struct {
	haClient        *homeassistant.Client
	chamberSuffixes []string // Настраиваемые суффиксы камер
}

// NewDiscoveryService creates a new discovery service
func NewDiscoveryService(haClient *homeassistant.Client) *DiscoveryService {
	return &DiscoveryService{
		haClient:        haClient,
		chamberSuffixes: []string{}, // Будет установлено позже через SetChamberSuffixes
	}
}

// SetChamberSuffixes устанавливает список поддерживаемых суффиксов камер
func (s *DiscoveryService) SetChamberSuffixes(suffixes []string) {
	s.chamberSuffixes = suffixes
	log.Printf("Discovery service configured with chamber suffixes: %v", suffixes)
}

// ChamberEntities represents entities grouped by room suffix
type ChamberEntities struct {
	RoomSuffix string
	Config     models.ChamberConfig
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

// extractRoomSuffix extracts room suffix from entity ID with support for custom chamber suffixes
func (s *DiscoveryService) extractRoomSuffix(entityID string) string {
	lowerEntityID := strings.ToLower(entityID)

	// Сначала проверяем настраиваемые суффиксы (galo, sb4, oreol, sb1, etc.)
	for _, suffix := range s.chamberSuffixes {
		lowerSuffix := strings.ToLower(suffix)

		// Проверяем различные варианты окончаний:
		// 1. _suffix в конце (например, input_number.temp_galo)
		if strings.HasSuffix(lowerEntityID, "_"+lowerSuffix) {
			log.Printf("Found chamber suffix '%s' in entity '%s' (pattern: _suffix)", suffix, entityID)
			return lowerSuffix
		}

		// 2. suffix в конце (например, input_number.tempgalo)
		if strings.HasSuffix(lowerEntityID, lowerSuffix) {
			// Проверяем, что это не просто совпадение в середине слова
			suffixStart := len(lowerEntityID) - len(lowerSuffix)
			if suffixStart > 0 {
				prevChar := lowerEntityID[suffixStart-1]
				// Суффикс должен быть отделен символом, не буквой
				if prevChar == '_' || prevChar == '.' || prevChar == '-' {
					log.Printf("Found chamber suffix '%s' in entity '%s' (pattern: suffix)", suffix, entityID)
					return lowerSuffix
				}
			} else {
				// Суффикс в самом начале
				log.Printf("Found chamber suffix '%s' in entity '%s' (pattern: suffix)", suffix, entityID)
				return lowerSuffix
			}
		}

		// 3. suffix_ в любом месте (например, input_number.galo_temp)
		if strings.Contains(lowerEntityID, lowerSuffix+"_") {
			log.Printf("Found chamber suffix '%s' in entity '%s' (pattern: suffix_)", suffix, entityID)
			return lowerSuffix
		}

		// 4. _suffix_ в любом месте (например, input_number.temp_galo_day)
		if strings.Contains(lowerEntityID, "_"+lowerSuffix+"_") {
			log.Printf("Found chamber suffix '%s' in entity '%s' (pattern: _suffix_)", suffix, entityID)
			return lowerSuffix
		}
	}

	// Затем проверяем стандартные паттерны room1, room2, etc.
	// re := regexp.MustCompile(`(room\d+)$`)
	// matches := re.FindStringSubmatch(lowerEntityID)
	// if len(matches) > 1 {
	// 	log.Printf("Found room suffix '%s' in entity '%s' (pattern: room\\d+)", matches[1], entityID)
	// 	return matches[1]
	// }

	// // Также ищем паттерны вида midi_room1, watering_room1
	// re2 := regexp.MustCompile(`_?(room\d+)$`)
	// matches2 := re2.FindStringSubmatch(lowerEntityID)
	// if len(matches2) > 1 {
	// 	log.Printf("Found room suffix '%s' in entity '%s' (pattern: _room\\d+)", matches2[1], entityID)
	// 	return matches2[1]
	// }

	return ""
}

// DiscoverChamberEntities discovers entities grouped by room suffixes
func (s *DiscoveryService) DiscoverChamberEntities() (map[string]*ChamberEntities, error) {
	// Get all input numbers from Home Assistant
	haEntities, err := s.haClient.GetInputNumbers()
	if err != nil {
		return nil, fmt.Errorf("failed to get input numbers: %v", err)
	}

	roomMap, err := s.AutomaticalyDiscoverChamberEntities(haEntities)

	if err != nil {
		return nil, fmt.Errorf("failed to discover chamber entities: %v", err)
	}

	log.Printf("Discovered entities grouped by rooms:")
	for roomSuffix, room := range roomMap {
		log.Printf("  Room '%s': %d lamps, %d watering zones",
			roomSuffix, len(room.Config.Lamps), len(room.Config.WateringZones))
	}

	return roomMap, nil
}

func (s *DiscoveryService) AutomaticalyDiscoverChamberEntities(haEntities []homeassistant.InputNumberEntity) (map[string]*ChamberEntities, error) {
	roomMap := make(map[string]*ChamberEntities)

	log.Printf("Discovering room entities with configured suffixes: %v", s.chamberSuffixes)

	// First collect all entities by rooms
	for _, entity := range haEntities {
		entityID := entity.EntityID
		lowerEntityID := strings.ToLower(entityID)
		friendlyName := entity.FriendlyName

		if strings.Contains(entityID, "prog") || strings.Contains(entityID, "test") {
			continue
		}

		roomSuffix := s.extractRoomSuffix(entityID)
		if roomSuffix == "" {
			roomSuffix = "default" // For entities without room suffix
		}

		// Create or get room
		if _, exists := roomMap[roomSuffix]; !exists {
			roomMap[roomSuffix] = &ChamberEntities{
				RoomSuffix: roomSuffix,
				Config: models.ChamberConfig{
					Lamps:                []models.Lamp{},
					WateringZones:        []models.WateringZone{},
					UnrecognisedEntities: []models.InputNumber{},
					DayDuration:          make(map[string]float64),
					DayStart:             make(map[string]float64),
					Temperature:          map[string]map[string]float64{"day": {}, "night": {}},
					Humidity:             map[string]map[string]float64{"day": {}, "night": {}},
					CO2:                  map[string]map[string]float64{"day": {}, "night": {}},
					LightIntensity:       make(map[string]float64),
					WateringSettings:     make(map[string]map[string]float64),
				},
			}
		}

		room := roomMap[roomSuffix]
		entityProcessed := false

		log.Printf("Processing entity for room '%s': %s (%s)", roomSuffix, entityID, friendlyName)

		// Process lamps
		if isLampEntity(lowerEntityID, friendlyName) {
			lampName := extractLampName(entityID, friendlyName)
			lamp := models.Lamp{
				Name:         lampName,
				EntityID:     entityID,
				IntensityMin: entity.Min,
				IntensityMax: entity.Max,
				CurrentValue: entity.Value,
			}
			room.Config.Lamps = append(room.Config.Lamps, lamp)
			room.Config.LightIntensity[entityID] = entity.Value
			entityProcessed = true
		}

		// Process watering
		if !entityProcessed {
			wateringType, zoneName := getWateringType(lowerEntityID, friendlyName)
			if wateringType != "" {
				// Find existing watering zone for this room
				var targetZone *models.WateringZone
				for i := range room.Config.WateringZones {
					if room.Config.WateringZones[i].Name == zoneName {
						targetZone = &room.Config.WateringZones[i]
						break
					}
				}

				if targetZone == nil {
					// Create new zone
					newZone := models.WateringZone{Name: zoneName}
					room.Config.WateringZones = append(room.Config.WateringZones, newZone)
					targetZone = &room.Config.WateringZones[len(room.Config.WateringZones)-1]
				}

				// Set corresponding entity ID
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

				// Add to watering settings
				if room.Config.WateringSettings[zoneName] == nil {
					room.Config.WateringSettings[zoneName] = make(map[string]float64)
				}

				switch wateringType {
				case models.InputNumberWateringStart:
					room.Config.WateringSettings[zoneName]["start_time"] = entity.Value
				case models.InputNumberWateringPeriod:
					room.Config.WateringSettings[zoneName]["period"] = entity.Value
				case models.InputNumberWateringPause:
					room.Config.WateringSettings[zoneName]["pause"] = entity.Value
				case models.InputNumberWateringDuration:
					room.Config.WateringSettings[zoneName]["duration"] = entity.Value
				}

				entityProcessed = true
			}
		}

		// Process regular input numbers (climate control)
		if !entityProcessed {
			inputType := getInputNumberType(lowerEntityID, friendlyName)
			if inputType != "" {
				switch inputType {
				case models.InputNumberDayStart:
					room.Config.DayStart[entityID] = entity.Value
				case models.InputNumberDayDuration:
					room.Config.DayDuration[entityID] = entity.Value
				case models.InputNumberTempDay:
					room.Config.Temperature["day"][entityID] = entity.Value
				case models.InputNumberTempNight:
					room.Config.Temperature["night"][entityID] = entity.Value
				case models.InputNumberHumidityDay:
					room.Config.Humidity["day"][entityID] = entity.Value
				case models.InputNumberHumidityNight:
					room.Config.Humidity["night"][entityID] = entity.Value
				case models.InputNumberCO2Day:
					room.Config.CO2["day"][entityID] = entity.Value
				case models.InputNumberCO2Night:
					room.Config.CO2["night"][entityID] = entity.Value
				}
				entityProcessed = true
			}
		}

		// If entity was not processed, add it as unrecognised
		if !entityProcessed {
			inputNumber := models.InputNumber{
				EntityID:     entityID,
				Name:         friendlyName,
				Type:         models.InputNumberUnrecognised,
				Min:          entity.Min,
				Max:          entity.Max,
				Step:         entity.Step,
				CurrentValue: entity.Value,
				Unit:         entity.Unit,
			}
			room.Config.UnrecognisedEntities = append(room.Config.UnrecognisedEntities, inputNumber)
			log.Printf("Unrecognized entity added to room '%s': %s (%s)", roomSuffix, entityID, friendlyName)
		}
	}

	// Log summary for each room
	for roomSuffix, room := range roomMap {
		log.Printf("Room '%s' summary:", roomSuffix)
		log.Printf("  - %d lamps", len(room.Config.Lamps))
		log.Printf("  - %d watering zones", len(room.Config.WateringZones))
		log.Printf("  - %d unrecognised entities", len(room.Config.UnrecognisedEntities))
		log.Printf("  - Climate mappings: %d day_start, %d day_duration",
			len(room.Config.DayStart), len(room.Config.DayDuration))
		log.Printf("  - Temperature: %d day, %d night",
			len(room.Config.Temperature["day"]), len(room.Config.Temperature["night"]))
		log.Printf("  - Humidity: %d day, %d night",
			len(room.Config.Humidity["day"]), len(room.Config.Humidity["night"]))
		log.Printf("  - CO2: %d day, %d night",
			len(room.Config.CO2["day"]), len(room.Config.CO2["night"]))
	}

	return roomMap, nil
}
