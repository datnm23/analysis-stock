#!/bin/bash
set -e
cd /home/datnm/projects/analysis-stock/python-sentiment

echo "=== Checking Python dependencies ==="
python3 -c "
import sys
print(f'Python: {sys.version}')
try:
    import torch
    print(f'torch: {torch.__version__}')
except ImportError:
    print('torch: NOT INSTALLED')
try:
    import transformers
    print(f'transformers: {transformers.__version__}')
except ImportError:
    print('transformers: NOT INSTALLED')
try:
    import onnxruntime
    print(f'onnxruntime: {onnxruntime.__version__}')
except ImportError:
    print('onnxruntime: NOT INSTALLED')
try:
    import onnx
    print(f'onnx: {onnx.__version__}')
except ImportError:
    print('onnx: NOT INSTALLED')
"

echo ""
echo "=== Running ONNX Export ==="
python3 -m scripts.export_onnx --quantize --output models/onnx

echo ""
echo "=== Output files ==="
ls -lh models/onnx/ 2>/dev/null || echo "(no output files found)"
