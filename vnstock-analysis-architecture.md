# Hệ Thống Phân Tích Thị Trường Chứng khoán Việt Nam
## Kiến Trúc Tổng Quan

```
┌─────────────────────────────────────────────────────────────────┐
│                    DATA INGESTION LAYER                          │
├─────────────────────────────────────────────────────────────────┤
│  RSS/News APIs  │  Market Data APIs  │  Social Sentiment       │
│  (VnEconomy,    │  (FiinGroup,       │  (Facebook, Telegram,   │
│   CafeF, HSX)   │   EODHD, VNDirect) │   Zalo Groups)          │
└──────────┬──────────────────┬────────────────────┬──────────────┘
           │                  │                    │
           ▼                  ▼                    ▼
┌─────────────────────────────────────────────────────────────────┐
│                    n8n WORKFLOW ORCHESTRATOR                     │
├─────────────────────────────────────────────────────────────────┤
│  • Scheduled Scraping (Daily 8:00 AM)                           │
│  • Data Validation & Deduplication                              │
│  • Format Normalization                                         │
│  • Error Handling & Retry Logic                                 │
└──────────┬──────────────────────────────────────────────────────┘
           │
           ▼
┌─────────────────────────────────────────────────────────────────┐
│                    AWS S3 STORAGE LAYER                          │
├─────────────────────────────────────────────────────────────────┤
│  /raw-data/          │  /processed/        │  /reports/         │
│  - news/             │  - sentiment/       │  - daily/          │
│  - market-data/      │  - technical/       │  - weekly/         │
│  - social/           │  - combined/        │  - alerts/         │
└──────────┬──────────────────────────────────────────────────────┘
           │
           ▼
┌─────────────────────────────────────────────────────────────────┐
│                    AI AGENT PROCESSING LAYER                     │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─────────────────┐  ┌──────────────────┐  ┌────────────────┐│
│  │ Technical Agent │  │ Sentiment Agent  │  │ Forecast Agent ││
│  │                 │  │                  │  │                ││
│  │ • RSI, MACD     │  │ • PhoBERT/GPT-4o │  │ • Trend Pred.  ││
│  │ • SMA, EMA      │  │ • News Analysis  │  │ • Risk Score   ││
│  │ • Bollinger     │  │ • Social Mining  │  │ • Recommend.   ││
│  │ • Volume        │  │ • Entity Extract │  │ • Confidence   ││
│  └────────┬────────┘  └────────┬─────────┘  └────────┬───────┘│
│           │                    │                      │        │
│           └────────────────────┼──────────────────────┘        │
│                                ▼                                │
│                    ┌──────────────────────┐                     │
│                    │   Master Agent       │                     │
│                    │   (Orchestrator)     │                     │
│                    │                      │                     │
│                    │ • Data Aggregation   │                     │
│                    │ • Cross-validation   │                     │
│                    │ • Report Generation  │                     │
│                    └──────────┬───────────┘                     │
│                               │                                 │
└───────────────────────────────┼─────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                    OUTPUT & DISTRIBUTION LAYER                   │
├─────────────────────────────────────────────────────────────────┤
│  Web Dashboard  │  Telegram Bot  │  Email Alerts  │  API       │
│  (Next.js)      │  (Real-time)   │  (Daily/Event) │  (REST)    │
└─────────────────────────────────────────────────────────────────┘
```

---

## 1. DATA INGESTION - Thu Thập Dữ Liệu

### 1.1 Nguồn Chính Thống (Official Sources)

**A. RSS Feeds & News APIs:**
```javascript
const officialSources = {
  vneconomy: {
    rss: 'https://vneconomy.vn/rss/chung-khoan.rss',
    priority: 'high',
    reliability: 0.9
  },
  cafef: {
    rss: 'https://cafef.vn/chung-khoan.rss',
    priority: 'high',
    reliability: 0.85
  },
  vietstock: {
    api: 'https://api.vietstock.vn/finance/...',
    priority: 'medium',
    reliability: 0.8
  },
  ndh: {
    rss: 'https://ndh.vn/chung-khoan.rss',
    priority: 'medium',
    reliability: 0.75
  }
}
```

**B. Exchange Data (HSX/HNX/UPCOM):**
- Dữ liệu giao dịch real-time từ HSX API
- Thông tin công bố thông tin từ HNX
- Báo cáo tài chính từ UPCOM

### 1.2 Dữ Liệu Số (Market Data APIs)

**A. Financial Data Providers:**
```python
# Ví dụ integration với vnstock
import vnstock
from vnstock import stock_historical_data

def get_technical_data(symbol, start_date, end_date):
    """
    Lấy dữ liệu lịch sử giá cổ phiếu
    """
    data = stock_historical_data(
        symbol=symbol,
        start_date=start_date,
        end_date=end_date,
        resolution='1D'
    )
    return data

# Hoặc sử dụng FiinGroup API (trả phí)
FIIN_API_CONFIG = {
    'endpoint': 'https://api.fiingroup.vn/StockInfo',
    'headers': {
        'Authorization': 'Bearer YOUR_TOKEN',
        'Content-Type': 'application/json'
    }
}
```

**B. Các chỉ số cần thu thập:**
- OHLC (Open, High, Low, Close)
- Volume (Khối lượng giao dịch)
- P/E, P/B, EPS, ROE, ROA
- Market Cap (Vốn hóa)
- Foreign Ownership (Room ngoại)
- Insider Trading Data

### 1.3 Nguồn Cộng Đồng (Social Sentiment)

**A. Facebook Groups:**
```javascript
// Sử dụng Apify hoặc Bright Data để scrape
const fbGroups = [
  'Chứng khoán Việt Nam',
  'Đầu tư chứng khoán thông minh',
  'Phân tích kỹ thuật chứng khoán VN'
]

// Scraping flow
async function scrapeFacebookGroup(groupId) {
  // Sử dụng Apify Actor
  const run = await apifyClient.actor("apify/facebook-pages-scraper").call({
    startUrls: [`https://www.facebook.com/groups/${groupId}`],
    maxPosts: 50
  });
  
  return run.dataset.items;
}
```

**B. Telegram Channels:**
```python
from telethon import TelegramClient

