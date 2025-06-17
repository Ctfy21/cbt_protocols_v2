#!/bin/bash

# Local API v2 Deployment Script with Updated Architecture
# This script sets up the complete environment for Local API v2 with simplified chamber management

set -e  # Exit on error

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration variables
GO_VERSION="1.23.2"
MONGODB_VERSION="6.0"
PROJECT_DIR="/opt/local_api_v2"
SERVICE_USER="local_api"
SERVICE_NAME="local_api_v2"
MONGODB_DATA_DIR="/var/lib/mongodb"
MONGODB_LOG_DIR="/var/log/mongodb"

# Function to print colored output
print_status() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[✓]${NC} $1"
}

print_error() {
    echo -e "${RED}[✗]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[!]${NC} $1"
}

# Check if running as root
check_root() {
    if [[ $EUID -ne 0 ]]; then
        print_error "This script must be run as root"
        exit 1
    fi
}

# Detect Ubuntu version
detect_ubuntu_version() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        OS=$ID
        VER=$VERSION_ID
        CODENAME=$VERSION_CODENAME
        print_status "Detected $OS $VER ($CODENAME)"
    else
        print_error "Cannot detect OS version"
        exit 1
    fi
}

# Update system packages
update_system() {
    print_status "Updating system packages..."
    apt update && apt upgrade -y
    apt install -y curl wget git build-essential software-properties-common apt-transport-https ca-certificates gnupg lsb-release
    print_success "System packages updated"
}

# Install Go
install_go() {
    print_status "Installing Go ${GO_VERSION}..."
    
    # Check if Go is already installed
    if command -v go &> /dev/null; then
        CURRENT_GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
        print_warning "Go ${CURRENT_GO_VERSION} is already installed"
        
        # Check if it's the correct version
        if [ "$CURRENT_GO_VERSION" == "$GO_VERSION" ]; then
            print_success "Go ${GO_VERSION} is already installed"
            return
        else
            print_status "Updating Go to version ${GO_VERSION}..."
            rm -rf /usr/local/go
        fi
    fi
    
    # Download and install Go
    wget -q "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -O /tmp/go.tar.gz
    tar -C /usr/local -xzf /tmp/go.tar.gz
    rm /tmp/go.tar.gz
    
    # Add Go to PATH for all users
    if ! grep -q "/usr/local/go/bin" /etc/profile; then
        echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
        echo 'export GOPATH=$HOME/go' >> /etc/profile
        echo 'export PATH=$PATH:$GOPATH/bin' >> /etc/profile
    fi
    
    # Create Go symlinks for system-wide access
    ln -sf /usr/local/go/bin/go /usr/bin/go
    ln -sf /usr/local/go/bin/gofmt /usr/bin/gofmt
    
    # Source the profile for current session
    export PATH=$PATH:/usr/local/go/bin
    
    # Verify installation
    if command -v go &> /dev/null; then
        print_success "Go ${GO_VERSION} installed successfully"
        go version
    else
        print_error "Failed to install Go"
        exit 1
    fi
}

