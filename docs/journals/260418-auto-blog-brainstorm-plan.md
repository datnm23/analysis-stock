# Auto Blog từ Crawl-Agent: Brainstorm & Lên Kế Hoạch

**Ngày**: 2026-04-18 09:54
**Mức độ**: Medium
**Thành phần**: crawl-agent, go-services, web-dashboard, blog-site (mới)
**Trạng thái**: Kế hoạch đã hoàn thành

## Chuyện gì xảy ra

Brainstorm và lên kế hoạch chi tiết feature "Auto Blog from Crawl-Agent" — tự động sinh bài blog từ dữ liệu cổ phiếu nóng, gọi Claude Haiku API, và quản lý qua admin panel.

## Quyết định chính

- **Architecture**: crawl-agent scheduler (Python) → Redis hot symbols → call `/api/v1/forecast/{symbol}` (Go) → Claude Haiku via httpx → POST draft → Telegram notify admin → admin review tại web-dashboard → publish
- **Loại n8n**: n8n directory không tồn tại, dùng crawl-agent scheduler thay thế
- **Loại Anthropic SDK**: Call Claude API trực tiếp via httpx (sẵn trong requirements), tránh dependency mới
- **Model**: Claude Haiku (~$0.05/ngày cho 10 bài)
- **DB**: GORM auto-migration cho articles table
- **Auth**: X-Internal-Key (internal POST), NEXT_PUBLIC_ADMIN_KEY (admin UI)
- **Blog**: SSG + ISR (revalidate 3600s) cho SEO + performance

## Rủi ro

- Latency: crawl-agent → go-services → Claude API có thể > 5s/bài (cần async queue)
- Admin review bottleneck nếu chưa có interface smooth
- Cost vượt nếu hot symbols > 20/ngày

## Bước tiếp theo

Triển khai 6 phases theo plan (tổng ~11h effort). Bắt đầu từ DB model.

**Plan file**: `/media/datnm/Data/Java/analysis-stock/plans/260418-0954-auto-blog-from-crawl-agent/`