# Channels phổ biến
TELEGRAM_CHANNELS = [
    '@chungkhoanvietnam',
    '@stockvietnam',
    '@vnstock_analysis'
]

async def scrape_telegram_messages(channel, limit=100):
    """
    Lấy tin nhắn từ Telegram channel
    """
    client = TelegramClient('session', API_ID, API_HASH)
    await client.start()
    
    messages = []
    async for message in client.iter_messages(channel, limit=limit):
        messages.append({
            'text': message.text,
            'date': message.date,
            'views': message.views
        })
    
    return messages
```

**C. Zalo Groups (Khó khăn hơn):**
- Yêu cầu Zalo API access (hạn chế)
- Có thể cần manual monitoring hoặc bot member
- Alternative: Sử dụng người dùng thật forward tin quan trọng

---

## 2. n8n WORKFLOW - Quy Trình Tự Động

### 2.1 Main Orchestration Workflow

```json
{
  "name": "Daily Stock Market Analysis",
  "nodes": [
    {
      "type": "n8n-nodes-base.cron",
      "name": "Daily Trigger",
      "parameters": {
        "triggerTimes": {
          "hour": 8,
          "minute": 0
        }
      }
    },
    {
      "type": "n8n-nodes-base.httpRequest",
      "name": "Scrape RSS Feeds",
      "parameters": {
        "url": "={{$node['Get Sources'].json['rss_url']}}",
        "method": "GET"
      }
    },
    {
      "type": "n8n-nodes-base.code",
      "name": "Parse & Filter News",
      "parameters": {
        "jsCode": `
          // Lọc bỏ tin rác, spam "lùa gà"
          const spamKeywords = ['khuyến mại', 'đăng ký ngay', 'cơ hội vàng'];
          const items = $input.all();
          
          return items.filter(item => {
            const text = item.json.title + ' ' + item.json.description;
            return !spamKeywords.some(keyword => 
              text.toLowerCase().includes(keyword)
            );
          });
        `
      }
    },
    {
      "type": "n8n-nodes-base.aws",
      "name": "Save to S3 Raw",
      "parameters": {
        "bucket": "vnstock-data",
        "key": "raw-data/news/{{$now.format('YYYY-MM-DD')}}.json",
        "data": "={{$json}}"
      }
    },
    {
      "type": "n8n-nodes-base.httpRequest",
      "name": "Call Sentiment Agent",
      "parameters": {
        "url": "http://agent-service:8000/analyze/sentiment",
        "method": "POST",
        "body": "={{$json}}"
      }
    },
    {
      "type": "n8n-nodes-base.httpRequest",
      "name": "Get Market Data",
      "parameters": {
        "url": "https://api.vietstock.vn/finance/stockprice",
        "authentication": "genericCredentialType"
      }
    },
    {
      "type": "n8n-nodes-base.httpRequest",
      "name": "Call Technical Agent",
      "parameters": {
        "url": "http://agent-service:8000/analyze/technical",
        "method": "POST"
      }
    },
    {
      "type": "n8n-nodes-base.httpRequest",
      "name": "Master Agent Synthesis",
      "parameters": {
        "url": "http://agent-service:8000/synthesize",
        "method": "POST",
        "body": {
          "sentiment": "={{$node['Call Sentiment Agent'].json}}",
          "technical": "={{$node['Call Technical Agent'].json}}"
        }
      }
    },
    {
      "type": "n8n-nodes-base.aws",
      "name": "Save Final Report",
      "parameters": {
        "bucket": "vnstock-data",
        "key": "reports/daily/{{$now.format('YYYY-MM-DD')}}.json"
      }
    },
    {
      "type": "n8n-nodes-base.telegram",
      "name": "Send to Telegram",
      "parameters": {
        "chatId": "@vnstock_alerts",
        "text": "={{$json['summary']}}"
      }
    }
  ]
}
```

### 2.2 Hot Stocks Detection Workflow

```javascript
// Workflow phát hiện cổ phiếu "hot" trong ngày
async function detectHotStocks(newsData, socialData, marketData) {
  const stockMentions = {};
  
  // Đếm số lần xuất hiện mỗi mã
  [...newsData, ...socialData].forEach(item => {
    const symbols = extractStockSymbols(item.text);
    symbols.forEach(symbol => {
      stockMentions[symbol] = (stockMentions[symbol] || 0) + 1;
    });
  });
  
  // Lọc top 10 mã được nhắc đến nhiều nhất
  const hotStocks = Object.entries(stockMentions)
    .sort((a, b) => b[1] - a[1])
    .slice(0, 10)
    .map(([symbol, count]) => ({
      symbol,
      mentions: count,
      priceChange: marketData[symbol]?.priceChange || 0,
      volume: marketData[symbol]?.volume || 0
    }));
  
  return hotStocks;
}

function extractStockSymbols(text) {
  // Regex để tìm mã chứng khoán VN (3 chữ cái in hoa)
  const regex = /\b[A-Z]{3}\b/g;
  const matches = text.match(regex) || [];
  
  // Filter ra các mã thật (so với danh sách mã niêm yết)
  return matches.filter(symbol => VALID_SYMBOLS.includes(symbol));
}
```

---

## 3. AI AGENT ARCHITECTURE - Kiến Trúc Agents

### 3.1 Technical Analysis Agent

```python
import pandas as pd
import ta

