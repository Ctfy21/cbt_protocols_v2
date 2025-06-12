# Local API v2

A Go-based web service that integrates climate chambers with a backend system and Home Assistant for automated experiment execution.

## Overview

The Local API v2 service acts as a bridge between:
- Climate chamber hardware controlled by Home Assistant
- A central backend system managing experiments
- MongoDB for local data persistence

## Key Features

### 1. Chamber Registration
- Automatically registers chamber with backend system
- Sends chamber information, local IP, and discovered entities
- Maintains regular heartbeat (every 30 seconds) to indicate chamber is online

### 2. Entity Discovery
- Automatically discovers Home Assistant input_numbers using substring matching
- Categorizes entities into:
  - **Climate Controls**: Temperature, humidity, CO2 (day/night values)
  - **Light Controls**: Lamp entities with intensity settings
  - **Watering Zones**: Start time, period, pause, and duration settings
  - **Day/Night Schedule**: Start time and duration controls

### 3. Experiment Synchronization
- Periodically syncs experiments from backend
- Stores experiments locally in MongoDB
- Identifies active experiments based on schedules

### 4. Experiment Execution
- Executes active experiment phases automatically
- Applies climate settings based on day/night schedule
- Controls lamp intensities per phase configuration
- Updates Home Assistant entities in real-time

## Project Structure

```
local_api_v2/
├── cmd/
│   └── api/           # Application entry point
├── internal/
│   ├── config/        # Configuration management
│   ├── database/      # MongoDB connection and operations
│   ├── handlers/      # HTTP request handlers
│   ├── models/        # Data models (Chamber, Experiment, etc.)
│   └── services/      # Business logic services
│       ├── discovery.go    # Home Assistant entity discovery
│       ├── registration.go # Backend registration and heartbeat
│       ├── sync.go        # Experiment synchronization
│       └── executor.go    # Experiment execution logic
├── pkg/
│   ├── homeassistant/ # Home Assistant API client
│   └── utils/         # Utility functions
├── go.mod             # Go module definition
├── go.sum             # Dependency checksums
└── env.example        # Example environment configuration
```

## Models

### Chamber
```go
type Chamber struct {
    ID                 ObjectID
    Name               string
    LocalIP            string
    BackendID          string  // ID assigned by backend
    Status             string
    DiscoveredEntities DiscoveredEntities
    CreatedAt          time.Time
    UpdatedAt          time.Time
}
```

### Experiment
```go
type Experiment struct {
    ID               ObjectID
    BackendID        string
    Title            string
    Description      string
    Status           string  // "active", "inactive", etc.
    ChamberID        string
    Phases           []Phase
    Schedule         []ScheduleItem
    ActivePhaseIndex *int
}
```

### Phase
```go
type Phase struct {
    Title          string
    Description    string
    DurationDays   int
    InputNumbers   map[string]*PhaseInputNumber  // Climate controls
    LightIntensity []LightIntensity             // Lamp settings
}
```

## Configuration

Create a `.env` file based on `env.example`:

```env
# Home Assistant Configuration
HA_URL=http://homeassistant.local:8123
HA_TOKEN=your_long_lived_access_token

# MongoDB Configuration
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=local_api_v2

# Backend Configuration
BACKEND_URL=http://backend.example.com
BACKEND_API_KEY=your_backend_api_key

# Chamber Configuration
CHAMBER_NAME=Growth Chamber 1
CHAMBER_DESCRIPTION=Climate controlled growth chamber
LOCAL_IP=192.168.1.100

# Server Configuration
PORT=8080
GIN_MODE=release

# Sync Configuration
SYNC_INTERVAL=5m
```

## Entity Discovery

The service uses substring matching to discover Home Assistant entities:

### Climate Controls
- `hours_day` → `day_start`
- `hours_night` → `day_duration`
- `temp_day` → `temperature_day`
- `temp_night` → `temperature_night`
- `humidity_day` → `humidity_day`
- `humidity_night` → `humidity_night`
- `co2_day` → `co2_day`
- `co2_night` → `co2_night`

### Lamp Controls
Detects any entity containing "lamp" or "light" in the name.

### Watering Zones
- `watering_zone_X_hour` → Zone start time
- `watering_zone_X_period` → Days between watering
- `watering_zone_X_pause` → Pause between pulses
- `watering_zone_X_duration` → Duration in seconds

## API Endpoints

### Health Check
```
GET /api/v1/health
```
Returns service health status and chamber information.

### Chamber Information
```
GET /api/v1/chamber
PUT /api/v1/chamber
```
Get or update chamber information.

### Experiments
```
GET /api/v1/experiments
GET /api/v1/experiments/:id
```
List all experiments or get specific experiment details.

### Sync Status
```
GET /api/v1/sync/status
POST /api/v1/sync/trigger
```
Check sync status or manually trigger synchronization.

### Registration Status
```
GET /api/v1/registration/status
```
Check registration and heartbeat status.

## Services

### DiscoveryService
- Discovers Home Assistant entities
- Categorizes entities by type
- Maps entities to standardized names

### RegistrationService
- Registers chamber with backend
- Sends periodic heartbeats
- Updates chamber status

### SyncService
- Fetches experiments from backend
- Stores experiments in MongoDB
- Identifies active experiments

### ExecutorService
- Monitors active experiments
- Applies phase settings to Home Assistant
- Handles day/night scheduling
- Updates experiment status

## Development

### Prerequisites
- Go 1.21 or higher
- MongoDB 4.4 or higher
- Home Assistant with Long-Lived Access Token
- Backend API access

### Setup
1. Clone the repository
2. Copy `env.example` to `.env` and configure
3. Install dependencies: `go mod download`
4. Run the service: `go run main.go`

### Building
```bash
go build -o local_api_v2
```

### Testing
```bash
go test ./...
```

## Docker Support

Build the Docker image:
```bash
docker build -t local_api_v2 .
```

Run with Docker:
```bash
docker run -d \
  --name local_api_v2 \
  --env-file .env \
  -p 8080:8080 \
  local_api_v2
```

## Troubleshooting

### Entity Discovery Issues
- Ensure Home Assistant token has proper permissions
- Check entity naming follows expected patterns
- Review discovery logs for substring matching details

### Sync Problems
- Verify backend URL and API key
- Check network connectivity
- Review sync interval settings

### Execution Issues
- Confirm experiments have valid schedules
- Check phase input number mappings
- Verify Home Assistant entity accessibility

## Future Enhancements

1. **WebSocket Support**: Real-time updates from Home Assistant
2. **Metrics & Monitoring**: Prometheus metrics endpoint
3. **Web UI**: Local dashboard for chamber monitoring
4. **Backup & Restore**: Automated data backup functionality
5. **Multi-Chamber Support**: Handle multiple chambers per instance 