import { NextRequest, NextResponse } from "next/server";
import { promises as fs } from "fs";
import path from "path";

const DATA_DIR = path.join(process.cwd(), "data");
const SUBSCRIBERS_FILE = path.join(DATA_DIR, "subscribers.json");

interface Subscriber {
  email: string;
  subscribedAt: string;
}

async function readSubscribers(): Promise<Subscriber[]> {
  try {
    const raw = await fs.readFile(SUBSCRIBERS_FILE, "utf-8");
    return JSON.parse(raw);
  } catch {
    return [];
  }
}

async function writeSubscribers(list: Subscriber[]): Promise<void> {
  await fs.mkdir(DATA_DIR, { recursive: true });
  await fs.writeFile(SUBSCRIBERS_FILE, JSON.stringify(list, null, 2));
}

const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

export async function POST(req: NextRequest) {
  try {
    const body = await req.json();
    const email: string = (body.email ?? "").trim().toLowerCase();

    if (!email || !EMAIL_RE.test(email) || email.length > 254) {
      return NextResponse.json({ error: "Email không hợp lệ" }, { status: 400 });
    }

    // Forward to external webhook if configured
    const webhookUrl = process.env.NEWSLETTER_WEBHOOK_URL;
    if (webhookUrl) {
      const r = await fetch(webhookUrl, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email }),
      });
      if (!r.ok) return NextResponse.json({ error: "Lỗi đăng ký" }, { status: 500 });
      return NextResponse.json({ success: true });
    }

    // Fallback: persist to local JSON file
    const subscribers = await readSubscribers();
    if (subscribers.some((s) => s.email === email)) {
      return NextResponse.json({ success: true }); // idempotent
    }
    subscribers.push({ email, subscribedAt: new Date().toISOString() });
    await writeSubscribers(subscribers);

    return NextResponse.json({ success: true });
  } catch {
    return NextResponse.json({ error: "Lỗi server" }, { status: 500 });
  }
}