class TechnicalAgent:
    """
    Agent phân tích kỹ thuật
    """
    
    def __init__(self, data: pd.DataFrame):
        self.data = data
        self.indicators = {}
    
    def calculate_indicators(self):
        """
        Tính toán các chỉ báo kỹ thuật
        """
        df = self.data
        
        # RSI (Relative Strength Index)
        self.indicators['rsi'] = ta.momentum.RSIIndicator(
            close=df['close'], 
            window=14
        ).rsi()
        
        # MACD
        macd = ta.trend.MACD(close=df['close'])
        self.indicators['macd'] = macd.macd()
        self.indicators['macd_signal'] = macd.macd_signal()
        
        # Bollinger Bands
        bb = ta.volatility.BollingerBands(close=df['close'])
        self.indicators['bb_upper'] = bb.bollinger_hband()
        self.indicators['bb_lower'] = bb.bollinger_lband()
        self.indicators['bb_middle'] = bb.bollinger_mavg()
        
        # Moving Averages
        self.indicators['sma_20'] = ta.trend.SMAIndicator(
            close=df['close'], 
            window=20
        ).sma_indicator()
        
        self.indicators['ema_50'] = ta.trend.EMAIndicator(
            close=df['close'],
            window=50
        ).ema_indicator()
        
        return self.indicators
    
    def generate_signals(self):
        """
        Tạo tín hiệu mua/bán dựa trên chỉ báo
        """
        signals = {
            'recommendation': 'HOLD',
            'confidence': 0.5,
            'reasons': []
        }
        
        current_rsi = self.indicators['rsi'].iloc[-1]
        current_price = self.data['close'].iloc[-1]
        sma_20 = self.indicators['sma_20'].iloc[-1]
        
        # RSI signals
        if current_rsi < 30:
            signals['recommendation'] = 'BUY'
            signals['confidence'] += 0.2
            signals['reasons'].append('RSI oversold (<30)')
        elif current_rsi > 70:
            signals['recommendation'] = 'SELL'
            signals['confidence'] += 0.2
            signals['reasons'].append('RSI overbought (>70)')
        
        # Price vs SMA
        if current_price > sma_20:
            signals['confidence'] += 0.1
            signals['reasons'].append('Price above SMA20 (bullish)')
        else:
            signals['confidence'] -= 0.1
            signals['reasons'].append('Price below SMA20 (bearish)')
        
        # MACD crossover
        macd_current = self.indicators['macd'].iloc[-1]
        macd_signal = self.indicators['macd_signal'].iloc[-1]
        
        if macd_current > macd_signal:
            signals['confidence'] += 0.15
            signals['reasons'].append('MACD bullish crossover')
        
        return signals
```

### 3.2 Sentiment Analysis Agent

```python
from transformers import AutoTokenizer, AutoModelForSequenceClassification
import torch

class SentimentAgent:
    """
    Agent phân tích tâm lý thị trường từ tin tức và MXH
    """
    
    def __init__(self, model_name='vinai/phobert-base'):
        self.tokenizer = AutoTokenizer.from_pretrained(model_name)
        self.model = AutoModelForSequenceClassification.from_pretrained(
            model_name,
            num_labels=3  # Positive, Neutral, Negative
        )
        
        # Từ điển từ lóng chứng khoán VN
        self.slang_dict = {
            'cây thông': 'bullish_pattern',
            'múa bên trăng': 'price_manipulation',
            'fomo': 'fear_of_missing_out',
            'sideway': 'sideways_trend',
            'breakout': 'price_breakout',
            'hốt': 'buy_opportunity',
            'chốt lời': 'take_profit',
            'cắt lỗ': 'stop_loss',
            'lùa gà': 'pump_and_dump',
            'con tép': 'small_investor'
        }
    
    def preprocess_text(self, text: str) -> str:
        """
        Chuẩn hóa text, thay thế slang
        """
        text_lower = text.lower()
        
        for slang, meaning in self.slang_dict.items():
            text_lower = text_lower.replace(slang, f' {meaning} ')
        
        return text_lower
    
    def analyze_sentiment(self, text: str) -> dict:
        """
        Phân tích cảm xúc của một đoạn text
        """
        processed_text = self.preprocess_text(text)
        inputs = self.tokenizer(
            processed_text,
            return_tensors='pt',
            truncation=True,
            max_length=256
        )
        
        with torch.no_grad():
            outputs = self.model(**inputs)
            logits = outputs.logits
            probabilities = torch.softmax(logits, dim=1)
        
        sentiment_map = {0: 'negative', 1: 'neutral', 2: 'positive'}
        predicted_class = torch.argmax(probabilities).item()
        
        return {
            'sentiment': sentiment_map[predicted_class],
            'confidence': probabilities[0][predicted_class].item(),
            'scores': {
                'negative': probabilities[0][0].item(),
                'neutral': probabilities[0][1].item(),
                'positive': probabilities[0][2].item()
            }
        }
    
    def analyze_batch(self, texts: list) -> dict:
        """
        Phân tích sentiment cho nhiều text (tin tức, bài đăng MXH)
        """
        results = [self.analyze_sentiment(text) for text in texts]
        
        # Tính điểm trung bình
        avg_sentiment = {
            'positive_ratio': sum(1 for r in results if r['sentiment'] == 'positive') / len(results),
            'negative_ratio': sum(1 for r in results if r['sentiment'] == 'negative') / len(results),
            'neutral_ratio': sum(1 for r in results if r['sentiment'] == 'neutral') / len(results),
            'overall_score': sum(r['scores']['positive'] - r['scores']['negative'] for r in results) / len(results)
        }
        
        return {
            'individual_results': results,
            'aggregate': avg_sentiment
        }
    
    def detect_rumors(self, social_texts: list, official_texts: list) -> dict:
        """
        So sánh tin từ MXH với tin chính thống để phát hiện tin đồn
        """
        # Extract entities (stock symbols) từ social media
        social_entities = self._extract_entities(social_texts)
        official_entities = self._extract_entities(official_texts)
        
        # Tìm các mã chỉ xuất hiện trên MXH (nghi ngờ tin đồn)
        rumor_candidates = set(social_entities.keys()) - set(official_entities.keys())
        
        rumors = []
        for symbol in rumor_candidates:
            rumors.append({
                'symbol': symbol,
                'mentions': social_entities[symbol],
                'risk_level': 'HIGH' if social_entities[symbol] > 10 else 'MEDIUM',
                'warning': 'Chưa có xác nhận từ nguồn chính thống'
            })
        
        return {
            'detected_rumors': rumors,
            'verified_news': list(official_entities.keys())
        }
    
    def _extract_entities(self, texts: list) -> dict:
        """
        Trích xuất mã chứng khoán từ text
        """
        import re
        entities = {}
        
        for text in texts:
            symbols = re.findall(r'\b[A-Z]{3}\b', text)
            for symbol in symbols:
                entities[symbol] = entities.get(symbol, 0) + 1
        
        return entities
