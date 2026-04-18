import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Phân tích chứng khoán Việt Nam",
  description: "Bài viết phân tích chứng khoán HSX/HNX/UPCOM được tạo tự động từ AI",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="vi">
      <body className="min-h-screen bg-gray-50">
        <header className="bg-white border-b">
          <nav className="max-w-4xl mx-auto px-6 py-4 flex gap-6 items-center">
            <a href="/" className="text-lg font-bold text-blue-700 hover:text-blue-800">
              StockBlog
            </a>
            <a href="/articles" className="text-sm text-gray-600 hover:text-blue-600">
              Tất cả bài viết
            </a>
          </nav>
        </header>
        <main className="max-w-4xl mx-auto px-6 py-8">{children}</main>
        <footer className="border-t mt-16 py-6 text-center text-xs text-gray-400">
          Nội dung chỉ mang tính tham khảo, không phải lời khuyên đầu tư.
        </footer>
      </body>
    </html>
  );
}
