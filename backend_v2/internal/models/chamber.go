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
	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name               string             `bson:"name" json:"name"`
	Suffix             string             `bson:"suffix" json:"suffix"` // e.g., "galo", "sb4", "room1", "default"
	HAUrl              string             `bson:"ha_url" json:"ha_url"`
	LocalAPIversion    int                `bson:"local_api_version" json:"local_api_version"`
	TimeOffset         int                `bson:"time_offset" json:"time_offset"`
	AccessToken        string             `bson:"access_token" json:"-"`
	LocalIP            string             `bson:"local_ip" json:"local_ip"`
	Status             ChamberStatus      `bson:"status" json:"status"`
	LastHeartbeat      time.Time          `bson:"last_heartbeat" json:"last_heartbeat"`
	DiscoveryCompleted bool               `bson:"discovery_completed" json:"discovery_completed"`
	Config             *ChamberConfig     `bson:"config,omitempty" json:"config,omitempty"`
	CreatedAt          time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt          time.Time          `bson:"updated_at" json:"updated_at"`
}

// ChamberConfig represents chamber configuration parameters
type ChamberConfig struct {
	ID                   primitive.ObjectID                `bson:"_id,omitempty" json:"id"`
	ChamberID            primitive.ObjectID                `bson:"chamber_id" json:"chamber_id"`
	Lamps                map[string]InputNumber            `bson:"lamps" json:"lamps"`
	WateringZones        []WateringZone                    `bson:"watering_zones" json:"watering_zones"`
	UnrecognisedEntities map[string]InputNumber            `bson:"unrecognised_entities" json:"unrecognised_entities"`
	DayDuration          map[string]InputNumber            `bson:"day_duration" json:"day_duration"`
	DayStart             map[string]InputNumber            `bson:"day_start" json:"day_start"`
	Temperature          map[string]map[string]InputNumber `bson:"temperature" json:"temperature"` // day/night -> entity_id -> value
	Humidity             map[string]map[string]InputNumber `bson:"humidity" json:"humidity"`       // day/night -> entity_id -> value
	CO2                  map[string]map[string]InputNumber `bson:"co2" json:"co2"`                 // day/night -> entity_id -> value
	UpdatedAt            time.Time                         `bson:"updated_at" json:"updated_at"`
	SyncedAt             *time.Time                        `bson:"synced_at,omitempty" json:"synced_at,omitempty"`
}

// InputNumber represents a Home Assistant input_number entity
type InputNumber struct {
	EntityID string  `bson:"entity_id" json:"entity_id"`
	Name     string  `bson:"name" json:"name"`
	Type     string  `bson:"type" json:"type"`
	Min      float64 `bson:"min" json:"min"`
	Max      float64 `bson:"max" json:"max"`
	Step     float64 `bson:"step" json:"step"`
	Value    float64 `bson:"value" json:"value"`
	Unit     string  `bson:"unit" json:"unit"`
}

// WateringZone represents a watering zone configuration
type WateringZone struct {
	Name                 string                 `bson:"name" json:"name"`
	StartTimeEntityID    map[string]InputNumber `bson:"start_time_entity_id" json:"start_time_entity_id"`
	PeriodEntityID       map[string]InputNumber `bson:"period_entity_id" json:"period_entity_id"`
	PauseBetweenEntityID map[string]InputNumber `bson:"pause_between_entity_id" json:"pause_between_entity_id"`
	DurationEntityID     map[string]InputNumber `bson:"duration_entity_id" json:"duration_entity_id"`
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
	InputNumberUnrecognised     = "unrecognised"
)

// InputNumberSubstrings defines the substrings to search for each input number type
// This is used by local_api_v2 for entity discovery
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

// InitializeConfig initializes the config for a chamber
func (c *Chamber) InitializeConfig() {
	if c.Config == nil {
		c.Config = &ChamberConfig{
			ID:                   primitive.NewObjectID(),
			ChamberID:            c.ID,
			Lamps:                make(map[string]InputNumber),
			WateringZones:        []WateringZone{},
			UnrecognisedEntities: make(map[string]InputNumber),
			DayDuration:          make(map[string]InputNumber),
			DayStart:             make(map[string]InputNumber),
			Temperature:          make(map[string]map[string]InputNumber),
			Humidity:             make(map[string]map[string]InputNumber),
			CO2:                  make(map[string]map[string]InputNumber),
			UpdatedAt:            time.Now(),
		}

		// Initialize sub-maps
		c.Config.Temperature["day"] = make(map[string]InputNumber)
		c.Config.Temperature["night"] = make(map[string]InputNumber)
		c.Config.Humidity["day"] = make(map[string]InputNumber)
		c.Config.Humidity["night"] = make(map[string]InputNumber)
		c.Config.CO2["day"] = make(map[string]InputNumber)
		c.Config.CO2["night"] = make(map[string]InputNumber)
	}
}
