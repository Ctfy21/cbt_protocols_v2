// API Response types
export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: string;
  message?: string;
}

// User types
export interface User {
  id: string;
  username: string;
  name: string;
  role: 'admin' | 'user';
  is_active: boolean;
  created_at: string;
  updated_at: string;
  last_login?: string;
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface AuthResponse {
  user: User;
  token: string;
  refresh_token: string;
  expires_in: number;
}

// Chamber types
export interface Chamber {
  id: string;
  name: string;
  suffix: string; // Added suffix field
  location: string;
  ha_url: string;
  local_ip: string;
  status: 'online' | 'offline';
  last_heartbeat: string;
  discovery_completed: boolean;
  config?: ChamberConfig;
  created_at: string;
  updated_at: string;
}

export interface ChamberConfig {
  id?: string;
  chamber_id?: string;
  lamps: Record<string, InputNumber>; // Changed from array to map
  watering_zones: WateringZone[];
  unrecognised_entities: Record<string, InputNumber>; // Changed from array to map
  day_duration: Record<string, InputNumber>;
  day_start: Record<string, InputNumber>;
  temperature: {
    day: Record<string, InputNumber>;
    night: Record<string, InputNumber>;
  };
  humidity: {
    day: Record<string, InputNumber>;
    night: Record<string, InputNumber>;
  };
  co2: {
    day: Record<string, InputNumber>;
    night: Record<string, InputNumber>;
  };
  updated_at?: string;
  synced_at?: string;
}

export interface InputNumber {
  entity_id: string;
  name: string;
  type: string;
  min: number;
  max: number;
  step: number;
  value: number;
  unit: string;
}

export interface WateringZone {
  name: string;
  start_time_entity_id: Record<string, InputNumber>;
  period_entity_id: Record<string, InputNumber>;
  pause_between_entity_id: Record<string, InputNumber>;
  duration_entity_id: Record<string, InputNumber>;
}

// Removed Lamp interface as it's now part of InputNumber

export type ExperimentStatus = 'active' | 'draft' | 'completed' | 'paused' | 'archived'

// Experiment types
export interface Experiment {
  id: string;
  title: string;
  description: string;
  status: 'draft' | 'active' | 'paused' | 'completed' | 'archived';
  chamber_id: string;
  phases: Phase[];
  schedule: ScheduleItem[];
  active_phase_index?: number;
  created_at: string;
  updated_at: string;
}

export interface Phase {
  title: string;
  description: string;
  duration_days: number;
  start_day?: Record<string, StartDayConfig>;
  work_day_schedule?: Record<string, ScheduleConfig>;
  temperature_day_schedule?: Record<string, ScheduleConfig>;
  temperature_night_schedule?: Record<string, ScheduleConfig>;
  humidity_day_schedule?: Record<string, ScheduleConfig>;
  humidity_night_schedule?: Record<string, ScheduleConfig>;
  co2_day_schedule?: Record<string, ScheduleConfig>;
  co2_night_schedule?: Record<string, ScheduleConfig>;
  light_intensity_schedule?: Record<string, ScheduleConfig>;
  watering_zones?: Record<string, WateringZoneSchedule>;
  last_executed?: string;
}

export interface StartDayConfig {
  entity_id: string;
  value: number;
}

export interface ScheduleConfig {
  entity_id: string;
  schedule: Record<number, number>;
}

export interface WateringZoneSchedule {
  name: string;
  start_time_entity_id: string;
  period_entity_id: string;
  pause_between_entity_id: string;
  duration_entity_id: string;
  start_time_schedule: Record<number, number>;
  period_schedule: Record<number, number>;
  pause_between_schedule: Record<number, number>;
  duration_schedule: Record<number, number>;
}

export interface ScheduleItem {
  phase_index: number;
  start_timestamp: number;
  end_timestamp: number;
}

// API Token types
export interface APIToken {
  id: string;
  name: string;
  token: string;
  type: 'service' | 'personal';
  service_name?: string;
  permissions: string[];
  is_active: boolean;
  expires_at?: string;
  last_used_at?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateAPITokenRequest {
  name: string;
  type: 'service' | 'personal';
  service_name?: string;
  permissions: string[];
  expires_at?: string;
}

// User Chamber Access types
export interface UserChamberAccess {
  user: User;
  chambers: Chamber[];
}

// Form Data types
export interface ExperimentFormData {
  title: string;
  description: string;
  status: ExperimentStatus;
  chamber_id: string;
  phases: Phase[];
  schedule?: ScheduleItem[];
}
