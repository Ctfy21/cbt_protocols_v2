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
	Title                    string                    `bson:"title" json:"title"`
	Description              string                    `bson:"description" json:"description"`
	DurationDays             int                       `bson:"duration_days" json:"duration_days"`
	StartDay                 int                       `bson:"start_day,omitempty" json:"start_day,omitempty"`
	WorkDaySchedule          map[string]ScheduleConfig `bson:"work_day_schedule,omitempty" json:"work_day_schedule,omitempty"`
	TemperatureDaySchedule   map[string]ScheduleConfig `bson:"temperature_day_schedule,omitempty" json:"temperature_day_schedule,omitempty"`
	TemperatureNightSchedule map[string]ScheduleConfig `bson:"temperature_night_schedule,omitempty" json:"temperature_night_schedule,omitempty"`
	HumidityDaySchedule      map[string]ScheduleConfig `bson:"humidity_day_schedule,omitempty" json:"humidity_day_schedule,omitempty"`
	HumidityNightSchedule    map[string]ScheduleConfig `bson:"humidity_night_schedule,omitempty" json:"humidity_night_schedule,omitempty"`
	CO2DaySchedule           map[string]ScheduleConfig `bson:"co2_day_schedule,omitempty" json:"co2_day_schedule,omitempty"`
	CO2NightSchedule         map[string]ScheduleConfig `bson:"co2_night_schedule,omitempty" json:"co2_night_schedule,omitempty"`
	LightIntensitySchedule   map[string]ScheduleConfig `bson:"light_intensity_schedule,omitempty" json:"light_intensity_schedule,omitempty"`
}

// ScheduleConfig represents a schedule configuration for various parameters
type ScheduleConfig struct {
	EntityID string          `bson:"entity_id" json:"entity_id"`
	Schedule map[int]float64 `bson:"schedule" json:"schedule"`
}

// ScheduleItem represents a schedule item for an experiment
type ScheduleItem struct {
	PhaseIndex     int    `bson:"phase_index" json:"phase_index"`
	StartDate      string `bson:"start_date" json:"start_date"`
	EndDate        string `bson:"end_date" json:"end_date"`
	StartTimestamp int64  `bson:"start_timestamp" json:"start_timestamp"`
	EndTimestamp   int64  `bson:"end_timestamp" json:"end_timestamp"`
}
