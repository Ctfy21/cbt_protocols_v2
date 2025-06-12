package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ExperimentStatus represents the status of an experiment
type ExperimentStatus string

const (
	ExperimentStatusActive    ExperimentStatus = "active"
	ExperimentStatusInactive  ExperimentStatus = "inactive"
	ExperimentStatusDraft     ExperimentStatus = "draft"
	ExperimentStatusCompleted ExperimentStatus = "completed"
)

// Experiment represents an experiment
type Experiment struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title            string             `bson:"title" json:"title"`
	Description      string             `bson:"description" json:"description"`
	Status           ExperimentStatus   `bson:"status" json:"status"`
	ChamberID        primitive.ObjectID `bson:"chamber_id" json:"chamber_id"`
	Phases           []Phase            `bson:"phases" json:"phases"`
	Schedule         []ScheduleItem     `bson:"schedule" json:"schedule"`
	ActivePhaseIndex *int               `bson:"active_phase_index,omitempty" json:"active_phase_index,omitempty"`
	CreatedAt        time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time          `bson:"updated_at" json:"updated_at"`
}

// Phase represents a phase in an experiment
type Phase struct {
	Title          string                       `bson:"title" json:"title"`
	Description    string                       `bson:"description" json:"description"`
	DurationDays   int                          `bson:"duration_days" json:"duration_days"`
	InputNumbers   map[string]*PhaseInputNumber `bson:"input_numbers" json:"input_numbers"`
	LightIntensity map[string]*LightIntensity   `bson:"light_intensity" json:"light_intensity"`
	WateringZones  map[string]*WateringZone     `bson:"watering_zones" json:"watering_zones"`
}

// PhaseInputNumber represents an input number value for a phase
type PhaseInputNumber struct {
	EntityID string  `bson:"entity_id" json:"entity_id"`
	Value    float64 `bson:"value" json:"value"`
}

// LightIntensity represents light intensity settings
type LightIntensity struct {
	EntityID  string  `bson:"entity_id" json:"entity_id"`
	Intensity float64 `bson:"intensity" json:"intensity"`
}

// ScheduleItem represents a schedule item for an experiment
type ScheduleItem struct {
	PhaseIndex     int    `bson:"phase_index" json:"phase_index"`
	StartDate      string `bson:"start_date" json:"start_date"`
	EndDate        string `bson:"end_date" json:"end_date"`
	StartTimestamp int64  `bson:"start_timestamp" json:"start_timestamp"`
	EndTimestamp   int64  `bson:"end_timestamp" json:"end_timestamp"`
}
