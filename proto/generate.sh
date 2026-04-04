#!/bin/bash
# Generate gRPC stubs from proto definitions
#
# Prerequisites:
#   Go:     go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
#           go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
#   Python: pip install grpcio-tools
#
# Usage:   ./proto/generate.sh

set -euo pipefail

PROTO_DIR="$(dirname "$0")"
PROJECT_ROOT="$(dirname "$PROTO_DIR")"

echo "=== Generating gRPC stubs ==="

# ---- Go stubs ----
GO_OUT="$PROJECT_ROOT/go-services/proto"
mkdir -p "$GO_OUT/sentiment"

echo "📦 Generating Go stubs..."
protoc \
    -I "$PROTO_DIR" \
    --go_out="$GO_OUT" \
    --go_opt=paths=source_relative \
    --go-grpc_out="$GO_OUT" \
    --go-grpc_opt=paths=source_relative \
    "$PROTO_DIR/sentiment.proto"

echo "   → $GO_OUT/sentiment/"

# ---- Python stubs ----
PY_OUT="$PROJECT_ROOT/python-sentiment/app/proto"
mkdir -p "$PY_OUT"

echo "📦 Generating Python stubs..."
python3 -m grpc_tools.protoc \
    -I "$PROTO_DIR" \
    --python_out="$PY_OUT" \
    --grpc_python_out="$PY_OUT" \
    "$PROTO_DIR/sentiment.proto"

# Fix Python imports (relative)
sed -i 's/^import sentiment_pb2/from . import sentiment_pb2/' "$PY_OUT/sentiment_pb2_grpc.py" 2>/dev/null || true

# Create __init__.py
cat > "$PY_OUT/__init__.py" << 'EOF'
"""Generated gRPC stubs for sentiment service."""
EOF

echo "   → $PY_OUT/"

echo ""
echo "✅ Done! Generated stubs for Go and Python."
echo ""
echo "Next steps:"
echo "  1. In Go: import sentimentpb \"vnstock-hybrid/proto/sentiment\""
echo "  2. In Python: from app.proto import sentiment_pb2, sentiment_pb2_grpc"
