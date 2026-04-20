# 🇻🇳 Hệ Thống Phân Tích Thị Trường Chứng Khoán Việt Nam

![Version](https://img.shields.io/badge/version-1.0.0-blue.svg)
![License](https://img.shields.io/badge/license-MIT-green.svg)
![Python](https://img.shields.io/badge/python-3.9+-blue.svg)

Hệ thống phân tích chứng khoán tự động sử dụng AI Agents, n8n workflows, và real-time data processing để cung cấp insights toàn diện về thị trường chứng khoán Việt Nam.

## 🌟 Tính Năng Chính

### 📊 Phân Tích Đa Chiều
- **Technical Analysis**: RSI, MACD, Bollinger Bands, SMA/EMA, Stochastic, ADX
- **Sentiment Analysis**: Phân tích tâm lý từ tin tức và mạng xã hội
- **Trend Forecasting**: Dự báo xu hướng dựa trên ML models
- **Risk Assessment**: Đánh giá rủi ro đầu tư

### 🤖 AI Agent System
- **Technical Agent**: Phân tích kỹ thuật chuyên sâu
- **Sentiment Agent**: Xử lý ngôn ngữ tự nhiên tiếng Việt (PhoBERT)
- **Forecast Agent**: Tổng hợp và dự báo
- **Master Agent**: Điều phối và tổng hợp kết quả

### 📰 Thu Thập Dữ Liệu Tự Động
- RSS feeds từ VnEconomy, CafeF, VietStock, Đầu tư
- Market data real-time từ HSX/HNX
- Social sentiment từ Facebook Groups, Telegram channels
- Anti-spam filtering và rumor detection

### 🔔 Thông Báo Real-time
- Telegram bot với daily reports
- Email alerts cho các sự kiện quan trọng
- Web dashboard với live updates
- Custom alerts theo tiêu chí người dùng

### 📝 Blog & Content Platform

- **Blog Site** (Next.js 14) - Nội dung phân tích chứng khoán
- **Market Board** - Bảng giá real-time VN30/HNX30/UPCOM
- **Stock Screener** - Lọc cổ phiếu theo tiêu chí
- **Symbol Pages** - Trang chi tiết từng mã cổ phiếu
- **Newsletter** - Đăng ký nhận tin hàng ngày
- **AI-Powered** - Nội dung được tạo bởi Claude AI

## 🏗️ Kiến Trúc Hệ Thống

```
┌─────────────────────────────────────────────────────────────┐
│                   Data Sources                               │
│  RSS Feeds | Market APIs | Social Media | Internal Data     │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                n8n Workflow Orchestration                    │
│  Scraping | Scheduling | Data Pipeline | Error Handling     │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                   AWS S3 Storage                             │
│    Raw Data | Processed Data | Reports | Backups            │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                  AI Agent Processing                         │
│  Technical | Sentiment | Forecast | Master Orchestrator     │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│            Output & Distribution                             │
│  Web Dashboard | Telegram | Email | API | Mobile App        │
└─────────────────────────────────────────────────────────────┘
```

## 🚀 Quick Start

### Prerequisites

- Docker & Docker Compose
- AWS Account (cho S3 storage)
- Telegram Bot Token
- Python 3.9+
- Node.js 18+ (cho web dashboard)

### Installation

1. **Clone Repository**
```bash
git clone https://github.com/your-username/vnstock-analysis.git
cd vnstock-analysis
```

2. **Setup Environment Variables**
```bash
cp .env.example .env
# Edit .env file với các thông tin cần thiết
nano .env
```

3. **Download AI Models** (Optional - nếu dùng PhoBERT local)
```bash
mkdir -p models/phobert
# Download PhoBERT model từ HuggingFace
python scripts/download_models.py
```

4. **Start Services**
```bash
# Development
docker-compose up -d

# Production (with monitoring)
docker-compose --profile production --profile monitoring up -d
```

5. **Access Services**
- n8n Workflow UI: http://localhost:5678
- Web Dashboard: http://localhost:3000
- Agent API: http://localhost:8000
- Grafana Monitoring: http://localhost:3001

### Initial Configuration

1. **Import n8n Workflow**
```bash
# Copy workflow file vào n8n
docker cp n8n-workflow-daily-analysis.json vnstock-n8n:/home/node/.n8n/workflows/
```

2. **Setup Telegram Bot**
```bash
# Trong Telegram, chat với @BotFather
/newbot
# Follow instructions và lấy token
# Add token vào .env file
```

3. **Configure AWS S3**
```bash
# Create S3 bucket
aws s3 mb s3://vnstock-data --region ap-southeast-1

# Setup bucket structure
python scripts/init_s3_structure.py
```

## 📖 Sử Dụng

### 0. Blog & Market Dashboard

**Start Frontend:**
```bash
cd blog-site
npm install
npm run dev
```

**Access:**
- Homepage: http://localhost:3000
- Market Data: http://localhost:3000/market
- Articles: http://localhost:3000/articles
- Screener: http://localhost:3000/screener
- Symbol: http://localhost:3000/symbols/VNM

**Features:**
- Real-time VN30/HNX30/UPCOM index
- Live market price board
- AI-powered stock analysis
- Interactive TradingView charts
- Newsletter subscription

### 1. Phân Tích Một Mã Cổ Phiếu

**Via Python:**
```python
from technical_agent import TechnicalAnalysisAgent

# Phân tích VNM (Vinamilk)
agent = TechnicalAnalysisAgent(symbol='VNM', days=90)
analysis = agent.get_full_analysis()

print(f"Khuyến nghị: {analysis['signals']['recommendation']}")
print(f"Độ tin cậy: {analysis['signals']['confidence']:.1%}")
```

**Via API:**
```bash
curl -X POST http://localhost:8000/api/analyze/technical \
  -H "Content-Type: application/json" \
  -d '{"symbol": "VNM", "days": 90}'
```

**Via Telegram Bot:**
```
/analyze VNM
```

### 2. Lấy Báo Cáo Hàng Ngày

```bash
# Auto-generated daily report được gửi vào 8:00 AM
# Hoặc trigger manually:
curl http://localhost:8000/api/reports/daily
```

### 3. Subscribe Alerts

```bash
# Trong Telegram bot
/subscribe VNM HPG FPT
/alert_settings high_risk=true rumor_detection=true
```

## 🔧 Configuration

### n8n Workflow Customization

Edit `n8n-workflow-daily-analysis.json` để:
- Thay đổi thời gian chạy (mặc định 8:00 AM)
- Thêm/bớt nguồn RSS feeds
- Customize filtering logic
- Add more agents

### Agent Configuration

Edit `agent-service/config.py`:
```python
# Technical indicators
TECHNICAL_CONFIG = {
    'rsi_period': 14,
    'macd_fast': 12,
    'macd_slow': 26,
    'bollinger_period': 20,
    'sma_periods': [20, 50, 200]
}

# Sentiment analysis
SENTIMENT_CONFIG = {
    'model': 'vinai/phobert-base',
    'confidence_threshold': 0.7,
    'batch_size': 32
}
```

## 📊 Data Flow

### Daily Workflow (8:00 AM)

1. **Scraping Phase** (8:00 - 8:15)
   - Fetch RSS feeds from 5+ sources
   - Parse and filter spam
   - Save raw data to S3

2. **Hot Stock Detection** (8:15 - 8:20)
   - Analyze mentions across all sources
   - Identify top 10 most discussed stocks
   - Fetch market data for hot stocks

3. **Analysis Phase** (8:20 - 8:35)
   - Technical Agent: Calculate 15+ indicators
   - Sentiment Agent: Analyze 100+ news articles
   - Forecast Agent: Generate predictions

4. **Report Generation** (8:35 - 8:40)
   - Master Agent synthesizes results
   - Generate Markdown/HTML/JSON reports
   - Save to S3 and database

5. **Distribution** (8:40 - 8:45)
   - Send Telegram notifications
   - Update web dashboard
   - Send email alerts (if enabled)
   - Trigger webhook for mobile app

## 🛡️ Security & Compliance

### Legal Disclaimer

⚠️ **QUAN TRỌNG**: Hệ thống này chỉ cung cấp thông tin tham khảo, không phải lời khuyên đầu tư. Người dùng phải tự chịu trách nhiệm cho các quyết định đầu tư của mình.

### Data Privacy

- User data được mã hóa
- Không lưu trữ thông tin nhạy cảm
- Tuân thủ GDPR và PDPA (Việt Nam)

### API Rate Limiting

```python
# Mặc định: 1000 requests/hour
RATE_LIMITS = {
    'anonymous': 100,    # requests/hour
    'registered': 1000,  # requests/hour
    'premium': 10000     # requests/hour
}
```

## 🧪 Testing

```bash
# Run unit tests
pytest tests/

# Run integration tests
pytest tests/integration/

# Test individual agent
python tests/test_technical_agent.py

# Test n8n workflow
npm run test:workflow
```

## 📈 Performance

### Benchmarks (on 4 CPU cores, 8GB RAM)

- **Scraping**: 100 articles in ~30 seconds
- **Technical Analysis**: 1 stock in ~2 seconds
- **Sentiment Analysis**: 100 articles in ~45 seconds
- **Full Daily Report**: 10 stocks in ~5 minutes

### Optimization Tips

1. **Caching**: Enable Redis caching
2. **Parallel Processing**: Increase Celery workers
3. **Database Indexing**: Add indexes on frequently queried columns
4. **CDN**: Use CloudFront for static assets

## 🔍 Monitoring & Logging

### Prometheus Metrics

```yaml
# Available at http://localhost:9090
- scraping_success_rate
- analysis_duration_seconds
- api_requests_total
- error_rate
- queue_length
```

### Grafana Dashboards

Pre-configured dashboards:
1. System Overview
2. Analysis Performance
3. Data Pipeline Health
4. User Activity

### Logs

```bash
# View logs
docker-compose logs -f agent-service
docker-compose logs -f n8n

# Structured logging format
{
  "timestamp": "2026-01-28T08:00:00Z",
  "level": "INFO",
  "service": "technical-agent",
  "symbol": "VNM",
  "message": "Analysis completed",
  "duration_ms": 1234
}
```

## 🐛 Troubleshooting

### Common Issues

**1. n8n workflow không chạy**
```bash
# Check n8n logs
docker logs vnstock-n8n

# Restart n8n
docker-compose restart n8n
```

**2. Agent service crashed**
```bash
# Check logs
docker logs vnstock-agents

# Common causes:
# - Out of memory (increase Docker memory)
# - Missing API keys (check .env)
# - Network timeout (increase API_TIMEOUT_SECONDS)
```

**3. S3 upload failed**
```bash
# Verify AWS credentials
aws s3 ls s3://vnstock-data

# Check IAM permissions
# Ensure policy includes: s3:PutObject, s3:GetObject, s3:ListBucket
```

**4. Telegram bot không response**
```bash
# Verify bot token
curl https://api.telegram.org/bot<YOUR_TOKEN>/getMe

# Check network connectivity
docker exec vnstock-telegram ping api.telegram.org
```

## 🚧 Roadmap

### Phase 1: MVP (Tháng 1-2) ✅
- [x] Core infrastructure setup
- [x] Technical analysis agent
- [x] Sentiment analysis agent
- [x] Basic n8n workflow
- [x] Telegram bot

### Phase 2: Enhancement (Tháng 3-4)
- [ ] Social media scraping (Facebook, Telegram)
- [ ] Advanced ML models (LSTM, Transformer)
- [ ] Rumor detection system
- [ ] Web dashboard với real-time updates
- [ ] Email notification system

### Phase 3: Scale (Tháng 5-6)
- [ ] Portfolio tracking & optimization
- [ ] Mobile app (React Native)
- [ ] Custom alerts engine
- [ ] Backtesting framework
- [ ] Multi-user support

### Phase 4: Monetization (Tháng 7+)
- [ ] Premium subscription tiers
- [ ] API marketplace
- [ ] White-label solution
- [ ] Institutional-grade analytics
- [ ] Partner integrations

## 🤝 Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details.

### Development Workflow

1. Fork the repository
2. Create feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit changes (`git commit -m 'Add AmazingFeature'`)
4. Push to branch (`git push origin feature/AmazingFeature`)
5. Open Pull Request

### Code Style

- Python: Follow PEP 8
- JavaScript: Use ESLint with Airbnb config
- Commit messages: Use Conventional Commits

## 📜 License

This project is licensed under the MIT License - see [LICENSE](LICENSE) file.

## 🙏 Acknowledgments

- **vnstock** library cho market data
- **Anthropic Claude** cho AI capabilities
- **n8n** cho workflow automation
- **PhoBERT** cho Vietnamese NLP
- Vietnamese stock community

## 📞 Contact & Support

- **Email**: support@vnstock-analysis.com
- **Telegram**: [@vnstock_support](https://t.me/vnstock_support)
- **GitHub Issues**: [Create Issue](https://github.com/your-username/vnstock-analysis/issues)
- **Documentation**: [Wiki](https://github.com/your-username/vnstock-analysis/wiki)

## ⚖️ Disclaimer

Hệ thống này được phát triển cho mục đích giáo dục và nghiên cứu. Tác giả không chịu trách nhiệm cho bất kỳ tổn thất tài chính nào phát sinh từ việc sử dụng hệ thống. Người dùng phải tự nghiên cứu kỹ lưỡng và chịu trách nhiệm cho các quyết định đầu tư của mình.

---

**Made with ❤️ in Vietnam 🇻🇳**

**⭐ Star this repo if you find it useful!**
