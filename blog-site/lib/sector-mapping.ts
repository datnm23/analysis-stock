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
  { key: "energy",      label: "Điện & NL" },
  { key: "transport",   label: "Vận tải" },
  { key: "healthcare",  label: "Y tế & Dược" },
  { key: "industrial",  label: "Công nghiệp" },
  { key: "finance",     label: "Tài chính" },
];

export const SYMBOL_SECTOR: Record<string, string> = {
  // Ngân hàng
  VCB:"banking", BID:"banking", CTG:"banking", TCB:"banking", VPB:"banking",
  MBB:"banking", ACB:"banking", HDB:"banking", STB:"banking", TPB:"banking",
  VIB:"banking", MSB:"banking", SHB:"banking", OCB:"banking", SSB:"banking",
  LPB:"banking", ABB:"banking", NVB:"banking", KLB:"banking", VAB:"banking",
  BVB:"banking", PGB:"banking", NAB:"banking", BAB:"banking", VBB:"banking",
  // Chứng khoán
  SSI:"securities", VND:"securities", VCI:"securities", HCM:"securities",
  MBS:"securities", FTS:"securities", SHS:"securities", BSI:"securities",
  AGR:"securities", CTS:"securities", ORS:"securities", VDS:"securities",
  TVS:"securities", TCI:"securities", PSI:"securities", BVS:"securities",
  // Bất động sản
  VIC:"real-estate", VHM:"real-estate", NVL:"real-estate", PDR:"real-estate",
  DXG:"real-estate", CEO:"real-estate", KBC:"real-estate", HDG:"real-estate",
  DIG:"real-estate", KDH:"real-estate", BCM:"real-estate", NLG:"real-estate",
  SZC:"real-estate", ITC:"real-estate", TDC:"real-estate", SCR:"real-estate",
  NBB:"real-estate", DXS:"real-estate", HQC:"real-estate", LDG:"real-estate",
  // Thép & Vật liệu xây dựng
  HPG:"steel", HSG:"steel", NKG:"steel", TVN:"steel", SMC:"steel",
  TNA:"steel", VGC:"steel", BMP:"steel", CSV:"steel", TIS:"steel",
  POM:"steel", DNY:"steel",
  // Dầu khí
  PVS:"oil-gas", PVD:"oil-gas", GAS:"oil-gas", BSR:"oil-gas", OIL:"oil-gas",
  PLX:"oil-gas", PVC:"oil-gas", CNG:"oil-gas", PCT:"oil-gas",
  PGS:"oil-gas", PSH:"oil-gas",
  // Tiêu dùng & Bán lẻ
  MWG:"consumer", PNJ:"consumer", MSN:"consumer", VNM:"consumer", KDC:"consumer",
  QNS:"consumer", MCH:"consumer", NET:"consumer", PAN:"consumer", SAB:"consumer",
  VHC:"consumer", ANV:"consumer", IDI:"consumer", FMC:"consumer", HVN:"consumer",
  // Công nghệ
  FPT:"technology", CMG:"technology", ELC:"technology", VGI:"technology",
  ITD:"technology", SGT:"technology", ONE:"technology",
  // Điện & Năng lượng
  POW:"energy", PC1:"energy", REE:"energy", SBA:"energy", BCG:"energy",
  GEG:"energy", PGV:"energy", HDC:"energy", HND:"energy", VSH:"energy",
  QTP:"energy", NT2:"energy", TBC:"energy", SHP:"energy", HJS:"energy",
  // Vận tải & Hàng không
  ACV:"transport", VJC:"transport", SGN:"transport",
  GMD:"transport", VSC:"transport", DVP:"transport", HAH:"transport",
  PVT:"transport", STG:"transport", VOS:"transport", VTO:"transport",
  // Y tế & Dược phẩm
  DHG:"healthcare", IMP:"healthcare", DMC:"healthcare", TRA:"healthcare",
  OPC:"healthcare", DCL:"healthcare", PME:"healthcare", DBD:"healthcare",
  SPM:"healthcare", TNH:"healthcare", DVN:"healthcare",
  // Công nghiệp
  GVR:"industrial", DGC:"industrial", PHR:"industrial", VRG:"industrial",
  CTD:"industrial", HBC:"industrial", VCG:"industrial", FCN:"industrial",
  VLB:"industrial", THG:"industrial", VLG:"industrial",
  // Tài chính (bảo hiểm, quỹ)
  BVH:"finance", MIG:"finance", PVI:"finance", PTI:"finance",
  BSH:"finance", VNR:"finance", ABI:"finance", BIC:"finance",
};

export function getSector(symbol: string): string {
  return SYMBOL_SECTOR[symbol.toUpperCase()] ?? "other";
}

export function getSectorLabel(key: string): string {
  return SECTORS.find((s) => s.key === key)?.label ?? key;
}
