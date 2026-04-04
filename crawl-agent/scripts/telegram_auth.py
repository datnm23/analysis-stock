#!/usr/bin/env python3
"""
Telegram Session Authentication Script.

Telethon requires a one-time interactive login (phone number + OTP code)
to create a session file. This script handles that setup.

Usage:
    cd /home/datnm/projects/analysis-stock/crawl-agent
    python3 -m scripts.telegram_auth

After authentication, a session file (vnstock_crawl.session) is created.
Then set ENABLE_TELEGRAM=true in .env to activate the scraper.
"""

import asyncio
import os
import sys

# Add parent dir so we can import from app/
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))


async def main():
    try:
        from telethon import TelegramClient
    except ImportError:
        print("ERROR: telethon not installed. Run: pip install telethon")
        sys.exit(1)

    # Load from env or prompt
    api_id = os.environ.get("TELEGRAM_APP_API_ID")
    api_hash = os.environ.get("TELEGRAM_APP_API_HASH")

    if not api_id or not api_hash:
        print("=" * 50)
        print("Telegram API Credentials Setup")
        print("Get yours at: https://my.telegram.org → API Development Tools")
        print("=" * 50)
        api_id = input("Enter API ID: ").strip()
        api_hash = input("Enter API Hash: ").strip()

    session_name = os.environ.get("TELEGRAM_SESSION_NAME", "vnstock_crawl")
    print(f"\nSession name: {session_name}")
    print("A .session file will be created in the current directory.\n")

    client = TelegramClient(session_name, int(api_id), api_hash)

    print("Connecting to Telegram...")
    await client.start()

    me = await client.get_me()
    print(f"\n✅ Authenticated as: {me.first_name} (@{me.username})")
    print(f"   Phone: {me.phone}")
    print(f"   Session file: {session_name}.session")

    # Test channel access
    channels = os.environ.get(
        "TELEGRAM_CHANNELS",
        "chungkhoanUG,ChungkhoanGalaxy,finbotrealtimenews"
    ).split(",")

    print(f"\nTesting access to {len(channels)} channels:")
    for ch in channels:
        ch = ch.strip()
        try:
            entity = await client.get_entity(ch)
            print(f"  ✅ @{ch} — {entity.title} ({entity.participants_count or '?'} members)")
        except Exception as e:
            print(f"  ❌ @{ch} — {e}")

    await client.disconnect()

    print("\n" + "=" * 50)
    print("Setup complete! Next steps:")
    print("1. Set ENABLE_TELEGRAM=true in .env")
    print("2. Copy the .session file to crawl-agent container if using Docker")
    print("3. Restart crawl-agent: docker compose restart crawl-agent")
    print("=" * 50)


if __name__ == "__main__":
    asyncio.run(main())
