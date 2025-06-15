#!/bin/bash

# CBT Protocols Ubuntu Initialization Script
# This script sets up a clean Ubuntu system with all dependencies for CBT Protocols

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo -e "${PURPLE}[HEADER]${NC} $1"
}

# Configuration
PROJECT_NAME="cbt-protocols"
PROJECT_DIR="/opt/$PROJECT_NAME"
SERVICE_USER="cbt"
GO_VERSION="1.21.5"
NODE_VERSION="20"
MONGODB_VERSION="7.0"

print_header "CBT Protocols Ubuntu Initialization Script"
print_header "=========================================="

# Check if running as root
if [[ $EUID -eq 0 ]]; then
    print_warning "Running as root. This script will create a dedicated user for the application."
    ROOT_USER=true
else
    print_error "This script requires root privileges. Please run with sudo:"
    print_error "sudo $0"
    exit 1
fi

# Check Ubuntu version
print_status "Checking Ubuntu version..."
if [[ -f /etc/os-release ]]; then
    . /etc/os-release
    if [[ "$ID" != "ubuntu" ]]; then
        print_error "This script is designed for Ubuntu. Detected: $ID"
        exit 1
    fi
    print_success "Ubuntu $VERSION_ID detected"
else
    print_error "Cannot determine OS version"
    exit 1
fi

# Update system
print_header "Step 1: Updating system packages..."
apt update
apt upgrade -y
apt install -y curl wget git build-essential software-properties-common apt-transport-https ca-certificates gnupg lsb-release

print_success "System packages updated"

# Install MongoDB
print_header "Step 2: Installing MongoDB..."
if ! command -v mongod &> /dev/null; then
    print_status "Installing MongoDB $MONGODB_VERSION..."
    
    # Import MongoDB GPG key
    curl -fsSL https://www.mongodb.org/static/pgp/server-$MONGODB_VERSION.asc | gpg -o /usr/share/keyrings/mongodb-server-$MONGODB_VERSION.gpg --dearmor
    
    # Add MongoDB repository
    echo "deb [ arch=amd64,arm64 signed-by=/usr/share/keyrings/mongodb-server-$MONGODB_VERSION.gpg ] https://repo.mongodb.org/apt/ubuntu $(lsb_release -cs)/mongodb-org/$MONGODB_VERSION multiverse" | tee /etc/apt/sources.list.d/mongodb-org-$MONGODB_VERSION.list
    
    # Install MongoDB
    apt update
    apt install -y mongodb-org
    
    # Start and enable MongoDB
    systemctl start mongod
    systemctl enable mongod
    
    print_success "MongoDB installed and started"
else
    print_success "MongoDB already installed"
fi

# Install Go
print_header "Step 3: Installing Go..."
if ! command -v go &> /dev/null; then
    print_status "Installing Go $GO_VERSION..."
    
    # Download Go
    cd /tmp
    wget https://go.dev/dl/go$GO_VERSION.linux-amd64.tar.gz
    
    # Remove any previous Go installation
    rm -rf /usr/local/go
    
    # Extract Go
    tar -C /usr/local -xzf go$GO_VERSION.linux-amd64.tar.gz
    
    # Add Go to PATH for all users
    echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
    echo 'export GOPATH=$HOME/go' >> /etc/profile
    echo 'export PATH=$PATH:$GOPATH/bin' >> /etc/profile
    
    # Set PATH for current session
    export PATH=$PATH:/usr/local/go/bin
    
    print_success "Go $GO_VERSION installed"
else
    print_success "Go already installed"
fi

# Install Node.js
print_header "Step 4: Installing Node.js..."
if ! command -v node &> /dev/null; then
    print_status "Installing Node.js $NODE_VERSION..."
    
    # Install NodeSource repository
    curl -fsSL https://deb.nodesource.com/setup_$NODE_VERSION.x | bash -
    
    # Install Node.js
    apt install -y nodejs
    
    print_success "Node.js $(node --version) installed"
else
    print_success "Node.js already installed"
fi

# Create service user
print_header "Step 5: Creating service user..."
if ! id "$SERVICE_USER" &>/dev/null; then
    print_status "Creating user: $SERVICE_USER"
    useradd --system --shell /bin/bash --home-dir /home/$SERVICE_USER --create-home $SERVICE_USER
    print_success "User $SERVICE_USER created"
else
    print_success "User $SERVICE_USER already exists"
fi

# Create project directory
print_header "Step 6: Creating project directory..."
print_status "Creating project directory: $PROJECT_DIR"
mkdir -p $PROJECT_DIR
chown $SERVICE_USER:$SERVICE_USER $PROJECT_DIR
print_success "Project directory created"

