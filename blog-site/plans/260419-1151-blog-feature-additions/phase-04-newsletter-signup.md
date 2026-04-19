# Phase 4: Newsletter Signup

**Priority:** High  
**Effort:** ~0.5 ngày  
**Status:** pending  
**Độc lập** — có thể làm song song với Phase 1

## Overview

Email capture form đơn giản — không cần auth, không cần backend phức tạp.  
**MVP approach:** Lưu email vào file JSON / Google Sheets qua API, hoặc dùng Resend free tier.

**Đặt ở 2 vị trí:**
1. Footer của `layout.tsx` (persistent, mọi trang)
2. CTA block cuối mỗi article detail (contextual, high intent)

## Related Code Files

**Tạo mới:**
- `blog-site/components/newsletter-form.tsx` — reusable form component
- `blog-site/app/api/subscribe/route.ts` — Next.js API route handler

**Sửa:**
- `blog-site/app/layout.tsx` — thêm NewsletterForm vào footer
- `blog-site/app/articles/[slug]/page.tsx` — thêm NewsletterForm trước Disclaimer block

## Implementation Steps

### API Route (Next.js)

1. Tạo `blog-site/app/api/subscribe/route.ts`:
```ts
// POST { email: string }
// Validation: email format, max 254 chars
// Option A: Ghi vào file subscribers.json (simplest, no external dep)
// Option B: POST tới Resend audience API (nếu có RESEND_API_KEY)
// Option C: POST tới Google Sheets via Apps Script webhook
// Response: { success: true } hoặc { error: "..." }
```

**MVP recommendation: Option A (file JSON)** — zero dependency, migrate sau.
```ts
// Append email + timestamp vào /data/subscribers.json
// Dedup check trước khi ghi
```

2. Env var: `NEWSLETTER_WEBHOOK_URL` (optional) — nếu có thì POST tới đó thay file.

### Component

3. Tạo `blog-site/components/newsletter-form.tsx` (Client component):
```tsx
// State: email, status ("idle" | "loading" | "success" | "error")
// Form: 1 input email + 1 button
// Success state: "✓ Đã đăng ký! Kiểm tra email của bạn."
// Error state: hiển thị error message
// No full-page reload — fetch API
```

4. UI mockup:
```
┌──────────────────────────────────────────────────────┐
│  📬 Nhận phân tích mỗi buổi sáng                    │
│  [email@example.com              ] [Đăng ký →]      │
│  Miễn phí · Không spam · Hủy bất cứ lúc nào        │
└──────────────────────────────────────────────────────┘
```

5. Thêm vào `layout.tsx` footer — trước copyright block:
```tsx
<div className="border-t-3 border-white/10 pt-8 mt-4">
  <NewsletterForm />
</div>
```

6. Thêm vào `articles/[slug]/page.tsx` — trước `<div className="border-3 ... bg-yellow">` (Disclaimer):
```tsx
<div className="border-3 border-ink bg-ink shadow-brutal mb-8 p-6">
  <NewsletterForm variant="article" />
</div>
```

### Variant

7. `variant="article"` hiển thị copy khác: "Nhận phân tích {symbol} khi có bài mới"

## Data Storage Decision

| Option | Pros | Cons | Recommend |
|--------|------|------|-----------|
| File JSON | Zero dep, simple | Không scale, mất khi redeploy | MVP only |
| Resend free | 3k emails/mo free, proper API | Cần signup | **Best nếu có account** |
| Google Sheets | Dễ xem, free | Setup phức tạp hơn | Alternative |

**Recommend:** Bắt đầu với file JSON, switch sang Resend khi cần gửi email thật.

## Success Criteria

- [ ] Form submit không reload trang
- [ ] Success/error state hiển thị đúng
- [ ] Email được lưu (file hoặc service)
- [ ] Dedup — cùng email không ghi 2 lần
- [ ] Mobile-friendly (button full-width trên mobile)