```

### 3.3 Forecast Agent

```python
import numpy as np
from sklearn.ensemble import RandomForestClassifier

class ForecastAgent:
    """
    Agent dự báo xu hướng tổng hợp
    """
    
    def __init__(self):
        self.model = RandomForestClassifier(n_estimators=100)
        self.is_trained = False
    
    def synthesize_analysis(
        self, 
        technical_signals: dict, 
        sentiment_data: dict,
        market_context: dict
    ) -> dict:
        """
        Tổng hợp kết quả từ Technical Agent và Sentiment Agent
        """
        # Tính điểm tổng hợp
        technical_score = self._score_technical(technical_signals)
        sentiment_score = self._score_sentiment(sentiment_data)
        market_score = self._score_market(market_context)
        
        # Weighted average
        final_score = (
            technical_score * 0.4 +
            sentiment_score * 0.3 +
            market_score * 0.3
        )
        
        # Đưa ra khuyến nghị
        recommendation = self._generate_recommendation(
            final_score,
            technical_signals,
            sentiment_data
        )
        
        return {
            'recommendation': recommendation['action'],
            'confidence': recommendation['confidence'],
            'risk_level': self._calculate_risk(technical_signals, sentiment_data),
            'target_price': self._estimate_target_price(technical_signals),
            'reasoning': recommendation['reasons'],
            'scores': {
                'technical': technical_score,
                'sentiment': sentiment_score,
                'market': market_score,
                'final': final_score
            }
        }
    
    def _score_technical(self, signals: dict) -> float:
        """Chuyển technical signals thành điểm số 0-1"""
        score = 0.5  # baseline
        
        if signals['recommendation'] == 'BUY':
            score += signals['confidence'] * 0.5
        elif signals['recommendation'] == 'SELL':
            score -= signals['confidence'] * 0.5
        
        return max(0, min(1, score))
    
    def _score_sentiment(self, data: dict) -> float:
        """Chuyển sentiment data thành điểm số 0-1"""
        if 'aggregate' in data:
            return (data['aggregate']['overall_score'] + 1) / 2  # Normalize từ [-1,1] sang [0,1]
        return 0.5
    
    def _score_market(self, context: dict) -> float:
        """Đánh giá bối cảnh thị trường chung"""
        score = 0.5
        
        # VN-Index trend
        if context.get('vnindex_change', 0) > 0:
            score += 0.2
        elif context.get('vnindex_change', 0) < -1:
            score -= 0.2
        
        # Volume
        if context.get('volume_ratio', 1) > 1.2:  # Volume tăng 20%
            score += 0.1
        
        # Foreign flow
        if context.get('foreign_net_value', 0) > 0:
            score += 0.15
        
        return max(0, min(1, score))
    
    def _generate_recommendation(
        self, 
        final_score: float,
        technical: dict,
        sentiment: dict
    ) -> dict:
        """
        Tạo khuyến nghị cuối cùng
        """
        reasons = []
        
        if final_score >= 0.7:
            action = 'STRONG BUY'
            confidence = final_score
            reasons.append('Tín hiệu kỹ thuật tích cực')
            reasons.append('Sentiment thị trường lạc quan')
        elif final_score >= 0.55:
            action = 'BUY'
            confidence = final_score
            reasons.append('Xu hướng tăng ngắn hạn')
        elif final_score >= 0.45:
            action = 'HOLD'
            confidence = 0.6
            reasons.append('Thị trường sideway, chờ tín hiệu rõ hơn')
        elif final_score >= 0.3:
            action = 'SELL'
            confidence = 1 - final_score
            reasons.append('Áp lực bán tăng')
        else:
            action = 'STRONG SELL'
            confidence = 1 - final_score
            reasons.append('Tín hiệu kỹ thuật tiêu cực mạnh')
        
        # Thêm cảnh báo rủi ro
        if sentiment.get('detected_rumors'):
            reasons.append('⚠️ CẢNH BÁO: Phát hiện tin đồn chưa xác minh')
        
        return {
            'action': action,
            'confidence': confidence,
            'reasons': reasons
        }
    
    def _calculate_risk(self, technical: dict, sentiment: dict) -> str:
        """Đánh giá mức độ rủi ro"""
        risk_score = 0
        
        # Volatility cao = rủi ro cao
        if technical.get('volatility', 0) > 3:
            risk_score += 2
        
        # Sentiment tiêu cực
        if sentiment.get('aggregate', {}).get('negative_ratio', 0) > 0.5:
            risk_score += 2
        
        # Có tin đồn
        if sentiment.get('detected_rumors'):
            risk_score += 3
        
        if risk_score >= 5:
            return 'HIGH'
        elif risk_score >= 3:
            return 'MEDIUM'
        else:
            return 'LOW'
    
    def _estimate_target_price(self, technical: dict) -> dict:
        """Ước tính giá mục tiêu"""
        current_price = technical.get('current_price', 0)
        bb_upper = technical.get('bb_upper', current_price * 1.05)
        bb_lower = technical.get('bb_lower', current_price * 0.95)
        
        return {
            'short_term_target': bb_upper,
            'support_level': bb_lower,
            'expected_return': ((bb_upper - current_price) / current_price) * 100
        }
