import Link from "next/link";
import { getLatestArticles } from "@/lib/articles-api";
import { ArticleCard } from "@/components/article-card";

export const revalidate = 60;

export default async function HomePage() {
  const articles = await getLatestArticles(6);

  return (
    <div>
      {/* ─── Hero — PAS formula ─── */}
      <div className="border-3 border-ink bg-cream shadow-brutal-xl mb-12 overflow-hidden">
        <div className="grid md:grid-cols-[1fr_auto] gap-0">

          {/* Left: PAS copy */}
          <div className="p-8 md:p-12 border-r-0 md:border-r-3 md:border-ink">
            {/* [P] Problem hook */}
            <div className="inline-block bg-ink text-yellow text-xs font-black px-3 py-1 uppercase tracking-widest mb-6">
              Đọc báo cáo môi giới xong vẫn không biết mua gì?
            </div>

            {/* [A] Agitate → [S] Solution headline — 4Us: Urgent+Unique+Useful+Ultra-specific */}
            <h1 className="text-4xl md:text-5xl font-black text-ink leading-[1.05] uppercase tracking-tight mb-5">
              Chúng tôi quét<br />
              <span className="text-yellow bg-ink px-2">hàng chục nguồn</span><br />
              — cho bạn 1 kết luận
            </h1>

            {/* Solution promise */}
            <p className="text-ink/70 text-base max-w-md leading-relaxed mb-8">
              Mỗi ngày AI tổng hợp dữ liệu kỹ thuật, sentiment tin tức và mô hình dự báo
              để trả lời thẳng: cổ phiếu này nên <strong className="text-ink bg-[#05D98F] px-1">MUA</strong>,{" "}
              <strong className="text-ink bg-[#FFD93D] px-1">GIỮ</strong> hay{" "}
              <strong className="text-ink bg-[#FF4A4A] px-1">BÁN</strong>.
            </p>

            {/* CTA */}
            <Link
              href="/articles"
              className="btn-brutal inline-block bg-yellow text-ink font-black text-sm px-6 py-3 uppercase tracking-widest"
            >
              Xem phân tích hôm nay →
            </Link>
          </div>

          {/* Right: stat block — FAB micro-copy */}
          <div className="bg-ink p-8 md:p-12 flex flex-col justify-center gap-6 min-w-[200px]">
            {[
              { label: "Sàn theo dõi", value: "HSX · HNX" },
              { label: "Mô hình AI", value: "Claude AI" },
              { label: "Tần suất", value: "Mỗi ngày" },
            ].map(({ label, value }) => (
              <div key={label}>
                <p className="text-white/40 text-[11px] font-semibold uppercase tracking-widest mb-1">{label}</p>
                <p className="text-yellow font-black text-lg">{value}</p>
              </div>
            ))}
          </div>
        </div>

        {/* Ticker strip — urgency micro-copy */}
        <div className="bg-yellow border-t-3 border-ink px-6 py-2 flex items-center gap-3 overflow-hidden">
          <span className="text-ink font-black text-[11px] uppercase tracking-widest shrink-0">
            ĐANG THEO DÕI:
          </span>
          <div className="flex gap-4 overflow-x-auto">
            {["VNM", "HPG", "VCB", "BID", "SSI", "MSN", "FPT", "VIC", "TCB"].map((s) => (
              <Link
                key={s}
                href={`/articles`}
                className="text-ink font-black text-xs shrink-0 hover:underline underline-offset-2"
              >
                {s}
              </Link>
            ))}
          </div>
        </div>
      </div>

      {/* ─── Latest articles ─── */}
      <div className="flex items-center justify-between mb-6">
        {/* 4Us headline: Unique + Useful */}
        <h2 className="text-2xl font-black text-ink uppercase tracking-tight border-b-3 border-yellow pb-1">
          Phân tích mới nhất
        </h2>
        <Link
          href="/articles"
          className="btn-brutal bg-ink text-yellow text-xs font-black px-4 py-2 uppercase tracking-widest"
        >
          Xem tất cả →
        </Link>
      </div>

      {articles.length === 0 ? (
        <div className="border-3 border-ink bg-white py-16 text-center shadow-brutal">
          <p className="font-black text-ink/30 text-xl uppercase tracking-widest mb-2">
            Chưa có bài nào.
          </p>
          <p className="text-ink/40 text-sm">Quay lại sau — AI đang xử lý dữ liệu.</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {articles.map((a) => (
            <ArticleCard key={a.id} article={a} />
          ))}
        </div>
      )}

      {/* ─── Bottom CTA — urgency + benefit ─── */}
      {articles.length > 0 && (
        <div className="mt-12 border-3 border-ink bg-yellow p-6 flex flex-col sm:flex-row items-center justify-between gap-4 shadow-brutal">
          <div>
            <p className="font-black text-ink text-xl uppercase tracking-tight">
              Không bỏ lỡ bất kỳ cơ hội nào.
            </p>
            <p className="text-ink/60 text-sm mt-1">
              Phân tích kỹ thuật + AI — cập nhật mỗi ngày giao dịch.
            </p>
          </div>
          <Link
            href="/articles"
            className="btn-brutal bg-ink text-yellow font-black px-6 py-3 uppercase text-sm tracking-widest shrink-0"
          >
            Đọc toàn bộ →
          </Link>
        </div>
      )}
    </div>
  );
}