# Download or clone project (placeholder - replace with actual repository)
print_header "Step 7: Setting up project files..."
print_status "Setting up project structure..."

# Create basic project structure (this should be replaced with actual git clone)
sudo -u $SERVICE_USER mkdir -p $PROJECT_DIR/{backend_v2,frontend_v2.1}

# Create a simple download script for the user
cat > /home/$SERVICE_USER/download-project.sh << 'EOF'
#!/bin/bash

# CBT Protocols Project Download Script
# Replace this with your actual project download method

echo "Please download or clone your CBT Protocols project to:"
echo "/opt/cbt-protocols/"
echo ""
echo "If you have a git repository:"
echo "cd /opt/cbt-protocols"
echo "git clone YOUR_REPOSITORY_URL ."
echo ""
echo "Or if you have project files:"
echo "1. Upload your project files to this server"
echo "2. Extract them to /opt/cbt-protocols/"
echo ""
echo "Make sure the structure looks like:"
echo "/opt/cbt-protocols/backend_v2/"
echo "/opt/cbt-protocols/frontend_v2.1/"
echo ""
echo "After placing the files, run:"
echo "sudo chown -R cbt:cbt /opt/cbt-protocols/"
echo "cd /opt/cbt-protocols"
echo "chmod +x build.sh"
echo "./build.sh"
EOF

chown $SERVICE_USER:$SERVICE_USER /home/$SERVICE_USER/download-project.sh
chmod +x /home/$SERVICE_USER/download-project.sh

print_success "Project setup completed"

# Create systemd service
print_header "Step 8: Creating systemd service..."
cat > /etc/systemd/system/cbt-protocols.service << EOF
[Unit]
Description=CBT Protocols - Climate Chamber Management System
After=network.target mongod.service
Requires=mongod.service

[Service]
Type=simple
User=$SERVICE_USER
Group=$SERVICE_USER
WorkingDirectory=$PROJECT_DIR/backend_v2
ExecStart=$PROJECT_DIR/backend_v2/cbt-protocols
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=cbt-protocols

# Environment variables
Environment=GIN_MODE=release
Environment=PORT=8081

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=$PROJECT_DIR

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd
systemctl daemon-reload

print_success "Systemd service created"

# Setup firewall (if ufw is available)
print_header "Step 9: Configuring firewall..."
if command -v ufw &> /dev/null; then
    print_status "Configuring UFW firewall..."
    ufw allow 8081/tcp comment "CBT Protocols Web Interface"
    print_success "Firewall configured"
else
    print_warning "UFW not available. Please manually configure firewall to allow port 8081"
fi

# Create useful scripts
print_header "Step 10: Creating management scripts..."

# Create start script
cat > /home/$SERVICE_USER/start-cbt.sh << 'EOF'
#!/bin/bash
echo "Starting CBT Protocols service..."
sudo systemctl start cbt-protocols
sudo systemctl status cbt-protocols
EOF

# Create stop script
cat > /home/$SERVICE_USER/stop-cbt.sh << 'EOF'
#!/bin/bash
echo "Stopping CBT Protocols service..."
sudo systemctl stop cbt-protocols
EOF

# Create status script
cat > /home/$SERVICE_USER/status-cbt.sh << 'EOF'
#!/bin/bash
echo "CBT Protocols service status:"
sudo systemctl status cbt-protocols
echo ""
echo "CBT Protocols logs:"
sudo journalctl -u cbt-protocols -n 20 --no-pager
EOF

# Create build script
cat > /home/$SERVICE_USER/build-cbt.sh << 'EOF'
#!/bin/bash
echo "Building CBT Protocols..."
cd /opt/cbt-protocols
sudo -u cbt ./build.sh
echo "Restarting service..."
sudo systemctl restart cbt-protocols
sudo systemctl status cbt-protocols
EOF