```

### 3.4 Master Orchestrator Agent

```python
class MasterAgent:
    """
    Agent tổng chỉ huy, điều phối các agents khác
    """
    
    def __init__(self):
        self.technical_agent = TechnicalAgent
        self.sentiment_agent = SentimentAgent()
        self.forecast_agent = ForecastAgent()
    
    async def analyze_stock(self, symbol: str, date: str) -> dict:
        """
        Phân tích toàn diện một mã cổ phiếu
        """
        # 1. Thu thập dữ liệu
        market_data = await self._fetch_market_data(symbol, date)
        news_data = await self._fetch_news_data(symbol, date)
        social_data = await self._fetch_social_data(symbol, date)
        
        # 2. Phân tích kỹ thuật
        tech_agent = self.technical_agent(market_data)
        tech_agent.calculate_indicators()
        technical_signals = tech_agent.generate_signals()
        
        # 3. Phân tích tâm lý
        sentiment_results = self.sentiment_agent.analyze_batch(
            news_data + social_data
        )
        
        # 4. Phát hiện tin đồn
        rumor_check = self.sentiment_agent.detect_rumors(
            social_data, 
            news_data
        )
        sentiment_results['detected_rumors'] = rumor_check['detected_rumors']
        
        # 5. Dự báo tổng hợp
        market_context = await self._get_market_context()
        forecast = self.forecast_agent.synthesize_analysis(
            technical_signals,
            sentiment_results,
            market_context
        )
        
        # 6. Tổng hợp báo cáo
        report = {
            'symbol': symbol,
            'date': date,
            'analysis': {
                'technical': technical_signals,
                'sentiment': sentiment_results,
                'forecast': forecast
            },
            'recommendation': forecast['recommendation'],
            'confidence': forecast['confidence'],
            'risk_level': forecast['risk_level'],
            'key_insights': self._extract_insights(
                technical_signals, 
                sentiment_results, 
                forecast
            )
        }
        
        return report
    
    async def generate_daily_report(self, date: str) -> dict:
        """
        Tạo báo cáo tổng hợp thị trường hàng ngày
        """
        # Phát hiện hot stocks
        hot_stocks = await self._detect_hot_stocks(date)
        
        # Phân tích từng mã hot
        analyses = []
        for stock in hot_stocks[:5]:  # Top 5
            analysis = await self.analyze_stock(stock['symbol'], date)
            analyses.append(analysis)
        
        # Market overview
        market_overview = await self._get_market_overview(date)
        
        # Tạo báo cáo
        report = {
            'date': date,
            'market_overview': market_overview,
            'hot_stocks': hot_stocks,
            'top_recommendations': self._rank_recommendations(analyses),
            'detailed_analyses': analyses,
            'alerts': self._generate_alerts(analyses)
        }
        
        return report
    
    def _extract_insights(self, technical, sentiment, forecast) -> list:
        """Trích xuất các insight quan trọng"""
        insights = []
        
        # Technical insights
        if 'RSI oversold' in str(technical.get('reasons', [])):
            insights.append('📊 RSI cho thấy cổ phiếu đang ở vùng oversold')
        
        # Sentiment insights
        if sentiment.get('detected_rumors'):
            insights.append('⚠️ Phát hiện tin đồn trên MXH, cần thận trọng')
        
        # Forecast insights
        if forecast['confidence'] > 0.8:
            insights.append(f"✅ Độ tin cậy cao ({forecast['confidence']:.1%})")
        
        return insights
    
    def _rank_recommendations(self, analyses: list) -> list:
        """Xếp hạng các khuyến nghị"""
        ranked = sorted(
            analyses,
            key=lambda x: x['confidence'],
            reverse=True
        )
        
        return [
            {
                'symbol': a['symbol'],
                'recommendation': a['recommendation'],
                'confidence': a['confidence'],
                'key_reason': a['key_insights'][0] if a['key_insights'] else ''
            }
            for a in ranked
        ]
    
    def _generate_alerts(self, analyses: list) -> list:
        """Tạo cảnh báo quan trọng"""
        alerts = []
        
        for analysis in analyses:
            # High risk alert
            if analysis['risk_level'] == 'HIGH':
                alerts.append({
                    'type': 'HIGH_RISK',
                    'symbol': analysis['symbol'],
                    'message': f"⚠️ {analysis['symbol']}: Mức rủi ro cao!"
                })
            
            # Strong buy/sell
            if analysis['recommendation'] in ['STRONG BUY', 'STRONG SELL']:
                alerts.append({
                    'type': 'STRONG_SIGNAL',
                    'symbol': analysis['symbol'],
                    'message': f"🔔 {analysis['symbol']}: {analysis['recommendation']}"
                })
            
            # Rumor detection
            if analysis['analysis']['sentiment'].get('detected_rumors'):
                alerts.append({
                    'type': 'RUMOR',
                    'symbol': analysis['symbol'],
                    'message': f"🚨 {analysis['symbol']}: Phát hiện tin đồn"
                })
        
        return alerts
```

---

## 4. OUTPUT & DISTRIBUTION - Xuất Báo Cáo

### 4.1 Report Generator

```python
from datetime import datetime
import json