# Install MongoDB natively
install_mongodb() {
    print_status "Installing MongoDB ${MONGODB_VERSION}..."
    
    # Check if MongoDB is already installed
    if command -v mongod &> /dev/null; then
        CURRENT_MONGO_VERSION=$(mongod --version | grep "db version" | awk '{print $3}' | cut -d'v' -f2)
        print_warning "MongoDB ${CURRENT_MONGO_VERSION} is already installed"
        
        # Check if MongoDB service is running
        if systemctl is-active --quiet mongod; then
            print_success "MongoDB is already running"
            return
        fi
    fi
    
    # Import MongoDB public GPG key
    print_status "Adding MongoDB repository..."
    wget -qO - https://www.mongodb.org/static/pgp/server-${MONGODB_VERSION}.asc | apt-key add -
    
    # Add MongoDB repository based on Ubuntu version
    echo "deb [ arch=amd64,arm64 ] https://repo.mongodb.org/apt/ubuntu $(lsb_release -cs)/mongodb-org/${MONGODB_VERSION} multiverse" | tee /etc/apt/sources.list.d/mongodb-org-${MONGODB_VERSION}.list
    
    # Update package database
    apt update
    
    # Install MongoDB packages
    apt install -y mongodb-org
    
    # Create MongoDB directories if they don't exist
    mkdir -p $MONGODB_DATA_DIR
    mkdir -p $MONGODB_LOG_DIR
    chown -R mongodb:mongodb $MONGODB_DATA_DIR
    chown -R mongodb:mongodb $MONGODB_LOG_DIR
    
    # Create MongoDB configuration file
    create_mongodb_config
    
    # Enable and start MongoDB service
    systemctl daemon-reload
    systemctl enable mongod
    systemctl start mongod
    
    # Wait for MongoDB to start
    print_status "Waiting for MongoDB to start..."
    for i in {1..30}; do
        if mongosh --eval "db.adminCommand('ping')" &>/dev/null; then
            print_success "MongoDB ${MONGODB_VERSION} installed and running"
            return
        fi
        sleep 1
    done
    
    print_error "MongoDB failed to start"
    systemctl status mongod
    exit 1
}

# Create MongoDB configuration
create_mongodb_config() {
    print_status "Creating MongoDB configuration..."
    
    cat > /etc/mongod.conf << 'EOF'
# MongoDB configuration file

# Where to store data
storage:
  dbPath: /var/lib/mongodb
  journal:
    enabled: true
  engine: wiredTiger

# Where to write logging data
systemLog:
  destination: file
  logAppend: true
  path: /var/log/mongodb/mongod.log

# Network interfaces
net:
  port: 27017
  bindIp: 127.0.0.1

# Process management
processManagement:
  timeZoneInfo: /usr/share/zoneinfo

# Security
security:
  authorization: disabled

# Operation profiling
operationProfiling:
  mode: off

setParameter:
  enableLocalhostAuthBypass: true
EOF

    print_success "MongoDB configuration created"
}

# Create MongoDB database and user (optional)
setup_mongodb_database() {
    print_status "Setting up MongoDB database..."
    
    # Create database initialization script
    cat > /tmp/init_mongo.js << 'EOF'
// Switch to local_api_v2 database
use local_api_v2;

// Create collections
db.createCollection("chambers");
db.createCollection("experiments");

// Create indexes
db.chambers.createIndex({ "suffix": 1 });
db.chambers.createIndex({ "backend_id": 1 });
db.experiments.createIndex({ "backend_id": 1 });
db.experiments.createIndex({ "status": 1 });
db.experiments.createIndex({ "chamber_id": 1 });

print("Database 'local_api_v2' initialized successfully");
EOF

    # Execute initialization script
    mongosh < /tmp/init_mongo.js
    rm -f /tmp/init_mongo.js
    
    print_success "MongoDB database initialized"
}

# Create service user
create_service_user() {
    print_status "Creating service user..."
    
    if id "$SERVICE_USER" &>/dev/null; then
        print_warning "User $SERVICE_USER already exists"
    else
        useradd -r -s /bin/false -m -d /home/$SERVICE_USER $SERVICE_USER
        print_success "Service user created"
    fi
}

# Create project structure
setup_project() {
    print_status "Setting up project directory..."
    
    # Create project directory
    mkdir -p $PROJECT_DIR
    cd $PROJECT_DIR
    
    # Create project structure
    mkdir -p {internal/{config,database,models,services},pkg/{homeassistant,ntp}}
    
    # Create logs directory
    mkdir -p $PROJECT_DIR/logs
    
    print_success "Project directory created"
}

