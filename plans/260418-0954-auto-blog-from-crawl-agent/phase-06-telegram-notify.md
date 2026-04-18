---
phase: 6
title: "Telegram Notify Admin"
status: completed
effort: 0.5h
completed: 2026-04-18
---

# Phase 6: Telegram Notify Admin

## Context Links
- Phase trước: [phase-03-article-generator.md](./phase-03-article-generator.md)
- Article generator: `crawl-agent/app/services/article_generator.py`
- Telegram Bot Token: `TELEGRAM_BOT_TOKEN` env var (đã có trong .env.example)

## Overview

Sau khi ArticleGenerator tạo xong drafts, gửi Telegram message thông báo cho admin:

```
📝 3 bài viết draft mới đã được tạo:
• VNM - Phân tích VNM ngày 18/04/2026
• HPG - Phân tích HPG ngày 18/04/2026
• VCB - Phân tích VCB ngày 18/04/2026

Duyệt tại: http://dashboard.local/admin/articles
```

**Không cần telegram-bot service** — gọi Telegram Bot API trực tiếp từ crawl-agent bằng `httpx`.

## Related Code Files

**Sửa:**
- `crawl-agent/app/services/article_generator.py` — thêm `_notify_telegram()`
- `crawl-agent/app/config.py` — thêm `telegram_admin_chat_id` và `dashboard_url`

## Implementation Steps

### 1. Thêm settings vào `crawl-agent/app/config.py`

```python
telegram_admin_chat_id: str = ""   # chat ID nhận notify
dashboard_url: str = "http://localhost:3000"
```

### 2. Sửa `ArticleGenerator` — thêm Telegram notify

Thêm vào `__init__`:
```python
self._telegram_token = os.environ.get("TELEGRAM_BOT_TOKEN", "")
self._telegram_chat_id = admin_chat_id  # từ settings
self._dashboard_url = dashboard_url
```

Thêm method `_notify_telegram()`:
```python
async def _notify_telegram(self, created_articles: List[Dict[str, str]]) -> None:
    if not self._telegram_token or not self._telegram_chat_id:
        return
    if not created_articles:
        return

    lines = [f"📝 {len(created_articles)} bài viết draft mới:\n"]
    for a in created_articles:
        lines.append(f"• {a['symbol']} — {a['title']}")
    lines.append(f"\nDuyệt tại: {self._dashboard_url}/admin/articles")
    text = "\n".join(lines)

    try:
        async with httpx.AsyncClient(timeout=10) as client:
            await client.post(
                f"https://api.telegram.org/bot{self._telegram_token}/sendMessage",
                json={"chat_id": self._telegram_chat_id, "text": text},
            )
    except Exception as exc:
        logger.warning("Telegram notify failed: %s", exc)
```

Sửa `run()` để track titles và gọi notify:
```python
async def run(self) -> Dict[str, int]:
    hot = await self.get_hot_symbols()
    generated = 0
    created = []
    for symbol in hot[: self._max_daily]:
        result = await self.generate_for_symbol(symbol)
        if result:
            generated += 1
            created.append(result)  # generate_for_symbol trả về dict {symbol, title}

    await self._notify_telegram(created)
    return {"symbols_processed": len(hot), "articles_created": generated}
```

> `generate_for_symbol()` cần sửa để trả về `{"symbol": ..., "title": ...}` thay vì `bool`.

### 3. Thêm env vào `.env.example`

```env
TELEGRAM_ADMIN_CHAT_ID=123456789
DASHBOARD_URL=http://localhost:3000
```

## Todo List

- [ ] Thêm `telegram_admin_chat_id`, `dashboard_url` vào `config.py`
- [ ] Thêm `_notify_telegram()` vào `ArticleGenerator`
- [ ] Sửa `run()` để collect created article info và gọi notify
- [ ] Sửa `generate_for_symbol()` return type: `Optional[Dict]` thay vì `bool`
- [ ] Test: TELEGRAM_BOT_TOKEN không set → skip gracefully

## Success Criteria

- Sau scheduler run có drafts → Telegram message gửi đến admin chat
- `TELEGRAM_BOT_TOKEN` không set → skip, không crash
- Message format đúng với list symbols và link duyệt

## Risk Assessment

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| Telegram API down | Low | try/except, chỉ log warning |
| Chat ID sai | Low | Silent fail (Telegram trả 400, chỉ log) |
