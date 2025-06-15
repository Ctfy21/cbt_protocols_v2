package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RoomChamber represents a room-specific chamber (sub-chamber of a main chamber)
type RoomChamber struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name            string             `bson:"name" json:"name"`
	RoomSuffix      string             `bson:"room_suffix" json:"room_suffix"`
	ParentChamberID primitive.ObjectID `bson:"parent_chamber_id" json:"parent_chamber_id"`
	Location        string             `bson:"location" json:"location"`
	HAUrl           string             `bson:"ha_url" json:"ha_url"`
	AccessToken     string             `bson:"access_token" json:"-"`
	LocalIP         string             `bson:"local_ip" json:"local_ip"`
	Status          ChamberStatus      `bson:"status" json:"status"`
	LastHeartbeat   time.Time          `bson:"last_heartbeat" json:"last_heartbeat"`
	InputNumbers    []InputNumber      `bson:"input_numbers" json:"input_numbers"`
	Lamps           []Lamp             `bson:"lamps" json:"lamps"`
	WateringZones   []WateringZone     `bson:"watering_zones" json:"watering_zones"`
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at"`
}

// RoomChamberResponse represents room chamber data for API responses
type RoomChamberResponse struct {
	ID                primitive.ObjectID `json:"id"`
	Name              string             `json:"name"`
	RoomSuffix        string             `json:"room_suffix"`
	ParentChamberID   primitive.ObjectID `json:"parent_chamber_id"`
	ParentChamberName string             `json:"parent_chamber_name,omitempty"`
	Location          string             `json:"location"`
	HAUrl             string             `json:"ha_url"`
	LocalIP           string             `json:"local_ip"`
	Status            ChamberStatus      `json:"status"`
	LastHeartbeat     time.Time          `json:"last_heartbeat"`
	InputNumbers      []InputNumber      `json:"input_numbers"`
	Lamps             []Lamp             `json:"lamps"`
	WateringZones     []WateringZone     `json:"watering_zones"`
	CreatedAt         time.Time          `json:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at"`
}

// GetInputNumbersByType returns all input numbers of a specific type for room chamber
func (rc *RoomChamber) GetInputNumbersByType(inputType string) []InputNumber {
	var result []InputNumber
	for _, in := range rc.InputNumbers {
		if in.Type == inputType {
			result = append(result, in)
		}
	}
	return result
}

// GetWateringInputNumbers returns all watering-related input numbers grouped by zone for room chamber
func (rc *RoomChamber) GetWateringInputNumbers() map[string]map[string]*InputNumber {
	wateringInputs := make(map[string]map[string]*InputNumber)

	// Process each watering zone
	for _, zone := range rc.WateringZones {
		zoneInputs := make(map[string]*InputNumber)

		// Find matching input numbers for this zone
		for i := range rc.InputNumbers {
			in := &rc.InputNumbers[i]
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
