"use client";

import { useState } from "react";

interface Props {
  variant?: "footer" | "article";
  symbol?: string;
}

export function NewsletterForm({ variant = "footer", symbol }: Props) {
  const [email, setEmail] = useState("");
  const [status, setStatus] = useState<"idle" | "loading" | "success" | "error">("idle");
  const [errorMsg, setErrorMsg] = useState("");

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!email) return;
    setStatus("loading");
    try {
      const res = await fetch("/api/subscribe", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email }),
      });
      if (res.ok) {
        setStatus("success");
        setEmail("");
      } else {
        const json = await res.json().catch(() => ({}));
        setErrorMsg(json.error ?? "Lỗi không xác định");
        setStatus("error");
      }
    } catch {
      setErrorMsg("Không kết nối được — thử lại sau.");
      setStatus("error");
    }
  }

  const heading =
    variant === "article" && symbol
      ? `Nhận phân tích ${symbol} khi có bài mới`
      : "Nhận phân tích mỗi buổi sáng";

  return (
    <div>
      <p className={`font-black uppercase tracking-wide mb-3 ${variant === "footer" ? "text-white text-sm" : "text-ink text-base"}`}>
        📬 {heading}
      </p>

      {status === "success" ? (
        <p className={`text-sm font-bold ${variant === "footer" ? "text-[#05D98F]" : "text-[#05D98F]"}`}>
          ✓ Đã đăng ký! Kiểm tra email của bạn.
        </p>
      ) : (
        <form onSubmit={handleSubmit} className="flex flex-col sm:flex-row gap-2">
          <input
            type="email"
            required
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="email@example.com"
            disabled={status === "loading"}
            className={`flex-1 px-3 py-2 font-mono text-sm border-3 border-ink outline-none focus:ring-0 disabled:opacity-50
              ${variant === "footer" ? "bg-white/10 text-white placeholder:text-white/40" : "bg-white text-ink placeholder:text-ink/40"}`}
          />
          <button
            type="submit"
            disabled={status === "loading"}
            className="btn-brutal bg-yellow text-ink font-black text-sm px-4 py-2 uppercase tracking-wide border-3 border-ink disabled:opacity-50 whitespace-nowrap"
          >
            {status === "loading" ? "..." : "Đăng ký →"}
          </button>
        </form>
      )}

      {status === "error" && (
        <p className="text-[#FF4A4A] text-xs font-mono mt-1">{errorMsg}</p>
      )}

      <p className={`text-[10px] mt-2 ${variant === "footer" ? "text-white/30" : "text-ink/40"}`}>
        Miễn phí · Không spam · Hủy bất cứ lúc nào
      </p>
    </div>
  );
}
