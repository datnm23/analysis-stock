"use client";

import { useRouter, useSearchParams } from "next/navigation";

interface Tab {
  label: string;
  symbol: string;
}

export function MarketTabSwitcher({ tabs, activeSymbol }: { tabs: Tab[]; activeSymbol: string }) {
  const router = useRouter();
  useSearchParams(); // subscribe to param changes

  return (
    <div className="flex gap-2 flex-wrap">
      {tabs.map((tab) => (
        <button
          key={tab.symbol}
          onClick={() => router.push(`/market?tab=${tab.symbol}`)}
          className={`font-black text-sm px-4 py-2 border-3 border-ink uppercase tracking-wide transition-none
            ${tab.symbol === activeSymbol
              ? "bg-ink text-yellow"
              : "bg-white text-ink hover:bg-yellow"}`}
        >
          {tab.label}
        </button>
      ))}
    </div>
  );
}