# Create environment file
create_env_file() {
    print_status "Creating environment configuration file..."
    
    cat > $PROJECT_DIR/.env.example << 'EOF'
# Home Assistant Configuration
HA_URL=http://homeassistant.local:8123
HA_TOKEN=your_long_lived_access_token_here

# MongoDB Configuration (Local installation)
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=local_api_v2

# Backend Configuration
BACKEND_URL=http://backend.example.com/api
BACKEND_API_KEY=your_backend_api_key_here

# Chamber Configuration
CHAMBER_NAME=Climate Chamber
CHAMBER_SUFFIXES=galo,sb4,oreol,sb1,room1,room2,room3
LOCAL_IP=auto

# Server Configuration
PORT=8090
GIN_MODE=release

# Heartbeat Configuration
HEARTBEAT_INTERVAL=30

# NTP Configuration
NTP_ENABLED=true
NTP_SERVERS=ru.pool.ntp.org,europe.pool.ntp.org,pool.ntp.org
NTP_SYNC_INTERVAL=5m
NTP_TIMEOUT=5s

# Logging
LOG_LEVEL=info
EOF

    # Copy example to actual .env if it doesn't exist
    if [ ! -f "$PROJECT_DIR/.env" ]; then
        cp $PROJECT_DIR/.env.example $PROJECT_DIR/.env
        print_warning "Created .env file from template. Please update with your actual values!"
    fi
    
    print_success "Environment configuration created"
}

# Create systemd service
create_systemd_service() {
    print_status "Creating systemd service..."
    
    cat > /etc/systemd/system/${SERVICE_NAME}.service << EOF
[Unit]
Description=Local API v2 Service
After=network.target mongod.service
Requires=mongod.service

[Service]
Type=simple
User=$SERVICE_USER
Group=$SERVICE_USER
WorkingDirectory=$PROJECT_DIR
Environment="PATH=/usr/local/go/bin:/usr/bin:/bin"
ExecStartPre=/bin/sleep 5
ExecStart=$PROJECT_DIR/local_api_v2
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=${SERVICE_NAME}

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=$PROJECT_DIR

# Environment
EnvironmentFile=$PROJECT_DIR/.env

[Install]
WantedBy=multi-user.target
EOF

    # Set permissions
    chown -R $SERVICE_USER:$SERVICE_USER $PROJECT_DIR
    
    # Reload systemd
    systemctl daemon-reload
    
    print_success "Systemd service created"
}

