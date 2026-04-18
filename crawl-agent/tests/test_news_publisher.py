"""Tests for NewsPublisher — uses fakeredis to avoid real Redis dependency."""

import json

import pytest
import pytest_asyncio

try:
    import fakeredis.aioredis as fakeredis
except ImportError:
    pytest.skip("fakeredis not installed", allow_module_level=True)

from app.services.news_publisher import NewsPublisher


@pytest.fixture
def redis_client():
    return fakeredis.FakeRedis(decode_responses=False)


@pytest.fixture
def publisher(redis_client):
    return NewsPublisher(redis_client, max_items=20, ttl_hours=24)


@pytest.mark.asyncio
async def test_publish_item_single_symbol(publisher, redis_client):
    item = {
        "id": "abc123",
        "title": "VNM tăng trưởng Q1",
        "content": "Doanh thu tăng 15%",
        "source": "cafef",
        "published_at": "2026-04-18T08:00:00+07:00",
        "symbols": ["VNM"],
    }
    count = await publisher.publish_item(item)

    assert count == 1
    raw = await redis_client.lrange("news:VNM:recent", 0, -1)
    assert len(raw) == 1
    data = json.loads(raw[0])
    assert data["title"] == "VNM tăng trưởng Q1"
    assert data["source"] == "cafef"
    assert data["id"] == "abc123"


@pytest.mark.asyncio
async def test_publish_item_multiple_symbols(publisher, redis_client):
    item = {
        "id": "x1",
        "title": "HPG và VNM đều tăng",
        "content": "...",
        "source": "vnexpress",
        "symbols": ["HPG", "VNM"],
    }
    count = await publisher.publish_item(item)

    assert count == 2
    assert await redis_client.llen("news:HPG:recent") == 1
    assert await redis_client.llen("news:VNM:recent") == 1


@pytest.mark.asyncio
async def test_publish_item_no_symbols(publisher, redis_client):
    count = await publisher.publish_item({"id": "x", "title": "...", "symbols": []})
    assert count == 0


@pytest.mark.asyncio
async def test_ltrim_enforced(redis_client):
    pub = NewsPublisher(redis_client, max_items=3, ttl_hours=24)

    for i in range(5):
        await pub.publish_item(
            {"id": f"n{i}", "title": f"News {i}", "content": "", "source": "s", "symbols": ["VNM"]}
        )

    assert await redis_client.llen("news:VNM:recent") == 3


@pytest.mark.asyncio
async def test_publish_batch_stats(publisher, redis_client):
    items = [
        {"id": "a", "title": "T1", "content": "", "source": "s", "symbols": ["VCB"]},
        {"id": "b", "title": "T2", "content": "", "source": "s", "symbols": ["TCB", "VCB"]},
        {"id": "c", "title": "T3", "content": "", "source": "s", "symbols": []},  # no symbols
    ]
    stats = await publisher.publish_batch(items)

    assert stats["items_published"] == 2   # 2 items had symbols
    assert stats["symbol_keys_updated"] == 3  # VCB + TCB + VCB = 3 key ops


@pytest.mark.asyncio
async def test_content_truncated(publisher, redis_client):
    long_content = "x" * 1000
    await publisher.publish_item(
        {"id": "t", "title": "y" * 300, "content": long_content, "source": "s", "symbols": ["FPT"]}
    )
    raw = await redis_client.lrange("news:FPT:recent", 0, 0)
    data = json.loads(raw[0])
    assert len(data["title"]) <= 200
    assert len(data["content"]) <= 500


@pytest.mark.asyncio
async def test_ttl_set(publisher, redis_client):
    await publisher.publish_item(
        {"id": "ttl1", "title": "T", "content": "", "source": "s", "symbols": ["SAB"]}
    )
    ttl = await redis_client.ttl("news:SAB:recent")
    assert ttl > 0
