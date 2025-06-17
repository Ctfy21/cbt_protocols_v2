package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ChamberStatus represents the status of a chamber
type ChamberStatus string

const (
	StatusOnline  ChamberStatus = "online"
	StatusOffline ChamberStatus = "offline"
)

// Chamber represents a climate chamber in the system
type Chamber struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name               string             `bson:"name" json:"name"`
	Suffix             string             `bson:"suffix" json:"suffix"` // e.g., "galo", "sb4", "default"
	BackendID          primitive.ObjectID `bson:"backend_id,omitempty" json:"backend_id,omitempty"`
	LocalIP            string             `bson:"local_ip" json:"local_ip"`
	HomeAssistantURL   string             `bson:"ha_url" json:"ha_url"`
	Status             ChamberStatus      `bson:"status" json:"status"`
	LastHeartbeat      time.Time          `bson:"last_heartbeat" json:"last_heartbeat"`
	DiscoveryCompleted bool               `bson:"discovery_completed" json:"discovery_completed"`
	Config             ChamberConfig      `bson:"config" json:"config"`
	CreatedAt          time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt          time.Time          `bson:"updated_at" json:"updated_at"`
}

// ChamberConfig represents chamber configuration
type ChamberConfig struct {
	InputNumbers  []InputNumber  `bson:"input_numbers" json:"input_numbers"`
	Lamps         []Lamp         `bson:"lamps" json:"lamps"`
	WateringZones []WateringZone `bson:"watering_zones" json:"watering_zones"`
	UpdatedAt     time.Time      `bson:"updated_at" json:"updated_at"`
	SyncedAt      *time.Time     `bson:"synced_at,omitempty" json:"synced_at,omitempty"`
}

// InputNumber represents a Home Assistant input_number entity
type InputNumber struct {
	EntityID     string  `bson:"entity_id" json:"entity_id"`
	Name         string  `bson:"name" json:"name"`
	Type         string  `bson:"type" json:"type"` // day_start, day_duration, temp_day, temp_night, etc.
	Min          float64 `bson:"min" json:"min"`
	Max          float64 `bson:"max" json:"max"`
	Step         float64 `bson:"step" json:"step"`
	CurrentValue float64 `bson:"current_value" json:"current_value"`
	Unit         string  `bson:"unit" json:"unit"`
}

// Lamp represents a light control entity
type Lamp struct {
	Name         string  `bson:"name" json:"name"`
	EntityID     string  `bson:"entity_id" json:"entity_id"`
	IntensityMin float64 `bson:"intensity_min" json:"intensity_min"`
	IntensityMax float64 `bson:"intensity_max" json:"intensity_max"`
	CurrentValue float64 `bson:"current_value" json:"current_value"`
}

// WateringZone represents a watering zone with its control parameters
type WateringZone struct {
	Name                 string `bson:"name" json:"name"`
	StartTimeEntityID    string `bson:"start_time_entity_id" json:"start_time_entity_id"`
	PeriodEntityID       string `bson:"period_entity_id" json:"period_entity_id"`
	PauseBetweenEntityID string `bson:"pause_between_entity_id" json:"pause_between_entity_id"`
	DurationEntityID     string `bson:"duration_entity_id" json:"duration_entity_id"`
}

// InputNumberType constants
const (
	InputNumberDayStart         = "day_start"
	InputNumberDayDuration      = "day_duration"
	InputNumberTempDay          = "temp_day"
	InputNumberTempNight        = "temp_night"
	InputNumberHumidityDay      = "humidity_day"
	InputNumberHumidityNight    = "humidity_night"
	InputNumberCO2Day           = "co2_day"
	InputNumberCO2Night         = "co2_night"
	InputNumberWateringStart    = "watering_start"
	InputNumberWateringPeriod   = "watering_period"
	InputNumberWateringPause    = "watering_pause"
	InputNumberWateringDuration = "watering_duration"
)

// InputNumberSubstrings defines the substrings to search for each input number type
var InputNumberSubstrings = map[string][]string{
	InputNumberDayStart:         {"hours_day", "hour_day", "day_start"},
	InputNumberDayDuration:      {"hours_work", "hour_work", "day_duration"},
	InputNumberTempDay:          {"temp_day", "temp_set_day", "temp_day_set"},
	InputNumberTempNight:        {"temp_night", "temp_set_night", "temp_night_set"},
	InputNumberHumidityDay:      {"hum_day", "hum_set_day", "hum_day_set"},
	InputNumberHumidityNight:    {"hum_night", "hum_set_night", "hum_night_set"},
	InputNumberCO2Day:           {"co2_day", "co2_set_day", "co2_day_set"},
	InputNumberCO2Night:         {"co2_night", "co2_set_night", "co2_night_set"},
	InputNumberWateringStart:    {"day_watering", "watering_start", "start_watering"},
	InputNumberWateringPeriod:   {"work_watering", "watering_period", "period_watering"},
	InputNumberWateringPause:    {"wait_watering", "pause_between", "pause_between_watering"},
	InputNumberWateringDuration: {"time_watering", "watering_seconds", "duration_watering"},
}

// GetInputNumbersByType returns all input numbers of a specific type
func (c *Chamber) GetInputNumbersByType(inputType string) []InputNumber {
	var result []InputNumber
	for _, in := range c.Config.InputNumbers {
		if in.Type == inputType {
			result = append(result, in)
		}
	}
	return result
}

// GetWateringInputNumbers returns all watering-related input numbers grouped by zone
func (c *Chamber) GetWateringInputNumbers() map[string]map[string]*InputNumber {
	wateringInputs := make(map[string]map[string]*InputNumber)

	// Process each watering zone
	for _, zone := range c.Config.WateringZones {
		zoneInputs := make(map[string]*InputNumber)

		// Find matching input numbers for this zone
		for i := range c.Config.InputNumbers {
			in := &c.Config.InputNumbers[i]
			switch in.EntityID {
			case zone.StartTimeEntityID:
				zoneInputs["start_time"] = in
			case zone.PeriodEntityID:
				zoneInputs["period"] = in
			case zone.PauseBetweenEntityID:
				zoneInputs["pause"] = in
			case zone.DurationEntityID:
				zoneInputs["duration"] = in
			}
		}

		if len(zoneInputs) > 0 {
			wateringInputs[zone.Name] = zoneInputs
		}
	}

	return wateringInputs
}