# Create helper scripts
create_helper_scripts() {
    print_status "Creating helper scripts..."
    
    # Create logs script
    cat > $PROJECT_DIR/logs.sh << 'EOF'
#!/bin/bash
echo "=== Local API v2 Logs ==="
journalctl -u local_api_v2 -f
EOF
    
    # Create restart script
    cat > $PROJECT_DIR/restart.sh << 'EOF'
#!/bin/bash
echo "Restarting Local API v2..."
systemctl restart local_api_v2
sleep 2
systemctl status local_api_v2
EOF
    
    # Create update script
    cat > $PROJECT_DIR/update.sh << 'EOF'
#!/bin/bash
cd /opt/local_api_v2
echo "Stopping service..."
systemctl stop local_api_v2
echo "Building application..."
sudo -u local_api go build -o local_api_v2 main.go
echo "Starting service..."
systemctl start local_api_v2
sleep 2
systemctl status local_api_v2
EOF
    
    # Create MongoDB check script
    cat > $PROJECT_DIR/check_mongo.sh << 'EOF'
#!/bin/bash
echo "=== MongoDB Status ==="
systemctl status mongod --no-pager
echo ""
echo "=== MongoDB Connection Test ==="
mongosh --eval "db.adminCommand('ping')"
echo ""
echo "=== Database Info ==="
mongosh local_api_v2 --eval "db.stats()"
echo ""
echo "=== Chambers Collection ==="
mongosh local_api_v2 --eval "db.chambers.countDocuments()"
EOF
    
    # Create backup script
    cat > $PROJECT_DIR/backup.sh << 'EOF'
#!/bin/bash
BACKUP_DIR="/opt/local_api_v2/backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
mkdir -p $BACKUP_DIR
echo "Creating backup..."
mongodump --db local_api_v2 --out $BACKUP_DIR/backup_$TIMESTAMP
echo "Backup created at: $BACKUP_DIR/backup_$TIMESTAMP"
# Keep only last 7 backups
find $BACKUP_DIR -type d -name "backup_*" -mtime +7 -exec rm -rf {} \;
EOF

    # Create status check script
    cat > $PROJECT_DIR/check_status.sh << 'EOF'
#!/bin/bash
echo "=== Service Status ==="
systemctl status local_api_v2 --no-pager
echo ""
echo "=== API Health Check ==="
curl -s http://localhost:8090/api/v1/health | python3 -m json.tool
echo ""
echo "=== Chambers Status ==="
curl -s http://localhost:8090/api/v1/chambers | python3 -m json.tool
echo ""
echo "=== Sync Status ==="
curl -s http://localhost:8090/api/v1/sync/status | python3 -m json.tool
EOF
    
    chmod +x $PROJECT_DIR/*.sh
    
    print_success "Helper scripts created"
}

# Configure firewall (optional)
configure_firewall() {
    print_status "Configuring firewall..."
    
    # Check if ufw is installed
    if ! command -v ufw &> /dev/null; then
        print_warning "UFW not installed. Skipping firewall configuration."
        return
    fi
    
    # Configure UFW rules
    ufw allow 22/tcp comment "SSH"
    ufw allow 8090/tcp comment "Local API v2"
    
    # MongoDB should only be accessible locally
    ufw deny 27017/tcp comment "Block external MongoDB access"
    
    print_success "Firewall configured (remember to enable with: ufw enable)"
}

# Create logrotate configuration
create_logrotate_config() {
    print_status "Creating log rotation configuration..."
    
    cat > /etc/logrotate.d/local_api_v2 << EOF
$PROJECT_DIR/logs/*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    create 0644 $SERVICE_USER $SERVICE_USER
    postrotate
        systemctl reload $SERVICE_NAME > /dev/null 2>&1 || true
    endscript
}
EOF
    
    print_success "Log rotation configured"
}

# Create deployment instructions
create_deployment_instructions() {
    cat > $PROJECT_DIR/DEPLOY_INSTRUCTIONS.md << 'EOF'
# Local API v2 Deployment Instructions

## Quick Start

1. **Clone your repository** (if using Git):
   ```bash
   git clone <your-repository-url> /tmp/local_api_v2_src
   cp -r /tmp/local_api_v2_src/* /opt/local_api_v2/
   rm -rf /tmp/local_api_v2_src
   ```

2. **Or copy files manually**:
   - Copy `main.go` to `/opt/local_api_v2/`
   - Copy `internal/*` to `/opt/local_api_v2/internal/`
   - Copy `pkg/*` to `/opt/local_api_v2/pkg/`

3. **Build the application**:
   ```bash
   cd /opt/local_api_v2
   sudo -u local_api go mod init local_api_v2
   sudo -u local_api go mod tidy
   sudo -u local_api go build -o local_api_v2 main.go
   ```

4. **Configure environment**:
   ```bash
   sudo nano /opt/local_api_v2/.env
   ```
   Update:
   - `HA_URL` - Your Home Assistant URL
   - `HA_TOKEN` - Your long-lived access token
   - `BACKEND_URL` - Your backend API URL
   - `BACKEND_API_KEY` - Your backend API key
   - `CHAMBER_NAME` - Your chamber name
   - `CHAMBER_SUFFIXES` - Your chamber suffixes

5. **Start the service**:
   ```bash
   sudo systemctl enable local_api_v2
   sudo systemctl start local_api_v2
   ```

6. **Check status**:
   ```bash
   sudo /opt/local_api_v2/check_status.sh
   ```

## Chamber Suffixes Configuration

The system supports multiple chambers identified by suffixes:
- `default` - Entities without suffix
- `galo`, `sb4`, `oreol`, `sb1` - Specific room chambers
- `room1`, `room2`, `room3` - Generic room chambers

Update `CHAMBER_SUFFIXES` in `.env` to match your setup.

## Troubleshooting

- **Check logs**: `sudo /opt/local_api_v2/logs.sh`
- **MongoDB issues**: `sudo /opt/local_api_v2/check_mongo.sh`
- **Restart service**: `sudo /opt/local_api_v2/restart.sh`
- **Update code**: `sudo /opt/local_api_v2/update.sh`

## API Endpoints

- Health: http://localhost:8090/api/v1/health
- Chambers: http://localhost:8090/api/v1/chambers
- Sync Status: http://localhost:8090/api/v1/sync/status
- Time: http://localhost:8090/api/v1/time
EOF
    
    print_success "Deployment instructions created"
}

# Verify installation
verify_installation() {
    print_status "Verifying installation..."
    
    local errors=0
    
    # Check Go installation
    if command -v go &> /dev/null; then
        print_success "Go is installed: $(go version)"
    else
        print_error "Go is not installed"
        ((errors++))
    fi
    
    # Check MongoDB installation
    if systemctl is-active --quiet mongod; then
        print_success "MongoDB is running"
    else
        print_error "MongoDB is not running"
        ((errors++))
    fi
    
    # Check MongoDB connectivity
    if mongosh --eval "db.adminCommand('ping')" &>/dev/null; then
        print_success "MongoDB is accessible"
    else
        print_error "Cannot connect to MongoDB"
        ((errors++))
    fi
    
    # Check project directory
    if [ -d "$PROJECT_DIR" ]; then
        print_success "Project directory exists"
    else
        print_error "Project directory not found"
        ((errors++))
    fi
    
    # Check service user
    if id "$SERVICE_USER" &>/dev/null; then
        print_success "Service user exists"
    else
        print_error "Service user not found"
        ((errors++))
    fi
    
    if [ $errors -eq 0 ]; then
        print_success "All components installed successfully!"
        return 0
    else
        print_error "Installation verification failed with $errors errors"
        return 1
    fi
}

# Display final instructions
display_instructions() {
    echo ""
    echo -e "${GREEN}============================================${NC}"
    echo -e "${GREEN}Local API v2 Installation Complete!${NC}"
    echo -e "${GREEN}============================================${NC}"
    echo ""
    echo -e "${YELLOW}Important next steps:${NC}"
    echo ""
    echo "1. Copy your source files to $PROJECT_DIR"
    echo "   See DEPLOY_INSTRUCTIONS.md for details"
    echo ""
    echo "2. Configure environment variables:"
    echo "   sudo nano $PROJECT_DIR/.env"
    echo ""
    echo "3. Build and start the service:"
    echo "   cd $PROJECT_DIR"
    echo "   sudo -u $SERVICE_USER go mod init local_api_v2"
    echo "   sudo -u $SERVICE_USER go mod tidy"
    echo "   sudo -u $SERVICE_USER go build -o local_api_v2 main.go"
    echo "   sudo systemctl enable ${SERVICE_NAME}"
    echo "   sudo systemctl start ${SERVICE_NAME}"
    echo ""
    echo -e "${BLUE}Helper scripts available:${NC}"
    echo "   - check_status.sh : Check service and API status"
    echo "   - logs.sh         : View service logs"
    echo "   - restart.sh      : Restart the service"
    echo "   - update.sh       : Rebuild and restart"
    echo "   - check_mongo.sh  : Check MongoDB status"
    echo "   - backup.sh       : Backup MongoDB database"
    echo ""
    echo -e "${BLUE}Service endpoints:${NC}"
    echo "   - API: http://localhost:8090"
    echo "   - MongoDB: mongodb://localhost:27017"
    echo ""
    echo -e "${GREEN}Full deployment instructions available at:${NC}"
    echo "   $PROJECT_DIR/DEPLOY_INSTRUCTIONS.md"
    echo ""
}

# Main installation flow
main() {
    echo -e "${BLUE}=================================${NC}"
    echo -e "${BLUE}Local API v2 Installation Script${NC}"
    echo -e "${BLUE}(Updated Chamber Architecture)${NC}"
    echo -e "${BLUE}=================================${NC}"
    echo ""
    
    check_root
    detect_ubuntu_version
    update_system
    install_go
    install_mongodb
    setup_mongodb_database
    create_service_user
    setup_project
    create_env_file
    create_systemd_service
    create_helper_scripts
    configure_firewall
    create_logrotate_config
    create_deployment_instructions
    
    # Verify installation
    verify_installation
    
    display_instructions
}

# Run main function
main "$@"