class ReportGenerator:
    """
    Tạo báo cáo dưới nhiều định dạng
    """
    
    def generate_markdown_report(self, data: dict) -> str:
        """
        Tạo báo cáo Markdown
        """
        md = f"""# Báo Cáo Phân Tích Thị Trường Chứng Khoán
**Ngày:** {data['date']}

## 📊 Tổng Quan Thị Trường

- **VN-Index:** {data['market_overview']['vnindex']} ({data['market_overview']['vnindex_change']:+.2f}%)
- **Thanh khoản:** {data['market_overview']['total_volume']} tỷ đồng
- **Khối ngoại:** {data['market_overview']['foreign_net_value']:+.0f} tỷ đồng

---

## 🔥 Top Cổ Phiếu Hot Trong Ngày

"""
        for stock in data['hot_stocks'][:5]:
            md += f"### {stock['symbol']}\n"
            md += f"- **Mentions:** {stock['mentions']}\n"
            md += f"- **Giá:** {stock['price']:,.0f} VND ({stock['price_change']:+.2f}%)\n"
            md += f"- **Khuyến nghị:** {stock['recommendation']}\n"
            md += f"- **Độ tin cậy:** {stock['confidence']:.1%}\n\n"
        
        md += "---\n\n## 🎯 Phân Tích Chi Tiết\n\n"
        
        for analysis in data['detailed_analyses']:
            md += self._format_stock_analysis(analysis)
        
        md += "---\n\n## ⚠️ Cảnh Báo\n\n"
        for alert in data['alerts']:
            md += f"- {alert['message']}\n"
        
        md += f"\n---\n\n*Báo cáo được tạo tự động bởi AI Agent System*\n"
        md += f"*Lưu ý: Đây chỉ là thông tin tham khảo, không phải lời khuyên đầu tư*\n"
        
        return md
    
    def _format_stock_analysis(self, analysis: dict) -> str:
        """Format phân tích một mã"""
        md = f"### {analysis['symbol']}\n\n"
        
        md += f"**Khuyến nghị:** {analysis['recommendation']} "
        md += f"(Confidence: {analysis['confidence']:.1%})\n\n"
        
        md += f"**Mức rủi ro:** {analysis['risk_level']}\n\n"
        
        md += "**Key Insights:**\n"
        for insight in analysis['key_insights']:
            md += f"- {insight}\n"
        
        md += "\n"
        
        return md
    
    def generate_json_report(self, data: dict) -> str:
        """Xuất báo cáo JSON cho API"""
        return json.dumps(data, ensure_ascii=False, indent=2)
    
    def generate_html_report(self, data: dict) -> str:
        """Tạo báo cáo HTML đẹp"""
        html = f"""
<!DOCTYPE html>
<html lang="vi">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Báo Cáo Phân Tích Thị Trường - {data['date']}</title>
    <style>
        body {{
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
            background: #f5f5f5;
        }}
        .header {{
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 30px;
            border-radius: 10px;
            margin-bottom: 30px;
        }}
        .stock-card {{
            background: white;
            padding: 20px;
            margin: 15px 0;
            border-radius: 8px;
            box-shadow: 0 2px 5px rgba(0,0,0,0.1);
        }}
        .recommendation {{
            display: inline-block;
            padding: 5px 15px;
            border-radius: 20px;
            font-weight: bold;
        }}
        .buy {{ background: #4caf50; color: white; }}
        .sell {{ background: #f44336; color: white; }}
        .hold {{ background: #ff9800; color: white; }}
        .risk-high {{ color: #f44336; }}
        .risk-medium {{ color: #ff9800; }}
        .risk-low {{ color: #4caf50; }}
    </style>
</head>
<body>
    <div class="header">
        <h1>📊 Báo Cáo Phân Tích Thị Trường</h1>
        <p>Ngày: {data['date']}</p>
    </div>
    
    <div class="market-overview">
        <h2>Tổng Quan Thị Trường</h2>
        <p>VN-Index: <strong>{data['market_overview']['vnindex']}</strong> 
           ({data['market_overview']['vnindex_change']:+.2f}%)</p>
    </div>
    
    <h2>🔥 Top Cổ Phiếu</h2>
"""
        
        for analysis in data['detailed_analyses']:
            rec_class = analysis['recommendation'].lower().replace(' ', '-')
            risk_class = f"risk-{analysis['risk_level'].lower()}"
            
            html += f"""
    <div class="stock-card">
        <h3>{analysis['symbol']}</h3>
        <span class="recommendation {rec_class}">{analysis['recommendation']}</span>
        <p>Độ tin cậy: {analysis['confidence']:.1%}</p>
        <p class="{risk_class}">Rủi ro: {analysis['risk_level']}</p>
        <ul>
"""
            for insight in analysis['key_insights']:
                html += f"            <li>{insight}</li>\n"
            
            html += """        </ul>
    </div>
"""
        
        html += """
    <footer style="text-align: center; margin-top: 50px; color: #666;">
        <p><em>Báo cáo được tạo tự động bởi AI Agent System</em></p>
        <p><small>Lưu ý: Đây chỉ là thông tin tham khảo, không phải lời khuyên đầu tư</small></p>
    </footer>
</body>
</html>
"""
        return html
```

### 4.2 Telegram Bot Integration

```python
from telegram import Bot, Update
from telegram.ext import Application, CommandHandler, ContextTypes

