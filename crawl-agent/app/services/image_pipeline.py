"""Image generation via Vertex AI Imagen + upload to S3/MinIO or Google Drive."""
import asyncio
import base64
import json
import logging
import re
from typing import Optional

import httpx

logger = logging.getLogger(__name__)

_CLAUDE_URL = "https://api.anthropic.com/v1/messages"
_ANTHROPIC_VERSION = "2023-06-01"
_VERTEX_TEMPLATE = (
    "https://{location}-aiplatform.googleapis.com/v1/projects/{project}"
    "/locations/{location}/publishers/google/models/{model}:predict"
)

_IMAGE_PROMPT_INSTRUCTIONS = """\
Tạo một câu mô tả ảnh minh họa ngắn gọn cho bài phân tích cổ phiếu {symbol}.
Phong cách: tối giản, màu xanh tài chính, biểu đồ abstract không hiển thị số liệu thực.
Chỉ trả về mô tả ảnh bằng tiếng Anh, tối đa 2 câu, không giải thích thêm.

Tóm tắt bài: {summary}
"""


class ImagePipeline:
    def __init__(
        self,
        anthropic_api_key: str = "",
        claude_model: str = "claude-haiku-4-5-20251001",
        # Vertex AI Imagen
        vertex_credentials_json: str = "",
        vertex_project_id: str = "",
        vertex_location: str = "us-central1",
        vertex_image_model: str = "imagen-3.0-generate-001",
        # Storage
        storage_backend: str = "s3",
        # S3/MinIO
        s3_endpoint: str = "",
        s3_bucket: str = "blog-images",
        s3_access_key: str = "",
        s3_secret_key: str = "",
        s3_public_url: str = "",
        # Google Drive
        gdrive_credentials_json: str = "",
        gdrive_folder_id: str = "",
    ):
        self._anthropic_key = anthropic_api_key
        self._claude_model = claude_model
        self._vertex_credentials_json = vertex_credentials_json
        self._vertex_project_id = vertex_project_id
        self._vertex_location = vertex_location
        self._vertex_image_model = vertex_image_model
        self._storage_backend = storage_backend
        self._s3_endpoint = s3_endpoint
        self._s3_bucket = s3_bucket
        self._s3_access_key = s3_access_key
        self._s3_secret_key = s3_secret_key
        self._s3_public_url = s3_public_url.rstrip("/")
        self._gdrive_credentials_json = gdrive_credentials_json
        self._gdrive_folder_id = gdrive_folder_id

    async def build_image_prompt(self, symbol: str, summary: str) -> str:
        """Use Claude Haiku to generate a concise English image prompt."""
        if not self._anthropic_key:
            return f"Abstract stock market chart for {symbol}, minimalist blue financial style, no numbers"
        prompt = _IMAGE_PROMPT_INSTRUCTIONS.format(symbol=symbol, summary=summary[:300])
        try:
            async with httpx.AsyncClient(timeout=15) as client:
                resp = await client.post(
                    _CLAUDE_URL,
                    headers={
                        "x-api-key": self._anthropic_key,
                        "anthropic-version": _ANTHROPIC_VERSION,
                        "content-type": "application/json",
                    },
                    json={
                        "model": self._claude_model,
                        "max_tokens": 120,
                        "messages": [{"role": "user", "content": prompt}],
                    },
                )
                if resp.status_code == 200:
                    return resp.json()["content"][0]["text"].strip()
        except Exception as exc:
            logger.warning("image prompt generation failed: %s", exc)
        return f"Abstract financial chart for {symbol} stock analysis, blue gradient, minimalist"

    async def generate_image(self, image_prompt: str) -> Optional[bytes]:
        """Generate PNG bytes via Vertex AI Imagen."""
        if not self._vertex_project_id or not self._vertex_credentials_json:
            logger.warning("Vertex AI not configured (need VERTEX_PROJECT_ID + VERTEX_CREDENTIALS_JSON)")
            return None
        try:
            creds_dict = json.loads(self._vertex_credentials_json)
        except json.JSONDecodeError as exc:
            logger.warning("VERTEX_CREDENTIALS_JSON invalid: %s", exc)
            return None

        url = _VERTEX_TEMPLATE.format(
            location=self._vertex_location,
            project=self._vertex_project_id,
            model=self._vertex_image_model,
        )
        try:
            from google.auth.transport.requests import Request
            from google.oauth2 import service_account

            def _get_token() -> str:
                creds = service_account.Credentials.from_service_account_info(
                    creds_dict,
                    scopes=["https://www.googleapis.com/auth/cloud-platform"],
                )
                creds.refresh(Request())
                return creds.token

            token = await asyncio.to_thread(_get_token)

            async with httpx.AsyncClient(timeout=60) as client:
                resp = await client.post(
                    url,
                    headers={"Authorization": f"Bearer {token}"},
                    json={
                        "instances": [{"prompt": image_prompt}],
                        "parameters": {"sampleCount": 1},
                    },
                )
                if resp.status_code == 200:
                    predictions = resp.json().get("predictions", [])
                    if predictions:
                        return base64.b64decode(predictions[0]["bytesBase64Encoded"])
                    logger.warning("Vertex AI: no predictions returned")
                else:
                    logger.warning("Vertex AI image API %d", resp.status_code)
        except Exception as exc:
            logger.warning("Vertex AI image generation failed: %s", exc)
        return None

    async def upload_image(self, key: str, image_bytes: bytes) -> Optional[str]:
        """Upload image bytes using the configured storage backend."""
        if self._storage_backend == "gdrive":
            return await self._upload_gdrive(key, image_bytes)
        return await self._upload_s3(key, image_bytes)

    async def _upload_s3(self, key: str, image_bytes: bytes) -> Optional[str]:
        """Upload PNG bytes to S3/MinIO, return public URL."""
        if not self._s3_endpoint or not self._s3_access_key:
            logger.warning("S3 not configured, skipping image upload")
            return None
        try:
            import boto3
            from botocore.client import Config

            def _do_upload() -> None:
                s3 = boto3.client(
                    "s3",
                    endpoint_url=self._s3_endpoint,
                    aws_access_key_id=self._s3_access_key,
                    aws_secret_access_key=self._s3_secret_key,
                    config=Config(signature_version="s3v4"),
                )
                s3.put_object(
                    Bucket=self._s3_bucket,
                    Key=key,
                    Body=image_bytes,
                    ContentType="image/png",
                )

            await asyncio.to_thread(_do_upload)
            base_url = self._s3_public_url or f"{self._s3_endpoint}/{self._s3_bucket}"
            return f"{base_url}/{key}"
        except Exception as exc:
            logger.warning("S3 upload failed: %s", exc)
        return None

    async def _upload_gdrive(self, key: str, image_bytes: bytes) -> Optional[str]:
        """Upload PNG bytes to Google Drive, return public view URL."""
        if not self._gdrive_credentials_json:
            logger.warning("GDRIVE_CREDENTIALS_JSON not set, skipping upload")
            return None
        try:
            creds_dict = json.loads(self._gdrive_credentials_json)
        except json.JSONDecodeError as exc:
            logger.warning("GDRIVE_CREDENTIALS_JSON is not valid JSON: %s", exc)
            return None

        # Filename = last path component (e.g. "articles/VNM/20260418.png" → "VNM-20260418.png")
        filename = key.replace("/", "-")

        try:
            from google.oauth2 import service_account
            from googleapiclient.discovery import build
            from googleapiclient.http import MediaInMemoryUpload

            def _do_upload() -> str:
                creds = service_account.Credentials.from_service_account_info(
                    creds_dict,
                    scopes=["https://www.googleapis.com/auth/drive.file"],
                )
                service = build("drive", "v3", credentials=creds, cache_discovery=False)

                metadata: dict = {"name": filename}
                if self._gdrive_folder_id:
                    metadata["parents"] = [self._gdrive_folder_id]

                media = MediaInMemoryUpload(image_bytes, mimetype="image/png", resumable=False)
                file = service.files().create(
                    body=metadata,
                    media_body=media,
                    fields="id",
                ).execute()

                file_id = file["id"]
                # Grant public read access so the URL works without auth
                service.permissions().create(
                    fileId=file_id,
                    body={"type": "anyone", "role": "reader"},
                ).execute()

                return f"https://drive.google.com/uc?export=view&id={file_id}"

            return await asyncio.to_thread(_do_upload)
        except Exception as exc:
            logger.warning("Google Drive upload failed: %s", exc)
        return None

    @staticmethod
    def _safe_symbol(symbol: str) -> str:
        """Strip non-alphanumeric chars to prevent path traversal in storage keys."""
        return re.sub(r"[^a-zA-Z0-9]", "", symbol)[:10]

    async def generate_and_upload(self, symbol: str, summary: str, date_str: str) -> Optional[str]:
        """Full pipeline: Claude builds prompt → Gemini generates image → upload → URL."""
        image_prompt = await self.build_image_prompt(symbol, summary)
        image_bytes = await self.generate_image(image_prompt)
        if not image_bytes:
            return None
        safe_sym = self._safe_symbol(symbol)
        return await self.upload_image(f"articles/{safe_sym}/{date_str}.png", image_bytes)
