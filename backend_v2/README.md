# Backend v2

Central management backend for CBT Protocols climate chamber network. This backend works with `local_api_v2` instances running at each chamber location.

## Overview

Backend v2 provides:
- Chamber registration and heartbeat monitoring
- Experiment management and distribution
- RESTful API for local_api_v2 communication
- Web interface integration

## Architecture

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  local_api_v2   │     │  local_api_v2   │     │  local_api_v2   │
│   (Chamber 1)   │     │   (Chamber 2)   │     │   (Chamber 3)   │
└────────┬────────┘     └────────┬────────┘     └────────┬────────┘
         │                       │                       │
         │   Registration        │                       │
         │   Heartbeat          │                       │
         │   Sync Experiments   │                       │
         └───────────────────────┴───────────────────────┘
                                 │
                        ┌────────▼────────┐
                        │   Backend v2    │
                        │  (Central API)  │
                        └────────┬────────┘
                                 │
                        ┌────────▼────────┐
                        │    MongoDB      │
                        └─────────────────┘
```

## Features

### Chamber Management
- **Auto-registration**: Chambers register themselves with their capabilities
- **Heartbeat monitoring**: Automatic online/offline status tracking
- **Entity discovery**: Stores discovered Home Assistant entities from each chamber

### Experiment Management
- Create and manage experiments with multiple phases
- Assign experiments to specific chambers
- Schedule-based experiment activation
- Real-time sync with local chambers

### API Endpoints

#### Chamber Endpoints
- `POST /chambers` - Register/update chamber
- `POST /chambers/:id/heartbeat` - Update chamber heartbeat
- `GET /chambers/:id` - Get chamber details
- `GET /chambers` - List all chambers

#### Experiment Endpoints
- `POST /experiments` - Create experiment
- `GET /experiments/:id` - Get experiment details
- `GET /experiments?chamber_id=:id` - List experiments (optionally by chamber)
- `PUT /experiments/:id` - Update experiment
- `DELETE /experiments/:id` - Delete experiment

#### Health Check
- `GET /health` - Service health status

## Installation

1. **Prerequisites**
   - Go 1.21 or higher
   - MongoDB 4.4 or higher

2. **Setup**
   ```bash
   cd backend_v2
   cp env.example .env
   # Edit .env with your configuration
   ```

3. **Install dependencies**
   ```bash
   go mod tidy
   ```

4. **Run the server**
   ```bash
   go run main.go
   ```

## Configuration

Create a `.env` file based on `env.example`:

```env
# MongoDB Configuration
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=cbt_protocols_v2

# Server Configuration
PORT=8081
GIN_MODE=release

# JWT Configuration (optional, for future use)
JWT_SECRET=your-secret-key-here
JWT_EXPIRATION=24h

# API Configuration
API_KEY=your-api-key-here

# Chamber Configuration
HEARTBEAT_TIMEOUT=60  # seconds - mark chamber offline after this
CLEANUP_INTERVAL=300  # seconds - how often to check chamber status
```

## API Authentication

If `API_KEY` is set in the configuration, all API requests (except health check) must include:

```
Authorization: Bearer your-api-key-here
```

## Data Models

### Chamber
```json
{
  "id": "ObjectId",
  "name": "Chamber 1",
  "location": "Local API v2",
  "ha_url": "http://homeassistant.local:8123",
  "local_ip": "192.168.1.100",
  "status": "online",
  "last_heartbeat": "2024-01-15T10:30:00Z",
  "input_numbers": [...],
  "lamps": [...],
  "watering_zones": [...],
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

### Experiment
```json
{
  "id": "ObjectId",
  "title": "Tomato Growth Experiment",
  "description": "Testing optimal conditions",
  "status": "active",
  "chamber_id": "ObjectId",
  "phases": [
    {
      "title": "Germination",
      "description": "Initial growth phase",
      "duration_days": 7,
      "input_numbers": {
        "temperature_day": {
          "entity_id": "input_number.temp_day",
          "value": 25
        }
      },
      "light_intensity": [
        {
          "entity_id": "input_number.lamp_1",
          "intensity": 80
        }
      ]
    }
  ],
  "schedule": [
    {
      "phase_index": 0,
      "start_date": "2024-01-15",
      "end_date": "2024-01-22",
      "start_timestamp": 1705276800,
      "end_timestamp": 1705881600
    }
  ],
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-15T00:00:00Z"
}
```

## Integration with local_api_v2

The `local_api_v2` instances:
1. Register with backend on startup
2. Send heartbeats every 30 seconds
3. Sync experiments every 60 seconds
4. Execute active experiments locally

## Development

### Project Structure
```
backend_v2/
├── main.go              # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── database/        # MongoDB connection
│   ├── handlers/        # HTTP request handlers
│   ├── models/          # Data models
│   ├── services/        # Business logic
│   └── middleware/      # HTTP middleware
└── pkg/
    └── utils/           # Utility functions
```

### Running Tests
```bash
go test ./...
```

### Building
```bash
go build -o backend_v2
```

## Docker Support

Build the Docker image:
```bash
docker build -t backend_v2 .
```

Run with Docker:
```bash
docker run -d \
  --name backend_v2 \
  --env-file .env \
  -p 8081:8081 \
  backend_v2
```

## Monitoring

The backend automatically:
- Marks chambers as offline if no heartbeat received within timeout
- Logs all chamber registrations and status changes
- Tracks experiment synchronization

## Troubleshooting

### Chamber Shows Offline
- Check network connectivity between chamber and backend
- Verify heartbeat timeout configuration
- Check local_api_v2 logs for errors

### Experiments Not Syncing
- Ensure chamber is registered and online
- Check experiment chamber_id matches
- Verify local_api_v2 sync interval

### MongoDB Connection Issues
- Verify MongoDB is running
- Check connection string in .env
- Ensure database user has proper permissions 