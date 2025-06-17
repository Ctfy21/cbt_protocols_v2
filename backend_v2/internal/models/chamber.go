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

// Chamber represents a climate chamber
type Chamber struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name          string             `bson:"name" json:"name"`
	Location      string             `bson:"location" json:"location"`
	HAUrl         string             `bson:"ha_url" json:"ha_url"`
	AccessToken   string             `bson:"access_token" json:"-"`
	LocalIP       string             `bson:"local_ip" json:"local_ip"`
	Status        ChamberStatus      `bson:"status" json:"status"`
	LastHeartbeat time.Time          `bson:"last_heartbeat" json:"last_heartbeat"`
	InputNumbers  []InputNumber      `bson:"input_numbers" json:"input_numbers"`
	Lamps         []Lamp             `bson:"lamps" json:"lamps"`
	WateringZones []WateringZone     `bson:"watering_zones" json:"watering_zones"`
	Config        *ChamberConfig     `bson:"config,omitempty" json:"config,omitempty"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at"`
}

// ChamberConfig represents chamber configuration parameters
type ChamberConfig struct {
	ID             primitive.ObjectID            `bson:"_id,omitempty" json:"id"`
	ChamberID      primitive.ObjectID            `bson:"chamber_id" json:"chamber_id"`
	DayDuration    map[string]float64            `bson:"day_duration" json:"day_duration"`
	DayStart       map[string]float64            `bson:"day_start" json:"day_start"`
	Temperature    map[string]map[string]float64 `bson:"temperature" json:"temperature"` // day/night -> entity_id -> value
	Humidity       map[string]map[string]float64 `bson:"humidity" json:"humidity"`       // day/night -> entity_id -> value
	CO2            map[string]map[string]float64 `bson:"co2" json:"co2"`                 // day/night -> entity_id -> value
	LightIntensity map[string]float64            `bson:"light_intensity" json:"light_intensity"`
	WateringZones  map[string]map[string]float64 `bson:"watering_zones" json:"watering_zones"` // zone_name -> param_type -> value
	UpdatedAt      time.Time                     `bson:"updated_at" json:"updated_at"`
	SyncedAt       *time.Time                    `bson:"synced_at,omitempty" json:"synced_at,omitempty"`
}

// InputNumber represents a Home Assistant input number entity
type InputNumber struct {
	EntityID     string  `bson:"entity_id" json:"entity_id"`
	Name         string  `bson:"name" json:"name"`
	FriendlyName string  `bson:"friendly_name" json:"friendly_name"`
	Type         string  `bson:"type" json:"type"`
	Min          float64 `bson:"min" json:"min"`
	Max          float64 `bson:"max" json:"max"`
	Step         float64 `bson:"step" json:"step"`
	Value        float64 `bson:"value" json:"value"`
	CurrentValue float64 `bson:"current_value" json:"current_value"`
	Unit         string  `bson:"unit" json:"unit"`
}

// Lamp represents a lamp entity
type Lamp struct {
	Name         string  `bson:"name" json:"name"`
	EntityID     string  `bson:"entity_id" json:"entity_id"`
	FriendlyName string  `bson:"friendly_name" json:"friendly_name"`
	IntensityMin float64 `bson:"intensity_min" json:"intensity_min"`
	IntensityMax float64 `bson:"intensity_max" json:"intensity_max"`
	CurrentValue float64 `bson:"current_value" json:"current_value"`
}

// WateringZone represents a watering zone configuration
type WateringZone struct {
	Name                 string `bson:"name" json:"name"`
	StartTimeEntityID    string `bson:"start_time_entity_id" json:"start_time_entity_id"`
	PeriodEntityID       string `bson:"period_entity_id" json:"period_entity_id"`
	PauseBetweenEntityID string `bson:"pause_between_entity_id" json:"pause_between_entity_id"`
	DurationEntityID     string `bson:"duration_entity_id" json:"duration_entity_id"`
}

// InputNumberType constants - these match what local_api_v2 uses
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

// GetInputNumbersByType returns all input numbers of a specific type
func (c *Chamber) GetInputNumbersByType(inputType string) []InputNumber {
	var result []InputNumber
	for _, in := range c.InputNumbers {
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
	for _, zone := range c.WateringZones {
		zoneInputs := make(map[string]*InputNumber)

		// Find matching input numbers for this zone
		for i := range c.InputNumbers {
			in := &c.InputNumbers[i]
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

// InitializeConfig initializes the config for a chamber
func (c *Chamber) InitializeConfig() {
	if c.Config == nil {
		c.Config = &ChamberConfig{
			ID:             primitive.NewObjectID(),
			ChamberID:      c.ID,
			DayDuration:    make(map[string]float64),
			DayStart:       make(map[string]float64),
			Temperature:    make(map[string]map[string]float64),
			Humidity:       make(map[string]map[string]float64),
			CO2:            make(map[string]map[string]float64),
			LightIntensity: make(map[string]float64),
			WateringZones:  make(map[string]map[string]float64),
			UpdatedAt:      time.Now(),
		}

		// Initialize sub-maps
		c.Config.Temperature["day"] = make(map[string]float64)
		c.Config.Temperature["night"] = make(map[string]float64)
		c.Config.Humidity["day"] = make(map[string]float64)
		c.Config.Humidity["night"] = make(map[string]float64)
		c.Config.CO2["day"] = make(map[string]float64)
		c.Config.CO2["night"] = make(map[string]float64)
	}
}
