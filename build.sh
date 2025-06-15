#!/bin/bash

# CBT Protocols Build Script
# This script builds the frontend and prepares it for serving from the backend

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$SCRIPT_DIR"

# Define paths
FRONTEND_DIR="$PROJECT_ROOT/frontend_v2.1"
BACKEND_DIR="$PROJECT_ROOT/backend_v2"
FRONTEND_DIST_DIR="$FRONTEND_DIR/dist"
BACKEND_FRONTEND_DIR="$BACKEND_DIR/frontend"

print_status "Starting CBT Protocols build process..."
print_status "Project root: $PROJECT_ROOT"

# Check if directories exist
if [[ ! -d "$FRONTEND_DIR" ]]; then
    print_error "Frontend directory not found: $FRONTEND_DIR"
    exit 1
fi

if [[ ! -d "$BACKEND_DIR" ]]; then
    print_error "Backend directory not found: $BACKEND_DIR"
    exit 1
fi

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    print_error "Node.js is not installed. Please install Node.js first."
    exit 1
fi

# Check if npm is installed
if ! command -v npm &> /dev/null; then
    print_error "npm is not installed. Please install npm first."
    exit 1
fi

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed. Please install Go first."
    exit 1
fi

print_status "All dependencies found."

# Step 1: Build Frontend
print_status "Building frontend..."
cd "$FRONTEND_DIR"

# Install dependencies if node_modules doesn't exist
if [[ ! -d "node_modules" ]]; then
    print_status "Installing frontend dependencies..."
    npm install
fi

# Build frontend
print_status "Running frontend build..."
npm run build

# Check if build was successful
if [[ ! -d "$FRONTEND_DIST_DIR" ]]; then
    print_error "Frontend build failed - dist directory not found"
    exit 1
fi

print_success "Frontend build completed successfully"

# Step 2: Copy frontend dist to backend
print_status "Copying frontend dist to backend..."

# Create backend frontend directory if it doesn't exist
mkdir -p "$BACKEND_FRONTEND_DIR"

# Remove old dist files
if [[ -d "$BACKEND_FRONTEND_DIR/dist" ]]; then
    print_status "Removing old frontend files..."
    rm -rf "$BACKEND_FRONTEND_DIR/dist"
fi

# Copy new dist files
print_status "Copying new frontend files..."
cp -r "$FRONTEND_DIST_DIR" "$BACKEND_FRONTEND_DIR/"

print_success "Frontend files copied to backend"

# Step 3: Build Backend
print_status "Building backend..."
cd "$BACKEND_DIR"

# Download Go dependencies
print_status "Downloading Go dependencies..."
go mod download

# Build backend
print_status "Running backend build..."
go build -o cbt-protocols .

# Check if build was successful
if [[ ! -f "cbt-protocols" ]]; then
    print_error "Backend build failed"
    exit 1
fi

print_success "Backend build completed successfully"

# Step 4: Create start script
print_status "Creating start script..."
cat > "$BACKEND_DIR/start.sh" << 'EOF'
#!/bin/bash

# CBT Protocols Start Script

# Check if .env file exists
if [[ ! -f ".env" ]]; then
    echo "Warning: .env file not found. Creating from example..."
    if [[ -f "env.example" ]]; then
        cp env.example .env
        echo "Please edit .env file with your configuration"
    else
        echo "Error: env.example not found"
        exit 1
    fi
fi

# Start the application
echo "Starting CBT Protocols..."
echo "Frontend available at: http://localhost:8081"
echo "API available at: http://localhost:8081/api"
echo "Press Ctrl+C to stop"

./cbt-protocols
EOF

chmod +x "$BACKEND_DIR/start.sh"

# Step 5: Create env.example if it doesn't exist
if [[ ! -f "$BACKEND_DIR/env.example" ]]; then
    print_status "Creating env.example..."
    cat > "$BACKEND_DIR/env.example" << 'EOF'
# MongoDB Configuration
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=cbt_protocols_v2

# Server Configuration
PORT=8081
GIN_MODE=release

# JWT Configuration
JWT_SECRET=your-secret-key-here-please-change-this
JWT_EXPIRATION=24h

# API Configuration
API_KEY=your-api-key-here

# Chamber Configuration
HEARTBEAT_TIMEOUT=60  # seconds
CLEANUP_INTERVAL=300  # seconds
EOF
fi

print_success "Build process completed successfully!"
print_status ""
print_status "Next steps:"
print_status "1. Configure your environment:"
print_status "   cd $BACKEND_DIR"
print_status "   cp env.example .env"
print_status "   # Edit .env with your configuration"
print_status ""
print_status "2. Start the application:"
print_status "   ./start.sh"
print_status ""
print_status "3. Or run directly:"
print_status "   ./cbt-protocols"
print_status ""
print_status "The application will be available at:"
print_status "- Frontend: http://localhost:8081"
print_status "- API: http://localhost:8081/api"