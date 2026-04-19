import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "VietStock AI – Phân tích chứng khoán Việt Nam",
  description: "Bài viết phân tích chứng khoán HSX/HNX/UPCOM kết hợp AI và phân tích kỹ thuật",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="vi">
      <body className="min-h-screen bg-cream">
        {/* ─── Header ─── */}
        <header className="bg-ink sticky top-0 z-20 border-b-3 border-ink">
          <nav className="max-w-6xl mx-auto px-4 md:px-6 h-14 flex items-center justify-between">
            {/* Logo */}
            <a href="/" className="flex items-center gap-2 group">
              <span className="bg-yellow text-ink text-xs font-black px-2 py-1 border-2 border-ink group-hover:bg-white transition-colors duration-100">
                AI
              </span>
              <span className="text-white font-black text-lg tracking-tight uppercase">
                VietStock
              </span>
            </a>

            {/* Nav links */}
            <div className="flex items-center gap-1">
              <a
                href="/"
                className="text-white/70 hover:text-yellow font-semibold text-sm px-3 py-1.5 uppercase tracking-wide transition-colors duration-100"
              >
                Trang chủ
              </a>
              <a
                href="/articles"
                className="bg-yellow text-ink font-black text-sm px-4 py-1.5 border-2 border-yellow uppercase tracking-wide hover:bg-white hover:border-white transition-colors duration-100"
              >
                Bài viết
              </a>
            </div>
          </nav>
        </header>

        {/* ─── Main ─── */}
        <main className="max-w-6xl mx-auto px-4 md:px-6 py-10">
          {children}
        </main>

        {/* ─── Footer ─── */}
        <footer className="border-t-3 border-ink bg-ink mt-16 py-10">
          <div className="max-w-6xl mx-auto px-4 md:px-6 flex flex-col md:flex-row items-start md:items-center justify-between gap-4">
            <div>
              <p className="font-black text-yellow text-lg uppercase tracking-tight">VietStock AI</p>
              <p className="text-white/50 text-xs mt-1 max-w-sm leading-relaxed">
                AI phân tích · Bạn quyết định.<br/>
                Không phải lời khuyên đầu tư.
              </p>
            </div>
            <div className="text-right">
              <p className="text-white/30 text-xs">HSX · HNX · UPCOM</p>
              <p className="text-white/30 text-xs mt-0.5">Powered by Claude AI · KB Securities</p>
            </div>
          </div>
        </footer>
      </body>
    </html>
  );
}