# Make scripts executable
chmod +x /home/$SERVICE_USER/*.sh
chown $SERVICE_USER:$SERVICE_USER /home/$SERVICE_USER/*.sh

print_success "Management scripts created"

# Setup MongoDB database and user
print_header "Step 11: Setting up MongoDB database..."
print_status "Creating MongoDB database and user..."

# Create MongoDB init script
cat > /tmp/mongodb-init.js << 'EOF'
use admin;

// Create database
use cbt_protocols_v2;

// Create user with read/write access
db.createUser({
  user: "cbt_user",
  pwd: "cbt_password_please_change_this",
  roles: [
    { role: "readWrite", db: "cbt_protocols_v2" }
  ]
});

// Create initial admin user (you should change this password)
db.users.insertOne({
  username: "admin",
  password: "$2a$10$rGhHaWzfKy1T6FhG2KHuWO4XnF5/lF2S3F1nP4CqR4Y5Z7D1H8X3Q",
  name: "Administrator",
  role: "admin",
  is_active: true,
  created_at: new Date(),
  updated_at: new Date()
});

print("MongoDB setup completed");
EOF

# Execute MongoDB script
if systemctl is-active --quiet mongod; then
    mongosh < /tmp/mongodb-init.js || true
    rm /tmp/mongodb-init.js
    print_success "MongoDB database and user created"
else
    print_warning "MongoDB is not running. Database setup skipped."
    print_status "You can run the database setup later with:"
    print_status "mongosh < /tmp/mongodb-init.js"
fi

# Create environment file template
print_header "Step 12: Creating environment configuration..."
cat > $PROJECT_DIR/backend_v2/.env.example << 'EOF'
# MongoDB Configuration
MONGODB_URI=mongodb://cbt_user:cbt_password_please_change_this@localhost:27017/cbt_protocols_v2?authSource=cbt_protocols_v2
MONGODB_DATABASE=cbt_protocols_v2

# Server Configuration
PORT=8081
GIN_MODE=release

# JWT Configuration (CHANGE THESE)
JWT_SECRET=your-very-secure-secret-key-please-change-this-to-something-random
JWT_EXPIRATION=24h

# API Configuration
API_KEY=your-api-key-here

# Chamber Configuration
HEARTBEAT_TIMEOUT=60  # seconds
CLEANUP_INTERVAL=300  # seconds
EOF

chown $SERVICE_USER:$SERVICE_USER $PROJECT_DIR/backend_v2/.env.example

print_success "Environment configuration created"

# Create setup completion script
cat > /home/$SERVICE_USER/complete-setup.sh << 'EOF'
#!/bin/bash

echo "CBT Protocols Setup Completion"
echo "=============================="
echo ""
echo "1. Download your project files:"
echo "   - Place your project files in /opt/cbt-protocols/"
echo "   - Or clone your git repository there"
echo ""
echo "2. Configure environment:"
echo "   cd /opt/cbt-protocols/backend_v2"
echo "   cp .env.example .env"
echo "   nano .env  # Edit with your settings"
echo ""
echo "3. Build the project:"
echo "   cd /opt/cbt-protocols"
echo "   ./build.sh"
echo ""
echo "4. Start the service:"
echo "   sudo systemctl enable cbt-protocols"
echo "   sudo systemctl start cbt-protocols"
echo ""
echo "5. Check status:"
echo "   ./status-cbt.sh"
echo ""
echo "Default admin login:"
echo "Username: admin"
echo "Password: admin123 (CHANGE THIS!)"
echo ""
echo "Application will be available at:"
echo "http://YOUR_SERVER_IP:8081"
EOF

chmod +x /home/$SERVICE_USER/complete-setup.sh
chown $SERVICE_USER:$SERVICE_USER /home/$SERVICE_USER/complete-setup.sh

# Final summary
print_header "Installation Complete!"
print_success "CBT Protocols Ubuntu setup completed successfully!"
print_status ""
print_status "Summary of what was installed:"
print_status "- MongoDB $MONGODB_VERSION"
print_status "- Go $GO_VERSION"
print_status "- Node.js $(node --version 2>/dev/null || echo 'Unknown')"
print_status "- Project structure in $PROJECT_DIR"
print_status "- Service user: $SERVICE_USER"
print_status "- Systemd service: cbt-protocols"
print_status ""
print_status "Next steps:"
print_status "1. Switch to the service user:"
print_status "   sudo su - $SERVICE_USER"
print_status ""
print_status "2. Run the completion script:"
print_status "   ./complete-setup.sh"
print_status ""
print_status "3. Or follow manual steps:"
print_status "   - Place project files in $PROJECT_DIR"
print_status "   - Configure .env file"
print_status "   - Run build script"
print_status "   - Start service"
print_status ""
print_status "Available management commands:"
print_status "- ./start-cbt.sh    - Start the service"
print_status "- ./stop-cbt.sh     - Stop the service"
print_status "- ./status-cbt.sh   - Check service status"
print_status "- ./build-cbt.sh    - Rebuild and restart"
print_status ""
print_status "Default admin credentials:"
print_status "Username: admin"
print_status "Password: admin123 (CHANGE THIS!)"
print_status ""
print_warning "Important: Change default passwords and configure .env file!"
print_status ""
print_status "Access the application at: http://$(hostname -I | awk '{print $1}'):8081"