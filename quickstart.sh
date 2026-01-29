#!/bin/bash

# ============================================
# VN Stock Analysis System - Quick Start Script
# ============================================

set -e  # Exit on error

echo "=========================================="
echo "🇻🇳 VN Stock Analysis System Setup"
echo "=========================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored messages
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

print_info() {
    echo -e "ℹ $1"
}

# Check if Docker is installed
check_docker() {
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed!"
        echo "Please install Docker: https://docs.docker.com/get-docker/"
        exit 1
    fi
    print_success "Docker is installed"
}

# Check if Docker Compose is installed
check_docker_compose() {
    if ! command -v docker-compose &> /dev/null; then
        print_error "Docker Compose is not installed!"
        echo "Please install Docker Compose: https://docs.docker.com/compose/install/"
        exit 1
    fi
    print_success "Docker Compose is installed"
}

# Check environment file
check_env_file() {
    if [ ! -f .env ]; then
        print_warning ".env file not found. Creating from .env.example..."
        cp .env.example .env
        print_info "Please edit .env file with your configuration:"
        print_info "  - AWS credentials"
        print_info "  - Telegram bot token"
        print_info "  - Database passwords"
        print_info "  - API keys"
        echo ""
        read -p "Press Enter after you've configured .env file..."
    else
        print_success ".env file exists"
    fi
}

# Create required directories
create_directories() {
    print_info "Creating required directories..."
    
    mkdir -p agent-service/{models,cache}
    mkdir -p web-dashboard
    mkdir -p telegram-bot
    mkdir -p nginx/{ssl,logs}
    mkdir -p monitoring/{prometheus,grafana/dashboards}
    mkdir -p init-scripts
    mkdir -p workflows
    mkdir -p models/phobert
    
    print_success "Directories created"
}

# Download AI models (optional)
download_models() {
    read -p "Do you want to download PhoBERT model for local sentiment analysis? (y/n): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        print_info "Downloading PhoBERT model..."
        python3 -c "
from transformers import AutoTokenizer, AutoModel
model_name = 'vinai/phobert-base'
tokenizer = AutoTokenizer.from_pretrained(model_name)
model = AutoModel.from_pretrained(model_name)
tokenizer.save_pretrained('./models/phobert')
model.save_pretrained('./models/phobert')
print('Model downloaded successfully!')
        " || print_warning "Failed to download model. You can download it later."
        print_success "Model download completed"
    else
        print_info "Skipping model download. Will use API-based sentiment analysis."
    fi
}

# Initialize database schema
init_database() {
    print_info "Creating database initialization script..."
    
    cat > init-scripts/01-init.sql <<'EOF'
-- Initialize VN Stock Analysis Database

-- Create tables
CREATE TABLE IF NOT EXISTS stocks (
    id SERIAL PRIMARY KEY,
    symbol VARCHAR(10) UNIQUE NOT NULL,
    name VARCHAR(255),
    exchange VARCHAR(20),
    sector VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS analysis_results (
    id SERIAL PRIMARY KEY,
    symbol VARCHAR(10) NOT NULL,
    analysis_date DATE NOT NULL,
    recommendation VARCHAR(20),
    confidence DECIMAL(3,2),
    technical_score DECIMAL(5,2),
    sentiment_score DECIMAL(5,2),
    risk_level VARCHAR(20),
    current_price DECIMAL(12,2),
    target_price DECIMAL(12,2),
    raw_data JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(symbol, analysis_date)
);

CREATE TABLE IF NOT EXISTS news_articles (
    id SERIAL PRIMARY KEY,
    source VARCHAR(100),
    title TEXT NOT NULL,
    description TEXT,
    url TEXT,
    published_at TIMESTAMP,
    symbols TEXT[],
    sentiment VARCHAR(20),
    sentiment_score DECIMAL(3,2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS user_alerts (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(100) NOT NULL,
    symbol VARCHAR(10) NOT NULL,
    alert_type VARCHAR(50),
    threshold DECIMAL(5,2),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX idx_analysis_symbol_date ON analysis_results(symbol, analysis_date DESC);
CREATE INDEX idx_news_published ON news_articles(published_at DESC);
CREATE INDEX idx_news_symbols ON news_articles USING GIN(symbols);
CREATE INDEX idx_user_alerts_active ON user_alerts(user_id, is_active);

-- Insert some popular Vietnamese stocks
INSERT INTO stocks (symbol, name, exchange, sector) VALUES
    ('VNM', 'Vinamilk', 'HSX', 'Consumer Goods'),
    ('HPG', 'Hòa Phát Group', 'HSX', 'Materials'),
    ('VCB', 'Vietcombank', 'HSX', 'Financials'),
    ('VHM', 'Vinhomes', 'HSX', 'Real Estate'),
    ('VIC', 'Vingroup', 'HSX', 'Conglomerate'),
    ('FPT', 'FPT Corporation', 'HSX', 'Technology'),
    ('MSN', 'Masan Group', 'HSX', 'Consumer Goods'),
    ('MBB', 'MB Bank', 'HSX', 'Financials'),
    ('TCB', 'Techcombank', 'HSX', 'Financials'),
    ('BID', 'BIDV', 'HSX', 'Financials')
ON CONFLICT (symbol) DO NOTHING;

COMMENT ON TABLE analysis_results IS 'Stores daily stock analysis results from AI agents';
COMMENT ON TABLE news_articles IS 'Aggregated news from multiple sources';
EOF
    
    print_success "Database initialization script created"
}

# Setup AWS S3 structure
setup_s3() {
    read -p "Do you want to create AWS S3 bucket structure? (y/n): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        print_info "Creating S3 bucket structure..."
        
        cat > scripts/init_s3_structure.py <<'EOF'
import boto3
import os
from datetime import datetime

# Read from .env
from dotenv import load_dotenv
load_dotenv()

s3_client = boto3.client('s3')
bucket_name = os.getenv('S3_BUCKET', 'vnstock-data')

# Create bucket
try:
    s3_client.create_bucket(
        Bucket=bucket_name,
        CreateBucketConfiguration={'LocationConstraint': os.getenv('AWS_REGION', 'ap-southeast-1')}
    )
    print(f"✓ Created bucket: {bucket_name}")
except Exception as e:
    print(f"Bucket might already exist: {e}")

# Create folder structure
folders = [
    'raw-data/news/',
    'raw-data/market-data/',
    'raw-data/social/',
    'processed/sentiment/',
    'processed/technical/',
    'processed/combined/',
    'reports/daily/',
    'reports/weekly/',
    'reports/alerts/',
    'backups/'
]

for folder in folders:
    try:
        s3_client.put_object(Bucket=bucket_name, Key=folder)
        print(f"✓ Created folder: {folder}")
    except Exception as e:
        print(f"Error creating {folder}: {e}")

print("\n✓ S3 structure initialized successfully!")
EOF
        
        python3 scripts/init_s3_structure.py || print_warning "Failed to setup S3. You can do this manually later."
    else
        print_info "Skipping S3 setup"
    fi
}

# Setup Telegram bot
setup_telegram() {
    read -p "Do you have a Telegram bot token? (y/n): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        read -p "Enter your Telegram bot token: " bot_token
        
        # Update .env file
        if grep -q "TELEGRAM_BOT_TOKEN=" .env; then
            sed -i.bak "s|TELEGRAM_BOT_TOKEN=.*|TELEGRAM_BOT_TOKEN=$bot_token|" .env
        else
            echo "TELEGRAM_BOT_TOKEN=$bot_token" >> .env
        fi
        
        print_success "Telegram bot token configured"
        print_info "You can now chat with your bot to receive alerts!"
    else
        print_warning "Skipping Telegram setup. You can configure it later in .env"
        print_info "To create a bot: https://t.me/BotFather"
    fi
}

# Start services
start_services() {
    echo ""
    print_info "Starting services with Docker Compose..."
    echo ""
    
    # Choose profile
    echo "Choose deployment profile:"
    echo "1) Development (basic services)"
    echo "2) Production (with monitoring)"
    echo "3) Production + Monitoring + Local S3 (MinIO)"
    read -p "Enter choice (1-3): " -n 1 -r choice
    echo ""
    
    case $choice in
        1)
            docker-compose up -d
            ;;
        2)
            docker-compose --profile production --profile monitoring up -d
            ;;
        3)
            docker-compose --profile production --profile monitoring --profile development up -d
            ;;
        *)
            print_warning "Invalid choice. Starting basic services..."
            docker-compose up -d
            ;;
    esac
    
    print_success "Services started!"
}

