package models

import (
	"time"

	"local_api_v2/pkg/ntp"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Experiment represents an experiment synchronized from backend
type Experiment struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	BackendID        primitive.ObjectID `bson:"backend_id" json:"backend_id"`
	Title            string             `bson:"title" json:"title"`
	Description      string             `bson:"description" json:"description"`
	Status           string             `bson:"status" json:"status"` // draft, active, paused, completed, archived
	ChamberID        primitive.ObjectID `bson:"chamber_id" json:"chamber_id"`
	ChamberName      string             `bson:"chamber_name" json:"chamber_name"`
	Phases           []Phase            `bson:"phases" json:"phases"`
	StartDate        *time.Time         `bson:"start_date,omitempty" json:"start_date,omitempty"`
	EndDate          *time.Time         `bson:"end_date,omitempty" json:"end_date,omitempty"`
	TotalDuration    int                `bson:"total_duration" json:"total_duration"`
	Schedule         []ScheduleItem     `bson:"schedule" json:"schedule"`
	ActivePhaseIndex *int               `bson:"active_phase_index,omitempty" json:"active_phase_index,omitempty"`
	CreatedAt        time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time          `bson:"updated_at" json:"updated_at"`
	SyncedAt         time.Time          `bson:"synced_at" json:"synced_at"`
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

// PhaseInputNumber represents input number configuration for a phase
type PhaseInputNumber struct {
	ID        string    `bson:"id" json:"id"`
	Name      string    `bson:"name" json:"name"`
	Type      string    `bson:"type" json:"type"`
	Value     float64   `bson:"value" json:"value"`
	Min       float64   `bson:"min" json:"min"`
	Max       float64   `bson:"max" json:"max"`
	Step      float64   `bson:"step" json:"step"`
	Unit      string    `bson:"unit" json:"unit"`
	Enabled   bool      `bson:"enabled" json:"enabled"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
	EntityID  string    `bson:"entity_id,omitempty" json:"entity_id,omitempty"` // Mapped HA entity ID
}

// LightIntensity represents light intensity setting for a specific lamp in a phase
type LightIntensity struct {
	LampName  string  `bson:"lamp_name" json:"lamp_name"`
	EntityID  string  `bson:"entity_id" json:"entity_id"`
	Intensity float64 `bson:"intensity" json:"intensity"` // 0-100%
	Enabled   bool    `bson:"enabled" json:"enabled"`
}

// ScheduleItem represents a scheduled phase with dates
type ScheduleItem struct {
	PhaseIndex     int   `bson:"phase_index" json:"phase_index"`
	StartTimestamp int64 `bson:"start_timestamp" json:"start_timestamp"`
	EndTimestamp   int64 `bson:"end_timestamp" json:"end_timestamp"`
}

// DayAndTimestamp represents a day number and its corresponding timestamp
type DayAndTimestamp struct {
	Day       int   `json:"day"`
	Timestamp int64 `json:"timestamp"`
}

// ExperimentStatus constants
const (
	StatusDraft     = "draft"
	StatusActive    = "active"
	StatusPaused    = "paused"
	StatusCompleted = "completed"
	StatusArchived  = "archived"
)

// GetCurrentPhase returns the current active phase based on the current date
func (e *Experiment) GetCurrentPhase(ntpService *ntp.TimeService) (*Phase, int, error) {
	if e.Status != StatusActive || e.StartDate == nil {
		return nil, -1, nil
	}

	now := ntpService.NowInMoscow()

	for _, scheduleItem := range e.Schedule {
		// Use timestamps for comparison
		startTime := time.Unix(scheduleItem.StartTimestamp, 0)
		endTime := time.Unix(scheduleItem.EndTimestamp, 0)

		if now.After(startTime) && now.Before(endTime) {
			if scheduleItem.PhaseIndex < len(e.Phases) {
				return &e.Phases[scheduleItem.PhaseIndex], scheduleItem.PhaseIndex, nil
			}
		}
	}

	return nil, -1, nil
}
