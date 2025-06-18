package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ExperimentStatus represents the status of an experiment
type ExperimentStatus string

const (
	ExperimentStatusActive    ExperimentStatus = "active"
	ExperimentStatusDraft     ExperimentStatus = "draft"
	ExperimentStatusCompleted ExperimentStatus = "completed"
	ExperimentStatusPaused    ExperimentStatus = "paused"
	ExperimentStatusArchived  ExperimentStatus = "archived"
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

type Phase struct {
	Title                    string                          `bson:"title" json:"title"`
	Description              string                          `bson:"description" json:"description"`
	LastExecuted             *time.Time                      `bson:"last_executed,omitempty" json:"last_executed,omitempty"`
	DurationDays             int                             `bson:"duration_days" json:"duration_days"`
	StartDay                 map[string]StartDayConfig       `bson:"start_day,omitempty" json:"start_day,omitempty"`
	WorkDaySchedule          map[string]ScheduleConfig       `bson:"work_day_schedule,omitempty" json:"work_day_schedule,omitempty"`
	TemperatureDaySchedule   map[string]ScheduleConfig       `bson:"temperature_day_schedule,omitempty" json:"temperature_day_schedule,omitempty"`
	TemperatureNightSchedule map[string]ScheduleConfig       `bson:"temperature_night_schedule,omitempty" json:"temperature_night_schedule,omitempty"`
	HumidityDaySchedule      map[string]ScheduleConfig       `bson:"humidity_day_schedule,omitempty" json:"humidity_day_schedule,omitempty"`
	HumidityNightSchedule    map[string]ScheduleConfig       `bson:"humidity_night_schedule,omitempty" json:"humidity_night_schedule,omitempty"`
	CO2DaySchedule           map[string]ScheduleConfig       `bson:"co2_day_schedule,omitempty" json:"co2_day_schedule,omitempty"`
	CO2NightSchedule         map[string]ScheduleConfig       `bson:"co2_night_schedule,omitempty" json:"co2_night_schedule,omitempty"`
	LightIntensitySchedule   map[string]ScheduleConfig       `bson:"light_intensity_schedule,omitempty" json:"light_intensity_schedule,omitempty"`
	WateringZones            map[string]WateringZoneSchedule `bson:"watering_zones,omitempty" json:"watering_zones,omitempty"`
}

// StartDayConfig represents the configuration for start day values
type StartDayConfig struct {
	EntityID string  `bson:"entity_id" json:"entity_id"`
	Value    float64 `bson:"value" json:"value"`
}

// WateringZoneSchedule represents the schedule configuration for watering zones
type WateringZoneSchedule struct {
	Name                 string          `bson:"name" json:"name"`
	StartTimeEntityID    string          `bson:"start_time_entity_id" json:"start_time_entity_id"`
	PeriodEntityID       string          `bson:"period_entity_id" json:"period_entity_id"`
	PauseBetweenEntityID string          `bson:"pause_between_entity_id" json:"pause_between_entity_id"`
	DurationEntityID     string          `bson:"duration_entity_id" json:"duration_entity_id"`
	StartTimeSchedule    map[int]float64 `bson:"start_time_schedule" json:"start_time_schedule"`
	PeriodSchedule       map[int]float64 `bson:"period_schedule" json:"period_schedule"`
	PauseBetweenSchedule map[int]float64 `bson:"pause_between_schedule" json:"pause_between_schedule"`
	DurationSchedule     map[int]float64 `bson:"duration_schedule" json:"duration_schedule"`
}

// ScheduleConfig represents a schedule configuration for various parameters
type ScheduleConfig struct {
	EntityID string          `bson:"entity_id" json:"entity_id"`
	Schedule map[int]float64 `bson:"schedule" json:"schedule"`
}

// ScheduleItem represents a schedule item for an experiment
type ScheduleItem struct {
	PhaseIndex     int   `bson:"phase_index" json:"phase_index"`
	StartTimestamp int64 `bson:"start_timestamp" json:"start_timestamp"`
	EndTimestamp   int64 `bson:"end_timestamp" json:"end_timestamp"`
}