# Wait for services to be ready
wait_for_services() {
    print_info "Waiting for services to be ready..."
    echo ""
    
    # Wait for PostgreSQL
    echo -n "Waiting for PostgreSQL..."
    until docker exec vnstock-db pg_isready -U admin -d vnstock &> /dev/null; do
        echo -n "."
        sleep 2
    done
    echo " Ready!"
    
    # Wait for Redis
    echo -n "Waiting for Redis..."
    until docker exec vnstock-cache redis-cli ping &> /dev/null; do
        echo -n "."
        sleep 2
    done
    echo " Ready!"
    
    # Wait for Agent Service
    echo -n "Waiting for Agent Service..."
    sleep 10
    until curl -f http://localhost:8000/health &> /dev/null; do
        echo -n "."
        sleep 3
    done
    echo " Ready!"
    
    print_success "All services are ready!"
}

# Import n8n workflow
import_workflow() {
    print_info "Importing n8n workflow..."
    
    sleep 5  # Give n8n time to start
    
    docker cp n8n-workflow-daily-analysis.json vnstock-n8n:/home/node/.n8n/workflows/ || {
        print_warning "Could not auto-import workflow. Please import manually via n8n UI."
        return
    }
    
    print_success "Workflow imported"
    print_info "You can activate it in n8n UI: http://localhost:5678"
}

# Print summary
print_summary() {
    echo ""
    echo "=========================================="
    echo "🎉 Setup Complete!"
    echo "=========================================="
    echo ""
    echo "Services are running at:"
    echo "  📊 Web Dashboard:    http://localhost:3000"
    echo "  🤖 n8n Workflows:    http://localhost:5678"
    echo "  🔧 Agent API:        http://localhost:8000"
    echo "  📈 Grafana:          http://localhost:3001"
    echo "  🗄️  MinIO (if enabled): http://localhost:9001"
    echo ""
    echo "Next steps:"
    echo "  1. Access n8n UI and activate the workflow"
    echo "  2. Test API: curl http://localhost:8000/health"
    echo "  3. Check Telegram bot: /start"
    echo "  4. View logs: docker-compose logs -f"
    echo ""
    echo "Documentation: README.md"
    echo "Troubleshooting: https://github.com/your-repo/wiki"
    echo ""
    print_success "Happy analyzing! 🚀"
    echo ""
}

# Main execution
main() {
    echo ""
    check_docker
    check_docker_compose
    check_env_file
    create_directories
    init_database
    
    # Optional steps
    download_models
    setup_s3
    setup_telegram
    
    # Start everything
    start_services
    wait_for_services
    import_workflow
    
    # Done!
    print_summary
}

# Run main function
main
