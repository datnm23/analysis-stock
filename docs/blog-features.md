# Blog Features Documentation

**Last Updated**: 2026-04-20

## Overview

Blog site là Next.js 14 frontend application phục vụ nội dung phân tích chứng khoán với real-time market data và AI-powered content.

## Tech Stack

| Component | Technology |
|-----------|------------|
| Framework | Next.js 14 (App Router) |
| Language | TypeScript |
| Styling | Tailwind CSS (Neo-brutalism) |
| Data | Server Components + ISR |
| API | Next.js API Routes |

## Project Structure

```
blog-site/
├── app/
│   ├── api/
│   │   └── subscribe/        # Newsletter subscription API
│   ├── articles/
│   │   ├── page.tsx          # Articles listing (/articles)
│   │   └── [slug]/           # Article detail ([slug])
│   ├── market/
│   │   └── page.tsx          # Market data page (/market)
│   ├── screener/
│   │   └── page.tsx          # Stock screener (/screener)
│   ├── symbols/
│   │   └── [symbol]/         # Symbol detail pages
│   ├── page.tsx              # Homepage (/)
│   └── layout.tsx            # Root layout
├── components/
│   ├── article-card.tsx      # Article preview card
│   ├── markdown-content.tsx   # Markdown renderer
│   ├── market-board-table.tsx # Market price table
│   ├── market-index-bar.tsx   # Index bar (real-time)
│   ├── newsletter-form.tsx    # Email subscription
│   ├── related-articles.tsx  # Related articles widget
│   ├── screener-table.tsx    # Filterable stock table
│   ├── stock-chart.tsx       # Interactive price chart
│   ├── stock-data-widget.tsx  # Symbol data widget
│   └── trending-stocks.tsx   # Trending stocks sidebar
└── lib/
    ├── articles-api.ts       # Article data fetching
    └── sector-mapping.ts     # HSX/HNX sector mapping
```

## Pages & Routes

### `/` - Homepage
- **Components**: `MarketIndexBar`, `MarketBoardTable`, `ArticleCard`, `TrendingStocks`
- **Features**:
  - Real-time market index bar (polls every 60s)
  - Live market price board
  - Latest 4 articles
  - Trending stocks sidebar (7-day, 8 stocks)

### `/articles` - Articles Listing
- **Features**:
  - Grid layout of all articles
  - Search functionality
  - Category filtering
  - Pagination

### `/articles/[slug]` - Article Detail
- **Components**: `MarkdownContent`, `RelatedArticles`
- **Features**:
  - Full article rendering
  - Related articles suggestions
  - Social sharing

### `/market` - Market Data
- **Components**: `MarketBoardTable`, `StockChart`
- **Features**:
  - Real-time HSX/HNX/UPCOM prices
  - Index performance
  - Price change indicators

### `/screener` - Stock Screener
- **Components**: `ScreenerTable`, `StockDataWidget`
- **Features**:
  - Filter by sector
  - Sort by price/volume/change
  - AI recommendations overlay

### `/symbols/[symbol]` - Symbol Detail
- **Components**: `StockChart`, `StockDataWidget`
- **Features**:
  - Individual stock analysis
  - Technical indicators
  - Historical price chart

## Components

### Market Components

| Component | Purpose | Client/Server |
|-----------|---------|----------------|
| `MarketIndexBar` | VN30, HNX30, UPCOM indices | Client (60s polling) |
| `MarketBoardTable` | Live price board | Server (ISR 60s) |

### Content Components

| Component | Purpose |
|-----------|---------|
| `ArticleCard` | Article preview in grid |
| `MarkdownContent` | Render markdown articles |
| `RelatedArticles` | Similar articles sidebar |
| `TrendingStocks` | Top mentioned stocks |

### Interactive Components

| Component | Purpose |
|-----------|---------|
| `StockChart` | Interactive TradingView-style chart |
| `ScreenerTable` | Filterable stock table |
| `StockDataWidget` | Symbol quick stats |
| `NewsletterForm` | Email subscription |

## API Routes

### `POST /api/subscribe`
Subscribe to newsletter for daily reports.

**Request:**
```typescript
{
  email: string;
}
```

**Response:**
```typescript
{
  success: boolean;
  message: string;
}
```

## Data Sources

### Market Data
- **API**: TCBS (apipubaws.tcbs.com.vn)
- **Update**: Real-time during market hours

### Articles
- **Source**: Local markdown files or CMS
- **API**: Internal `/lib/articles-api.ts`
- **Cache**: ISR with 60s revalidation

## Styling

### Neo-Brutalism Design System

```css
/* Colors */
--ink: #1a1a1a;       /* Primary text */
--yellow: #fbbf24;    /* Accent */
--red: #ef4444;       /* Sell/negative */
--green: #22c55e;     /* Buy/positive */

/* Borders */
border-3: 3px solid var(--ink);
shadow-brutal: 4px 4px 0 var(--ink);

/* Buttons */
.btn-brutal {
  border: 3px solid var(--ink);
  box-shadow: 4px 4px 0 var(--ink);
  transition: all 0.1s;
}
.btn-brutal:hover {
  transform: translate(-2px, -2px);
  box-shadow: 6px 6px 0 var(--ink);
}
```

## Environment Variables

```bash
NEXT_PUBLIC_API_URL=          # Backend API URL
NEXT_PUBLIC_MARKET_API_URL=   # Market data API
NEWSLETTER_API_KEY=           # Newsletter service key
```

## Development

```bash
cd blog-site

# Install dependencies
npm install

# Development server
npm run dev

# Build for production
npm run build

# Start production
npm start
```

## Deployment

**Vercel (Recommended):**
```bash
vercel --prod
```

**Docker:**
```dockerfile
FROM node:20-alpine
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build
EXPOSE 3000
CMD ["npm", "start"]
```

## Performance

- **ISR**: 60-second revalidation for articles
- **Client Polling**: 60-second interval for market data
- **Image Optimization**: Next.js Image component
- **Font Optimization**: `next/font` for Geist

## Related Documentation

- [System Architecture](./system-architecture.md)
- [Code Standards](./code-standards.md)
- [Project Overview](./project-overview-pdr.md)
