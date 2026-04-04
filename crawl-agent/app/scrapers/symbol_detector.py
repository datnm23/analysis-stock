"""
Vietnamese stock symbol detection and company name resolution.

Maps company names (Vietnamese) to stock ticker symbols and provides
enhanced symbol extraction from news articles.

Coverage: ~200 most-traded stocks on HOSE, HNX, UPCOM.
"""

import re
from typing import List, Set

# ---------------------------------------------------------------------------
# Company name → Ticker mapping (top VN stocks)
# ---------------------------------------------------------------------------

COMPANY_TO_TICKER = {
    # Banking
    "vietcombank": "VCB", "vcb": "VCB",
    "vietinbank": "CTG", "ctg": "CTG",
    "bidv": "BID",
    "techcombank": "TCB", "tcb": "TCB",
    "mb bank": "MBB", "mb ": "MBB", "quân đội": "MBB",
    "vpbank": "VPB", "vp bank": "VPB",
    "acb": "ACB",
    "sacombank": "STB", "stb": "STB",
    "shb": "SHB", "sài gòn - hà nội": "SHB",
    "tpbank": "TPB", "tp bank": "TPB",
    "hdbank": "HDB", "hd bank": "HDB",
    "eximbank": "EIB",
    "lienvietpostbank": "LPB", "liên việt": "LPB",
    "ocb": "OCB",
    "msb": "MSB",
    "vib": "VIB",
    "bắc á": "BAB",
    "kiên long": "KLB",
    "nam á": "NAB",
    "seabank": "SSB",
    "bảo việt bank": "BVB",

    # Real Estate
    "vingroup": "VIC", "vin group": "VIC",
    "vinhomes": "VHM",
    "vincom retail": "VRE",
    "novaland": "NVL",
    "khang điền": "KDH",
    "đất xanh": "DXG",
    "phát đạt": "PDR",
    "nam long": "NLG",
    "hoa phát": "HPG",  # Steel but often grouped
    "becamex": "BCM",
    "kinh bắc": "KBC",
    "cen land": "CRE",
    "an gia": "AGG",
    "văn phú": "VPI",

    # Technology & Telecom
    "fpt": "FPT",
    "viettel": "VGI",
    "vnpt": "VNP",
    "cmg": "CMG",

    # Retail & Consumer
    "thế giới di động": "MWG", "thegioididong": "MWG", "mobile world": "MWG",
    "điện máy xanh": "MWG",
    "masan": "MSN",
    "vinamilk": "VNM",
    "sabeco": "SAB",
    "petrolimex": "PLX",
    "pnj": "PNJ", "phú nhuận": "PNJ",

    # Steel & Materials
    "hòa phát": "HPG",
    "hoa sen": "HSG",
    "nam kim": "NKG",
    "pomina": "POM",
    "vinaconex": "VCG",

    # Oil & Gas
    "petrovietnam": "GAS", "pv gas": "GAS", "pvgas": "GAS",
    "pvs": "PVS", "pv drilling": "PVD",
    "bsr": "BSR", "lọc hoá dầu bình sơn": "BSR",
    "pvpower": "POW",

    # Securities
    "ssi": "SSI",
    "vndirect": "VND",
    "hsc": "HCM", "hồ chí minh city securities": "HCM",
    "vcsc": "VCI",
    "mirae asset": "MAS",
    "vnds": "VND",
    "mb securities": "MBS",
    "kafi": "KAF",
    "dnse": "DSE",
    "chứng khoán rồng việt": "VDS",
    "vietcap": "VCI",

    # Utilities
    "evn": "POW",  # EVN subsidiary
    "pha lai": "PPC",
    "điện quang": "DQC",

    # Airlines & Transport
    "vietnam airlines": "HVN",
    "vietjet": "VJC", "vietjet air": "VJC",
    "bamboo airways": "BAV",
    "gemadept": "GMD",
    "cảng sài gòn": "SGP",

    # Agriculture & Food
    "hau giang pharma": "DHG", "dược hậu giang": "DHG",
    "dabaco": "DBC",
    "hoàng anh gia lai": "HAG", "hagl": "HAG", "bầu đức": "HAG",
    "đức giang": "DGC", "hoá chất đức giang": "DGC",

    # Insurance
    "bảo việt": "BVH",

    # FLC Group
    "flc": "FLC", "trịnh văn quyết": "FLC",

    # Construction
    "coteccons": "CTD",
    "hòa bình": "HBC",
    "fecon": "FCN",
    "pc1": "PC1",

    # Indices (not stocks but useful to detect)
    "vn-index": "VNINDEX", "vnindex": "VNINDEX", "vn index": "VNINDEX",
    "vn30": "VN30", "hn30": "HN30",
    "hose": "HOSE", "hnx": "HNX", "upcom": "UPCOM",
}


