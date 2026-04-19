---
title: Phase 03 – Sector mapping static data
status: completed
progress: 100%
completed: 2026-04-20
---

# Phase 03 – Sector mapping: ~200 mã → 12 ngành

## Overview

Tạo file TypeScript tĩnh mapping symbol → sector. Dùng phía frontend để group/filter bảng giá mà không cần DB hay API call thêm.

## File

- **Tạo mới**: `blog-site/lib/sector-mapping.ts`

## 12 Ngành

| Sector key | Tên hiển thị | Mã tiêu biểu |
|-----------|--------------|--------------|
| `banking` | Ngân hàng | VCB, BID, CTG, TCB, VPB, MBB, ACB, HDB, STB, TPB, VIB, MSB, SHB, OCB |
| `securities` | Chứng khoán | SSI, VND, VCI, HCM, MBS, FTS, SHS, BSI, AGR, CTS, ORS |
| `real-estate` | Bất động sản | VIC, VHM, NVL, PDR, DXG, CEO, KBC, HDG, DIG, KDH, BCM, NLG, SZC |
| `steel` | Thép & VL XD | HPG, HSG, NKG, TVN, SMC, TNA, VGC, BMP |
| `oil-gas` | Dầu khí | PVS, PVD, GAS, BSR, OIL, PLX, PVC, PVT, CNG, PCT |
| `consumer` | Tiêu dùng & BL | MWG, PNJ, MSN, VNM, KDC, QNS, MCH, NET, PAN, SAB, VHC, ANV |
| `technology` | Công nghệ | FPT, CMG, ELC, VGI, ITD |
| `energy` | Điện & Năng lượng | POW, PC1, REE, SBA, BCG, GEG, PGV, HDC, HND |
| `transport` | Vận tải & Hàng không | ACV, HVN, VJC, SGN, GMD, VSC, PVT, DVP, HAH |
| `healthcare` | Y tế & Dược phẩm | DHG, IMP, DMC, TRA, OPC, DCL, PME, DBD, SPM |
| `industrial` | Công nghiệp | GVR, CSV, DGC, PHR, VRG, HVH, CTD, HBC, VCG, FCN |
| `finance` | Tài chính khác | BVH, MIG, PVI, PTI, BSH, VNR, ABI, BIC |

## Implementation

```typescript
export interface SectorInfo {
  key: string;
  label: string;
}

export const SECTORS: SectorInfo[] = [
  { key: "all",         label: "Tất cả" },
  { key: "banking",     label: "Ngân hàng" },
  { key: "securities",  label: "Chứng khoán" },
  { key: "real-estate", label: "Bất động sản" },
  { key: "steel",       label: "Thép & VL XD" },
  { key: "oil-gas",     label: "Dầu khí" },
  { key: "consumer",    label: "Tiêu dùng" },
  { key: "technology",  label: "Công nghệ" },
  { key: "energy",      label: "Điện & Năng lượng" },
  { key: "transport",   label: "Vận tải" },
  { key: "healthcare",  label: "Y tế & Dược" },
  { key: "industrial",  label: "Công nghiệp" },
  { key: "finance",     label: "Tài chính" },
];

// Symbol → sector key. Unlisted symbols fall into "other".
export const SYMBOL_SECTOR: Record<string, string> = {
  // Ngân hàng
  VCB:"banking", BID:"banking", CTG:"banking", TCB:"banking", VPB:"banking",
  MBB:"banking", ACB:"banking", HDB:"banking", STB:"banking", TPB:"banking",
  VIB:"banking", MSB:"banking", SHB:"banking", OCB:"banking", SSB:"banking",
  LPB:"banking", ABB:"banking", NVB:"banking", KLB:"banking", VAB:"banking",
  // Chứng khoán
  SSI:"securities", VND:"securities", VCI:"securities", HCM:"securities",
  MBS:"securities", FTS:"securities", SHS:"securities", BSI:"securities",
  AGR:"securities", CTS:"securities", ORS:"securities", VDS:"securities",
  // Bất động sản
  VIC:"real-estate", VHM:"real-estate", NVL:"real-estate", PDR:"real-estate",
  DXG:"real-estate", CEO:"real-estate", KBC:"real-estate", HDG:"real-estate",
  DIG:"real-estate", KDH:"real-estate", BCM:"real-estate", NLG:"real-estate",
  SZC:"real-estate", ITC:"real-estate", TDC:"real-estate", SCR:"real-estate",
  // Thép & VL XD
  HPG:"steel", HSG:"steel", NKG:"steel", TVN:"steel", SMC:"steel",
  TNA:"steel", VGC:"steel", BMP:"steel", CSV:"steel",
  // Dầu khí
  PVS:"oil-gas", PVD:"oil-gas", GAS:"oil-gas", BSR:"oil-gas", OIL:"oil-gas",
  PLX:"oil-gas", PVC:"oil-gas", PVT:"oil-gas", CNG:"oil-gas", PCT:"oil-gas",
  // Tiêu dùng & Bán lẻ
  MWG:"consumer", PNJ:"consumer", MSN:"consumer", VNM:"consumer", KDC:"consumer",
  QNS:"consumer", MCH:"consumer", NET:"consumer", PAN:"consumer", SAB:"consumer",
  VHC:"consumer", ANV:"consumer", IDI:"consumer",
  // Công nghệ
  FPT:"technology", CMG:"technology", ELC:"technology", VGI:"technology",
  // Điện & Năng lượng
  POW:"energy", PC1:"energy", REE:"energy", SBA:"energy", BCG:"energy",
  GEG:"energy", PGV:"energy", HDC:"energy", HND:"energy", VSH:"energy",
  // Vận tải & Hàng không
  ACV:"transport", HVN:"transport", VJC:"transport", SGN:"transport",
  GMD:"transport", VSC:"transport", DVP:"transport", HAH:"transport",
  // Y tế & Dược phẩm
  DHG:"healthcare", IMP:"healthcare", DMC:"healthcare", TRA:"healthcare",
  OPC:"healthcare", DCL:"healthcare", PME:"healthcare", DBD:"healthcare",
  SPM:"healthcare",
  // Công nghiệp
  GVR:"industrial", DGC:"industrial", PHR:"industrial", VRG:"industrial",
  CTD:"industrial", HBC:"industrial", VCG:"industrial", FCN:"industrial",
  // Tài chính khác
  BVH:"finance", MIG:"finance", PVI:"finance", PTI:"finance",
  BSH:"finance", VNR:"finance", ABI:"finance", BIC:"finance",
};

export function getSector(symbol: string): string {
  return SYMBOL_SECTOR[symbol.toUpperCase()] ?? "other";
}

export function getSectorLabel(key: string): string {
  return SECTORS.find((s) => s.key === key)?.label ?? key;
}
```

## Notes

- Mã không có trong mapping → hiện ở tab "Tất cả", không có sector tag
- Có thể bổ sung thêm symbol bất kỳ lúc nào mà không cần backend change
- Không dùng DB vì data tĩnh, hiếm thay đổi, và không cần join với price data
