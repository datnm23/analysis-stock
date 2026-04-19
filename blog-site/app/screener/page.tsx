import type { Metadata } from "next";
import { ScreenerTable } from "@/components/screener-table";

export const metadata: Metadata = {
  title: "Screener — VietStock AI",
  description: "Lọc cổ phiếu theo khuyến nghị AI, độ tin cậy và điểm kỹ thuật — HSX/HNX/UPCOM",
};

export default function ScreenerPage() {
  return (
    <div>
      {/* Header */}
      <div className="mb-8">
        <div className="flex items-end gap-4 mb-1">
          <h1 className="text-4xl font-black text-ink uppercase tracking-tight">Screener</h1>
          <span className="bg-yellow border-2 border-ink text-ink font-black text-sm px-3 py-1 mb-1">
            AI Powered
          </span>
        </div>
        <p className="text-ink/50 text-sm mt-2 mb-1">
          Lọc cổ phiếu theo khuyến nghị AI mới nhất · Sắp xếp theo độ tin cậy · Phân tích kỹ thuật + Sentiment
        </p>
        <div className="h-1 w-24 bg-ink" />
      </div>

      <ScreenerTable />

      {/* Disclaimer */}
      <div className="border-3 border-ink bg-yellow shadow-brutal-sm p-4 mt-8">
        <p className="text-xs font-black text-ink uppercase tracking-wide mb-1">⚠ Lưu ý</p>
        <p className="text-xs text-ink/70 leading-relaxed">
          Dữ liệu do AI tổng hợp từ phân tích kỹ thuật và sentiment. Chỉ mang tính tham khảo — không phải lời khuyên đầu tư.
        </p>
      </div>
    </div>
  );
}