# Ticker symbols that are REAL stock codes (not abbreviations)
# Only 3-char codes that might be confused with common words
VALID_TICKERS: Set[str] = {
    # Top 50 most traded on HOSE
    "VCB", "CTG", "BID", "TCB", "MBB", "VPB", "ACB", "STB", "SHB", "TPB",
    "HDB", "EIB", "LPB", "OCB", "MSB", "VIB", "SSB",
    "VIC", "VHM", "VRE", "NVL", "KDH", "DXG", "PDR", "NLG", "BCM", "KBC",
    "HPG", "HSG", "NKG", "POM", "VCG",
    "FPT", "CMG", "VGI",
    "MWG", "MSN", "VNM", "SAB", "PLX", "PNJ",
    "GAS", "PVS", "PVD", "BSR", "POW",
    "SSI", "VND", "HCM", "VCI", "MBS", "MAS", "KAF", "DSE", "VDS",
    "HVN", "VJC", "GMD",
    "DHG", "DBC", "HAG", "DGC",
    "BVH", "FLC",
    "CTD", "HBC", "FCN", "PC1",
    "AGG", "VPI", "CRE",
    # Other actively traded
    "REE", "PPC", "DQC", "HVA", "TNG", "TCM", "VGC", "BWE", "PAN",
    "VHC", "ANV", "IDI", "ASM", "DIG", "CEO", "QCG", "SCR", "LDG",
    "HDG", "GEX", "SBT", "DPM", "DCM", "CSV", "IMP", "DVN",
    "VTP", "VOS", "HAH",
    "ELC", "RAL", "FMC", "KSB", "LCG", "IJC", "NBB", "PHR",
    "MPC", "APH", "VND", "SZC",
}

# Words that look like tickers but aren't
NON_STOCK_WORDS: Set[str] = {
    "THE", "AND", "FOR", "VND", "USD", "CEO", "IPO", "ETF", "GDP", "FDI",
    "IMF", "ADB", "WTO", "TPP", "BTC", "FED", "ECB", "BOJ", "CPI", "PMI",
    "RSS", "URL", "API", "XML", "FAQ", "PDF", "HOT", "NEW", "OLD", "BIG",
    "NET", "LPG", "RON", "OIL", "BUY", "PUT", "IFC", "LTD", "JSC", "CNN",
    "BBC", "VOV", "VTV", "VNS", "JPG", "PNG", "DAT", "TOP",
}


# ---------------------------------------------------------------------------
# Enhanced symbol extraction
# ---------------------------------------------------------------------------

# Patterns to match stock symbols in text
_TICKER_3CHAR = re.compile(r'\b([A-Z]{3})\b')
_TICKER_HASHTAG = re.compile(r'#([A-Z]{2,4})\b')  # Telegram-style #MWG
_TICKER_PAREN = re.compile(r'\(([A-Z]{2,4})\)')   # In-text (MWG)
_TICKER_COLON = re.compile(r'(?:HOSE|HNX|UPCOM)[:\s]+([A-Z]{2,4})')  # HOSE: MWG


def extract_symbols(text: str) -> List[str]:
    """Extract stock symbols from text using multiple strategies.

    Strategies:
    1. Direct ticker patterns: [A-Z]{3}, #TICKER, (TICKER), HOSE:TICKER
    2. Company name → ticker mapping
    3. Validation against known ticker list

    Returns deduplicated list of ticker symbols.
    """
    found: Set[str] = set()
    text_upper = text.upper()
    text_lower = text.lower()

    # Strategy 1: Direct ticker patterns
    for pattern in [_TICKER_3CHAR, _TICKER_HASHTAG, _TICKER_PAREN, _TICKER_COLON]:
        for match in pattern.finditer(text):
            code = match.group(1) if match.lastindex else match.group(0)
            code = code.upper()
            if code in VALID_TICKERS:
                found.add(code)
            elif code not in NON_STOCK_WORDS and len(code) == 3:
                # Include unknown 3-char codes if not obvious noise
                found.add(code)

    # Strategy 2: Company name matching
    for name, ticker in COMPANY_TO_TICKER.items():
        if name in text_lower:
            if ticker not in ("VNINDEX", "VN30", "HN30", "HOSE", "HNX", "UPCOM"):
                found.add(ticker)

    # Strategy 3: 2-char and 4-char tickers in specific patterns
    # Match "cổ phiếu XYZ" or "mã XYZ"
    cp_pattern = re.compile(r'(?:cổ phiếu|mã|ticker|mã ck)\s+([A-Z]{2,4})', re.IGNORECASE)
    for match in cp_pattern.finditer(text):
        code = match.group(1).upper()
        if code in VALID_TICKERS:
            found.add(code)

    # Remove any remaining noise
    found -= NON_STOCK_WORDS
    found -= {"VNINDEX", "VN30", "HN30", "HOSE", "HNX", "UPCOM"}

    return sorted(found)
