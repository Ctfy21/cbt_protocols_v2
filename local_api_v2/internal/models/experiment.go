package models

import (
	"time"

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

// Phase represents a phase in an experiment
type Phase struct {
	ID             string                       `bson:"id" json:"id"`
	Title          string                       `bson:"title" json:"title"`
	Description    string                       `bson:"description" json:"description"`
	Duration       int                          `bson:"duration" json:"duration"` // duration in days
	InputNumbers   map[string]*PhaseInputNumber `bson:"input_numbers" json:"input_numbers"`
	LightIntensity map[string]*LightIntensity   `bson:"light_intensity" json:"light_intensity"`
	LastExecuted   *time.Time                   `bson:"last_executed,omitempty" json:"last_executed,omitempty"`
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
	PhaseIndex     int    `bson:"phase_index" json:"phase_index"`
	StartDate      string `bson:"start_date" json:"start_date"`
	EndDate        string `bson:"end_date" json:"end_date"`
	StartTimestamp int64  `bson:"start_timestamp" json:"start_timestamp"`
	EndTimestamp   int64  `bson:"end_timestamp" json:"end_timestamp"`
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
func (e *Experiment) GetCurrentPhase() (*Phase, int, error) {
	if e.Status != StatusActive || e.StartDate == nil {
		return nil, -1, nil
	}

	now := time.Now()

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

// IsScheduledTime checks if current time matches the phase schedule
func (e *Experiment) IsScheduledTime(phaseType string) bool {
	phase, _, err := e.GetCurrentPhase()
	if err != nil || phase == nil {
		return false
	}

	now := time.Now()
	currentHour := float64(now.Hour()) + float64(now.Minute())/60

	// Check day/night schedule
	if startTime, ok := phase.InputNumbers["start_time"]; ok && startTime.Enabled {
		if duration, ok := phase.InputNumbers["duration_hours"]; ok && duration.Enabled {
			dayStart := startTime.Value
			dayDuration := duration.Value
			dayEnd := dayStart + dayDuration

			if dayEnd <= 24 {
				// Normal day schedule (doesn't cross midnight)
				isDayTime := currentHour >= dayStart && currentHour < dayEnd
				if phaseType == "day" {
					return isDayTime
				}
				return !isDayTime // night time
			} else {
				// Day schedule crosses midnight
				isDayTime := currentHour >= dayStart || currentHour < (dayEnd-24)
				if phaseType == "day" {
					return isDayTime
				}
				return !isDayTime
			}
		}
	}

	// Default: assume day time from 6:00 to 18:00
	isDayTime := currentHour >= 6 && currentHour < 18
	if phaseType == "day" {
		return isDayTime
	}
	return !isDayTime
}
