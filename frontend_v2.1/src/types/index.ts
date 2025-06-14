// Chamber related types
export type ChamberStatus = 'online' | 'offline'

export interface Chamber {
  id: string
  name: string
  location: string
  ha_url: string
  local_ip: string
  status: ChamberStatus
  last_heartbeat: string
  input_numbers: InputNumber[]
  lamps: Lamp[]
  watering_zones: WateringZone[]
  created_at: string
  updated_at: string
}

export interface InputNumber {
  entity_id: string
  name: string
  friendly_name: string
  type: string
  min: number
  max: number
  step: number
  value: number
  current_value: number
  unit: string
}

export interface Lamp {
  name: string
  entity_id: string
  friendly_name: string
  intensity_min: number
  intensity_max: number
  current_value: number
}

export interface WateringZone {
  name: string
  start_time_entity_id: string
  period_entity_id: string
  pause_between_entity_id: string
  duration_entity_id: string
}

// Experiment related types
export type ExperimentStatus = 'active' | 'inactive' | 'draft' | 'completed' | 'paused'

export interface Experiment {
  id: string
  title: string
  description: string
  status: ExperimentStatus
  chamber_id: string
  phases: Phase[]
  schedule: ScheduleItem[]
  active_phase_index?: number
  created_at: string
  updated_at: string
  // Frontend specific fields
  total_duration?: number
  start_date?: string
  end_date?: string
}

export interface Phase {
  id?: number
  title: string
  description: string
  duration_days: number
  start_day?: Record<string, { entity_id: string; value: number }>
  work_day_schedule?: Record<string, { entity_id: string; schedule: Record<number, number> }>
  temperature_day_schedule?: Record<string, { entity_id: string; schedule: Record<number, number> }>
  temperature_night_schedule?: Record<string, { entity_id: string; schedule: Record<number, number> }>
  humidity_day_schedule?: Record<string, { entity_id: string; schedule: Record<number, number> }>
  humidity_night_schedule?: Record<string, { entity_id: string; schedule: Record<number, number> }>
  co2_day_schedule?: Record<string, { entity_id: string; schedule: Record<number, number> }>
  co2_night_schedule?: Record<string, { entity_id: string; schedule: Record<number, number> }>
  light_intensity_schedule?: Record<string, { entity_id: string; schedule: Record<number, number> }>
  watering_zones?: Record<string, WateringZoneSchedule>
}

export interface WateringZoneSchedule {
  name: string
  start_time_entity_id: string
  period_entity_id: string
  pause_between_entity_id: string
  duration_entity_id: string
  start_time_schedule: Record<number, number>
  period_schedule: Record<number, number>
  pause_between_schedule: Record<number, number>
  duration_schedule: Record<number, number>
}

export interface ScheduleItem {
  phase_index: number
  start_timestamp: number
  end_timestamp: number
}

// API Response types
export interface ApiResponse<T> {
  success: boolean
  data?: T
  error?: string
  message?: string
}

// Form types
export interface ExperimentFormData {
  title: string
  description: string
  status: ExperimentStatus
  chamber_id: string
  phases: Phase[]
  start_date: string
}