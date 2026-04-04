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

# Create data directory for scheduler output
RUN mkdir -p data/scheduled

# Set environment variables
ENV PYTHONUNBUFFERED=1
ENV PORT=8085

# Health check
HEALTHCHECK --interval=30s --timeout=10s --retries=3 \
    CMD curl -f http://localhost:8085/health || exit 1

EXPOSE 8085

CMD ["python", "-m", "app.main"]