class TelegramDistributor:
    """
    Gửi cảnh báo và báo cáo qua Telegram
    """
    
    def __init__(self, bot_token: str):
        self.bot = Bot(token=bot_token)
        self.app = Application.builder().token(bot_token).build()
        
        # Register handlers
        self.app.add_handler(CommandHandler("start", self.start_command))
        self.app.add_handler(CommandHandler("report", self.daily_report))
        self.app.add_handler(CommandHandler("alert", self.subscribe_alerts))
    
    async def start_command(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Welcome message"""
        await update.message.reply_text(
            "🤖 Chào mừng đến với VN Stock Analysis Bot!\n\n"
            "Commands:\n"
            "/report - Nhận báo cáo hàng ngày\n"
            "/alert - Đăng ký nhận cảnh báo real-time\n"
            "/stock <symbol> - Phân tích mã cụ thể"
        )
    
    async def send_daily_report(self, chat_id: str, report_data: dict):
        """Gửi báo cáo hàng ngày"""
        message = self._format_telegram_message(report_data)
        await self.bot.send_message(chat_id=chat_id, text=message, parse_mode='Markdown')
    
    async def send_alert(self, chat_id: str, alert: dict):
        """Gửi cảnh báo real-time"""
        message = f"{alert['message']}\n\n"
        message += f"Symbol: `{alert['symbol']}`\n"
        message += f"Type: {alert['type']}"
        
        await self.bot.send_message(chat_id=chat_id, text=message, parse_mode='Markdown')
    
    def _format_telegram_message(self, data: dict) -> str:
        """Format message cho Telegram"""
        msg = f"📊 *Báo Cáo Thị Trường - {data['date']}*\n\n"
        
        msg += f"*VN-Index:* {data['market_overview']['vnindex']} "
        msg += f"({data['market_overview']['vnindex_change']:+.2f}%)\n\n"
        
        msg += "*🔥 Top Khuyến Nghị:*\n"
        for rec in data['top_recommendations'][:3]:
            msg += f"• `{rec['symbol']}` - {rec['recommendation']} "
            msg += f"({rec['confidence']:.0%})\n"
        
        msg += "\n_Chi tiết: /report_"
        
        return msg
```

---

## 5. DEPLOYMENT & INFRASTRUCTURE

### 5.1 Docker Compose Setup

```yaml
version: '3.8'

services:
  n8n:
    image: n8nio/n8n
    container_name: vnstock-n8n
    ports:
      - "5678:5678"
    environment:
      - N8N_BASIC_AUTH_ACTIVE=true
      - N8N_BASIC_AUTH_USER=admin
      - N8N_BASIC_AUTH_PASSWORD=${N8N_PASSWORD}
      - WEBHOOK_URL=https://your-domain.com/
    volumes:
      - n8n_data:/home/node/.n8n
    networks:
      - vnstock-net

  agent-service:
    build: ./agent-service
    container_name: vnstock-agents
    ports:
      - "8000:8000"
    environment:
      - AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY}
      - AWS_SECRET_ACCESS_KEY=${AWS_SECRET_KEY}
      - S3_BUCKET=vnstock-data
    volumes:
      - ./models:/app/models
    networks:
      - vnstock-net
    depends_on:
      - postgres
      - redis

  postgres:
    image: postgres:15
    container_name: vnstock-db
    environment:
      - POSTGRES_DB=vnstock
      - POSTGRES_USER=admin
      - POSTGRES_PASSWORD=${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - vnstock-net

  redis:
    image: redis:7-alpine
    container_name: vnstock-cache
    ports:
      - "6379:6379"
    networks:
      - vnstock-net

  web-dashboard:
    build: ./web-dashboard
    container_name: vnstock-web
    ports:
      - "3000:3000"
    environment:
      - API_URL=http://agent-service:8000
    networks:
      - vnstock-net

volumes:
  n8n_data:
  postgres_data:

networks:
  vnstock-net:
    driver: bridge
```

### 5.2 AWS S3 Bucket Structure

```
vnstock-data/
├── raw-data/
│   ├── news/
│   │   ├── 2026-01-28/
│   │   │   ├── vneconomy.json
│   │   │   ├── cafef.json
│   │   │   └── vietstock.json
│   │   └── ...
│   ├── market-data/
│   │   ├── 2026-01-28/
│   │   │   ├── VNM.json
│   │   │   ├── HPG.json
│   │   │   └── ...
│   │   └── ...
│   └── social/
│       ├── 2026-01-28/
│       │   ├── facebook.json
│       │   ├── telegram.json
│       │   └── ...
│       └── ...
├── processed/
│   ├── sentiment/
│   ├── technical/
│   └── combined/
└── reports/
    ├── daily/
    │   ├── 2026-01-28.json
    │   ├── 2026-01-28.md
    │   └── 2026-01-28.html
    ├── weekly/
    └── alerts/
```

---

## 6. LƯU Ý QUAN TRỌNG & BEST PRACTICES

### 6.1 Đặc Thù Ngôn Ngữ Việt

```python
# Từ điển slang chứng khoán VN (mở rộng)
STOCK_SLANG_VIETNAMESE = {
    # Mô tả xu hướng
    'cây thông': 'Mô hình nến tăng mạnh',
    'cây súng': 'Mô hình nến giảm mạnh',
    'múa bên trăng': 'Thao túng giá',
    'sideway': 'Di chuyển ngang',
    'breakout': 'Vượt ngưỡng kháng cự',
    'breakdown': 'Thủng ngưỡng hỗ trợ',
    
    # Hành vi nhà đầu tư
    'fomo': 'Sợ bỏ lỡ cơ hội',
    'panic sell': 'Bán tháo hoảng loạn',
    'hốt': 'Mua vào',
    'ôm': 'Giữ cổ phiếu lâu dài',
    'chốt lời': 'Bán để chốt lãi',
    'cắt lỗ': 'Bán chấp nhận lỗ',
    'all in': 'Đầu tư toàn bộ vốn',
    
    # Loại nhà đầu tư
    'con tép': 'Nhà đầu tư nhỏ lẻ',
    'cá mập': 'Nhà đầu tư lớn',
    'lùa gà': 'Thao túng để bán cho nhà đầu tư nhỏ',
    'gà mờ': 'Nhà đầu tư thiếu kinh nghiệm',
    
    # Thuật ngữ kỹ thuật (Viet-style)
    'lệch pha': 'Divergence',
    'vùng kháng cự': 'Resistance zone',
    'vùng hỗ trợ': 'Support zone',
    'chạm đáy': 'Bottom hit',
    'đỉnh': 'Peak/Top'
}
```

### 6.2 Các Lưu Ý Pháp Lý

```python
# Disclaimer template
LEGAL_DISCLAIMER = """
⚠️ LƯU Ý QUAN TRỌNG:

1. Báo cáo này được tạo tự động bởi hệ thống AI và chỉ mang tính chất THAM KHẢO.

2. ĐÂY KHÔNG PHẢI LÀ LỜI KHUYÊN ĐẦU TƯ. Nhà đầu tư cần tự nghiên cứu và 
   chịu trách nhiệm cho quyết định đầu tư của mình.

3. Dữ liệu được thu thập từ nhiều nguồn công khai nhưng có thể không đầy đủ 
   hoặc chính xác 100%.

4. Thị trường chứng khoán có rủi ro cao. Chỉ đầu tư số tiền bạn có thể chấp 
   nhận mất.

5. Hệ thống không chịu trách nhiệm cho bất kỳ tổn thất tài chính nào phát 
   sinh từ việc sử dụng thông tin này.

📧 Liên hệ: support@vnstock-analysis.com
"""

def add_disclaimer_to_report(report: str) -> str:
    """Thêm disclaimer vào mọi báo cáo"""
    return report + "\n\n" + LEGAL_DISCLAIMER
```

### 6.3 Rate Limiting & API Quotas

```python
from time import sleep
import redis

class RateLimiter:
    """
    Giới hạn số lượng requests để tránh bị block
    """
    
    def __init__(self, redis_client):
        self.redis = redis_client
    
    def check_limit(self, key: str, max_requests: int, window: int) -> bool:
        """
        Check if request is allowed
        
        Args:
            key: Identifier (e.g., 'vietstock_api')
            max_requests: Max requests allowed
            window: Time window in seconds
        """
        current = self.redis.get(key)
        
        if current is None:
            self.redis.setex(key, window, 1)
            return True
        
        if int(current) < max_requests:
            self.redis.incr(key)
            return True
        
        return False
    
    def wait_if_needed(self, key: str, max_requests: int, window: int):
        """Wait if rate limit exceeded"""
        while not self.check_limit(key, max_requests, window):
            print(f"Rate limit exceeded for {key}, waiting...")
            sleep(5)

# Usage
rate_limiter = RateLimiter(redis.Redis())

# Khi gọi API
rate_limiter.wait_if_needed('vietstock_api', max_requests=100, window=3600)
data = fetch_from_vietstock_api()
```

### 6.4 Error Handling & Retry Logic

```python
import backoff
from requests.exceptions import RequestException

class RobustScraper:
    """
    Scraper với retry logic và error handling mạnh mẽ
    """
    
    @backoff.on_exception(
        backoff.expo,
        RequestException,
        max_tries=5,
        max_time=300
    )
    def fetch_with_retry(self, url: str) -> dict:
        """
        Fetch data với exponential backoff retry
        """
        try:
            response = requests.get(url, timeout=10)
            response.raise_for_status()
            return response.json()
        except RequestException as e:
            print(f"Error fetching {url}: {e}")
            raise
    
    def safe_scrape_multiple(self, urls: list) -> list:
        """
        Scrape nhiều URL, không bị fail toàn bộ nếu 1 URL lỗi
        """
        results = []
        for url in urls:
            try:
                data = self.fetch_with_retry(url)
                results.append({'url': url, 'data': data, 'status': 'success'})
            except Exception as e:
                results.append({'url': url, 'error': str(e), 'status': 'failed'})
                continue
        
        return results
```

### 6.5 Monitoring & Alerting

```python
import logging
from datetime import datetime

class SystemMonitor:
    """
    Giám sát hoạt động của hệ thống
    """
    
    def __init__(self):
        self.logger = logging.getLogger('vnstock_monitor')
        self.metrics = {
            'scrapes_successful': 0,
            'scrapes_failed': 0,
            'agents_executed': 0,
            'reports_generated': 0
        }
    
    def log_scrape_result(self, source: str, success: bool):
        """Log kết quả scraping"""
        if success:
            self.metrics['scrapes_successful'] += 1
            self.logger.info(f"✅ Scraped {source} successfully")
        else:
            self.metrics['scrapes_failed'] += 1
            self.logger.error(f"❌ Failed to scrape {source}")
            
            # Send alert if too many failures
            if self.metrics['scrapes_failed'] > 10:
                self.send_alert("High scrape failure rate!")
    
    def send_alert(self, message: str):
        """Gửi cảnh báo qua Telegram"""
        # Implementation
        pass
    
    def get_health_status(self) -> dict:
        """Trả về trạng thái hệ thống"""
        success_rate = self.metrics['scrapes_successful'] / (
            self.metrics['scrapes_successful'] + self.metrics['scrapes_failed'] + 1
        )
        
        return {
            'status': 'healthy' if success_rate > 0.8 else 'degraded',
            'metrics': self.metrics,
            'success_rate': success_rate,
            'timestamp': datetime.now().isoformat()
        }
```

---

## 7. ROADMAP PHÁT TRIỂN

### Phase 1: MVP (Tháng 1-2)
- [x] Thiết lập infrastructure (n8n + S3 + Agents)
- [ ] Implement scraping cho 3 nguồn chính (VnEconomy, CafeF, VietStock)
- [ ] Xây dựng Technical Agent cơ bản (RSI, MACD, SMA)
- [ ] Xây dựng Sentiment Agent với PhoBERT
- [ ] Tạo daily report Markdown/HTML
- [ ] Telegram bot cơ bản

### Phase 2: Enhancement (Tháng 3-4)
- [ ] Thêm social media scraping (Facebook, Telegram)
- [ ] Nâng cấp Forecast Agent với ML models
- [ ] Implement rumor detection
- [ ] Web dashboard với real-time updates
- [ ] Email notification system
- [ ] Backtesting framework

### Phase 3: Advanced Features (Tháng 5-6)
- [ ] Portfolio tracking & optimization
- [ ] Customizable alerts (price, volume, news)
- [ ] Multi-timeframe analysis
- [ ] Comparative analysis (sector, peers)
- [ ] Mobile app (React Native)
- [ ] Premium subscription model

### Phase 4: Scale & Monetization (Tháng 7+)
- [ ] API for developers
- [ ] White-label solution cho brokers
- [ ] Advanced AI models (GPT-4, Claude)
- [ ] Market maker detection
- [ ] Institutional-grade analytics

---

## 8. KẾT LUẬN

Hệ thống này kết hợp sức mạnh của:
1. **Automation** (n8n workflows)
2. **Data Engineering** (S3, structured storage)
3. **AI/ML** (Multiple specialized agents)
4. **Real-time Distribution** (Telegram, Dashboard)

Để tạo ra một nền tảng phân tích chứng khoán VN toàn diện, giúp nhà đầu tư:
- Theo dõi thị trường 24/7
- Phát hiện cơ hội sớm
- Tránh rủi ro từ tin đồn
- Ra quyết định dựa trên dữ liệu

**Next Steps:**
1. Clone repository này
2. Setup infrastructure (Docker Compose)
3. Configure API keys (VietStock, Telegram, AWS)
4. Deploy n8n workflows
5. Train/fine-tune AI models
6. Launch MVP!

🚀 Happy Building!
