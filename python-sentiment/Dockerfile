FROM python:3.11-slim

WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Copy requirements first for caching
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy application code
COPY app/ ./app/
COPY scripts/ ./scripts/

# Create directories
RUN mkdir -p .cache models/onnx

# ---------------------------------------------------------------------------
# ONNX Model Export (optional — set EXPORT_ONNX=true to enable)
# ---------------------------------------------------------------------------
# When EXPORT_ONNX is enabled during build, the model is downloaded,
# exported to ONNX INT8, and the tokenizer is saved alongside it.
#
# Build with:   docker build --build-arg EXPORT_ONNX=true -t vnstock-sentiment .
# Run with:     docker run -e USE_ONNX=true vnstock-sentiment
#
# Without EXPORT_ONNX, the service runs in PyTorch mode (larger memory footprint).
# You can also mount a pre-exported model: -v ./models/onnx:/app/models/onnx
# ---------------------------------------------------------------------------

ARG EXPORT_ONNX=false
RUN if [ "$EXPORT_ONNX" = "true" ]; then \
        echo "=== Exporting PhoBERT to ONNX INT8 ===" && \
        python -m scripts.export_onnx --quantize --output models/onnx && \
        echo "=== ONNX export complete ===" ; \
    fi

# Set environment variables
ENV PYTHONUNBUFFERED=1
ENV PORT=8000

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=120s --retries=3 \
    CMD curl -f http://localhost:8000/health || exit 1

EXPOSE 8000

CMD ["python", "-m", "app.main"]
