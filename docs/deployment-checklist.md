# Deployment Checklist

**Project**: VNStock Analysis System
**Last Updated**: 2026-04-20

## Mục Lục

1. [Prerequisites](#1-prerequisites)
2. [Infrastructure Setup](#2-infrastructure-setup)
3. [Environment Variables](#3-environment-variables)
4. [Database Setup](#4-database-setup)
5. [Docker Services](#5-docker-services)
6. [Blog Frontend](#6-blog-frontend)
7. [n8n Workflows](#7-n8n-workflows)
8. [Monitoring](#8-monitoring)
9. [Security](#9-security)
10. [Post-Deployment](#10-post-deployment)

---

## 1. Prerequisites

### System Requirements
| Resource | Minimum | Recommended |
|----------|---------|-------------|
| CPU | 4 cores | 8 cores |
| RAM | 8 GB | 16 GB |
| Storage | 50 GB SSD | 100 GB SSD |
| OS | Ubuntu 20.04+ | Ubuntu 22.04 LTS |

### Required Accounts
- [ ] Docker Hub account
- [ ] AWS account (for S3)
- [ ] Telegram Bot (@BotFather)
- [ ] Anthropic API key
- [ ] (Optional) Vercel account (for blog)

### Software
- [ ] Docker 20.10+
- [ ] Docker Compose 2.0+
- [ ] Git
- [ ] Node.js 18+ (for local dev)
- [ ] Go 1.21+ (for local dev)
- [ ] Python 3.9+ (for local dev)

---

## 2. Infrastructure Setup

### Cloud Resources
- [ ] **AWS Region**: ap-southeast-1 (Singapore)
- [ ] **VPC**: Create VPC với private subnets
- [ ] **RDS**: PostgreSQL 15+ instance
- [ ] **ElastiCache**: Redis 7+ cluster
- [ ] **S3**: Buckets for data và backups
- [ ] **ECR**: Container registry

### DNS Configuration
- [ ] Domain registered (e.g., `vnstock.com`)
- [ ] DNS A records:
  - `api.vnstock.com` → API Gateway
  - `blog.vnstock.com` → Vercel/CDN
  - `n8n.vnstock.com` → n8n instance
  - `grafana.vnstock.com` → Monitoring
- [ ] SSL certificates (Let's Encrypt hoặc AWS ACM)

### Security Groups
- [ ] Allow port 22 (SSH) from trusted IPs only
- [ ] Allow port 443 (HTTPS) from 0.0.0.0/0
- [ ] Allow port 80 (HTTP) from 0.0.0.0/0 (redirect to HTTPS)
- [ ] Allow internal ports: 5432, 6379, 8080-8089

---

## 3. Environment Variables

### Critical Variables (Production)
```bash
# .env.production
ENVIRONMENT=production

# Database
POSTGRES_PASSWORD=<secure-random-32-chars>
DB_HOST=<rds-endpoint>.rds.amazonaws.com
DB_PORT=5432
DB_USER=vnstock
DB_NAME=vnstock

# Redis
REDIS_PASSWORD=<secure-random-32-chars>
REDIS_HOST=<elasticache-endpoint>.cache.amazonaws.com

# AI APIs
ANTHROPIC_API_KEY=sk-ant-...

# Telegram
TELEGRAM_BOT_TOKEN=123456:ABC...
TELEGRAM_CHANNEL_ID=@your_channel

# AWS
AWS_ACCESS_KEY_ID=AKIA...
AWS_SECRET_ACCESS_KEY=...
AWS_REGION=ap-southeast-1
S3_BUCKET=vnstock-data

# Security
JWT_SECRET_KEY=<secure-random-64-chars>
ENABLE_AUTH=true
CORS_ALLOWED_ORIGINS=https://blog.vnstock.com
```

### Blog Frontend Variables
```bash
# blog-site/.env.local
NEXT_PUBLIC_API_URL=https://api.vnstock.com
NEXT_PUBLIC_MARKET_API_URL=https://api.vnstock.com/market
```

### Generate Secure Secrets
```bash
# Generate JWT secret
openssl rand -base64 64

# Generate database password
openssl rand -hex 32
```

---

## 4. Database Setup

### PostgreSQL
```bash
# Connect to RDS
psql -h <rds-endpoint> -U admin -d postgres

# Create database
CREATE DATABASE vnstock;
CREATE USER vnstock WITH PASSWORD '<password>';
GRANT ALL PRIVILEGES ON DATABASE vnstock TO vnstock;

# Run migrations
psql -h <rds-endpoint> -U vnstock -d vnstock -f migrations/001_init.sql
```

### Redis
```bash
# Configure ElastiCache
# Enable encryption in-transit
# Enable AUTH token
```

### S3 Bucket Structure
```
vnstock-data/
├── raw-data/
│   ├── news/{year}/{month}/{day}/
│   ├── market-data/{year}/{month}/{day}/
│   └── social/{year}/{month}/{day}/
├── processed/
│   ├── sentiment/
│   ├── technical/
│   └── combined/
└── reports/
    ├── daily/
    ├── weekly/
    └── alerts/
```

---

## 5. Docker Services

### Build & Push Images
```bash
# Login to ECR
aws ecr get-login-password --region ap-southeast-1 | docker login --username AWS --password-stdin <account>.dkr.ecr.ap-southeast-1.amazonaws.com

# Build Go services
cd go-services
docker build -f ../docker/go.Dockerfile -t vnstock-api:latest .
docker tag vnstock-api:latest <ecr-url>/vnstock-api:latest
docker push <ecr-url>/vnstock-api:latest

# Build sentiment service
cd python-sentiment
docker build -t vnstock-sentiment:latest .
docker tag vnstock-sentiment:latest <ecr-url>/vnstock-sentiment:latest
docker push <ecr-url>/vnstock-sentiment:latest
```

### Deploy Services
```bash
# Copy compose file
scp docker-compose.yml user@server:/opt/vnstock/

# SSH to server
ssh user@server

# Pull images
docker-compose pull

# Start services
docker-compose --profile production up -d

# Check status
docker-compose ps
docker-compose logs -f
```

### Service Health Checks
```bash
# API Gateway
curl https://api.vnstock.com/health

# Technical Agent
curl https://api.vnstock.com/technical/health

# Sentiment Service
curl https://api.vnstock.com/sentiment/health
```

---

## 6. Blog Frontend

### Vercel Deployment (Recommended)
```bash
cd blog-site

# Install Vercel CLI
npm i -g vercel

# Login
vercel login

# Deploy
vercel --prod

# Set environment variables in Vercel Dashboard
# NEXT_PUBLIC_API_URL=https://api.vnstock.com
```

### Alternative: Docker + Nginx
```bash
# Build
docker build -t vnstock-blog:latest -f docker/Dockerfile blog/

# docker-compose.yml addition
blog:
  image: vnstock-blog:latest
  ports:
    - "3000:3000"
  environment:
    - NEXT_PUBLIC_API_URL=https://api.vnstock.com
```

---

## 7. n8n Workflows

### Setup n8n
```bash
# Access n8n
# https://n8n.vnstock.com

# Import workflows
# 1. n8n-workflow-daily-analysis.json
# 2. n8n-workflow-error-handler.json

# Configure credentials
# - Telegram Bot API
# - AWS S3
# - Anthropic API
# - PostgreSQL
```

### Workflow Configuration
- [ ] Set trigger schedule (daily 8:00 AM)
- [ ] Configure webhook URLs
- [ ] Set up error notifications
- [ ] Test workflow execution

---

## 8. Monitoring

### Prometheus
```bash
# Access: http://prometheus.vnstock.com

# Verify targets
curl https://api.vnstock.com/metrics
```

### Grafana
```bash
# Access: http://grafana.vnstock.com
# Default: admin / <GRAFANA_PASSWORD>

# Import dashboards from:
# - monitoring/grafana/dashboards/
```

### Alert Rules
- [ ] API error rate > 1%
- [ ] Response time > 2s
- [ ] Service down
- [ ] Disk usage > 80%
- [ ] Memory usage > 85%

### Telegram Alerts
- [ ] Critical errors
- [ ] Service restarts
- [ ] Daily reports

---

## 9. Security

### SSL/TLS
```bash
# Let's Encrypt (auto-renewal)
certbot --nginx -d api.vnstock.com -d blog.vnstock.com

# AWS ACM (if using ALB)
# Import certificates in ACM console
```

### Firewall
```bash
# UFW (Ubuntu)
ufw default deny incoming
ufw default allow outgoing
ufw allow ssh
ufw allow 443/tcp
ufw allow 80/tcp
ufw enable
```

### Secrets Management
- [ ] Use AWS Secrets Manager for production
- [ ] Rotate API keys quarterly
- [ ] Enable MFA for all accounts

### Backup
```bash
# Automated backups enabled in RDS
# Point-in-time recovery tested
# S3 versioning enabled
# Backup schedule: Daily 2 AM
```

---

## 10. Post-Deployment

### Testing Checklist
```bash
# API Tests
curl -X POST https://api.vnstock.com/api/v1/technical/VNM
curl -X POST https://api.vnstock.com/api/v1/sentiment -d '{"text":"VNM tăng giá"}'

# Blog Tests
curl https://blog.vnstock.com/
curl https://blog.vnstock.com/market
curl https://blog.vnstock.com/articles

# Telegram Bot
/subscribe VNM
/analyze VNM
/daily_report
```

### DNS Propagation Check
```bash
# Verify DNS
dig api.vnstock.com
nslookup vnstock.com

# SSL verification
openssl s_client -connect api.vnstock.com:443
```

### Performance Baseline
- [ ] TTFB < 200ms
- [ ] API response < 500ms
- [ ] Blog page load < 2s
- [ ] Database query < 100ms

### Documentation
- [ ] Update architecture diagrams
- [ ] Document incident procedures
- [ ] Create runbooks
- [ ] Backup configuration

### Team Handoff
- [ ] Share access credentials (securely)
- [ ] Train team members
- [ ] Establish on-call rotation
- [ ] Create escalation procedures

---

## Quick Reference

### Service URLs (Production)
| Service | URL |
|---------|-----|
| API Gateway | https://api.vnstock.com |
| Blog | https://blog.vnstock.com |
| n8n | https://n8n.vnstock.com |
| Grafana | https://grafana.vnstock.com |
| Prometheus | https://prometheus.vnstock.com |

### Ports
| Port | Service |
|------|---------|
| 80/443 | Nginx (HTTP/HTTPS) |
| 5432 | PostgreSQL |
| 6379 | Redis |
| 8080 | API Gateway |
| 8081 | Technical Agent |
| 8082 | Forecast Agent |
| 8000 | Sentiment Service |

### Emergency Commands
```bash
# Restart all services
docker-compose restart

# View logs
docker-compose logs -f --tail=100

# Scale API
docker-compose up -d --scale api-gateway=3

# Rollback
docker-compose down && docker-compose -f docker-compose.yml.backup up -d
```

---

## Related Documentation
- [System Architecture](./system-architecture.md)
- [Blog Features](./blog-features.md)
- [Code Standards](./code-standards.md)
