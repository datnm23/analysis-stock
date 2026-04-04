"""
PhoBERT → ONNX Export + INT8 Quantization

Exports the PhoBERT sentiment model to ONNX format with optional
INT8 dynamic quantization to reduce memory from ~4GB → ~1GB.

Usage:
    python -m scripts.export_onnx                          # FP32 export
    python -m scripts.export_onnx --quantize               # INT8 quantized
    python -m scripts.export_onnx --output models/phobert  # Custom output dir
"""

import argparse
import logging
import os
import sys
from pathlib import Path

import numpy as np

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
logger = logging.getLogger(__name__)


def export_to_onnx(
    model_name: str = "vinai/phobert-base",
    cache_dir: str = ".cache",
    output_dir: str = "models/onnx",
    quantize: bool = False,
    opset_version: int = 14,
) -> str:
    """Export PhoBERT to ONNX format.

    Returns the path to the exported ONNX model.
    """
    try:
        import torch
        from transformers import AutoModelForSequenceClassification, AutoTokenizer
    except ImportError:
        logger.error("torch and transformers are required. pip install torch transformers")
        sys.exit(1)

    os.makedirs(output_dir, exist_ok=True)
    onnx_path = os.path.join(output_dir, "phobert_sentiment.onnx")

    # ---- Load model & tokenizer ----
    logger.info("Loading model %s …", model_name)
    tokenizer = AutoTokenizer.from_pretrained(model_name, cache_dir=cache_dir)
    model = AutoModelForSequenceClassification.from_pretrained(
        model_name,
        cache_dir=cache_dir,
        num_labels=3,  # positive / negative / neutral
    )
    model.eval()

    # ---- Dummy input for tracing ----
    dummy_text = "Cổ phiếu VNM tăng mạnh hôm nay"
    inputs = tokenizer(
        dummy_text,
        return_tensors="pt",
        max_length=256,
        padding="max_length",
        truncation=True,
    )

    input_ids = inputs["input_ids"]
    attention_mask = inputs["attention_mask"]

    # ---- Export ----
    logger.info("Exporting to ONNX (opset=%d) → %s", opset_version, onnx_path)

    torch.onnx.export(
        model,
        (input_ids, attention_mask),
        onnx_path,
        export_params=True,
        opset_version=opset_version,
        do_constant_folding=True,
        input_names=["input_ids", "attention_mask"],
        output_names=["logits"],
        dynamic_axes={
            "input_ids": {0: "batch", 1: "seq_len"},
            "attention_mask": {0: "batch", 1: "seq_len"},
            "logits": {0: "batch"},
        },
    )

    fp32_size = os.path.getsize(onnx_path) / (1024 * 1024)
    logger.info("FP32 ONNX model: %.1f MB", fp32_size)

    # ---- Optional INT8 quantization ----
    if quantize:
        try:
            from onnxruntime.quantization import quantize_dynamic, QuantType
        except ImportError:
            logger.error("onnxruntime is required for quantization. pip install onnxruntime")
            sys.exit(1)

        quant_path = os.path.join(output_dir, "phobert_sentiment_int8.onnx")
        logger.info("Quantizing to INT8 → %s", quant_path)

        try:
            quantize_dynamic(
                model_input=onnx_path,
                model_output=quant_path,
                weight_type=QuantType.QInt8,
                per_channel=True,
                reduce_range=True,
            )

            int8_size = os.path.getsize(quant_path) / (1024 * 1024)
            reduction = (1 - int8_size / fp32_size) * 100
            logger.info(
                "INT8 ONNX model: %.1f MB (%.0f%% reduction)",
                int8_size,
                reduction,
            )
            onnx_path = quant_path
        except Exception as e:
            logger.warning(
                "⚠️  INT8 quantization failed: %s. "
                "This is expected for non-fine-tuned models. "
                "Using FP32 model instead. Fine-tune first, then re-run with --quantize.",
                e,
            )

    # ---- Verify ----
    logger.info("Verifying ONNX model …")
    try:
        import onnxruntime as ort

        sess = ort.InferenceSession(onnx_path)
        result = sess.run(
            None,
            {
                "input_ids": input_ids.numpy(),
                "attention_mask": attention_mask.numpy(),
            },
        )
        logits = result[0]
        probs = _softmax(logits[0])
        labels = ["negative", "neutral", "positive"]
        pred = labels[int(np.argmax(probs))]
        logger.info(
            "Verification OK: '%s' → %s (%.1f%%)",
            dummy_text,
            pred,
            max(probs) * 100,
        )
    except ImportError:
        logger.warning("onnxruntime not installed — skipping verification")

    # Save tokenizer alongside model for inference
    tokenizer.save_pretrained(output_dir)
    logger.info("Tokenizer saved to %s", output_dir)

    logger.info("✅ Export complete: %s", onnx_path)
    return onnx_path


def _softmax(x):
    e = np.exp(x - np.max(x))
    return e / e.sum()


def main():
    parser = argparse.ArgumentParser(description="Export PhoBERT to ONNX")
    parser.add_argument(
        "--model", default="vinai/phobert-base", help="HuggingFace model name"
    )
    parser.add_argument("--cache-dir", default=".cache", help="Model cache directory")
    parser.add_argument(
        "--output", default="models/onnx", help="Output directory for ONNX model"
    )
    parser.add_argument(
        "--quantize", action="store_true", help="Apply INT8 dynamic quantization"
    )
    parser.add_argument(
        "--opset", type=int, default=14, help="ONNX opset version"
    )

    args = parser.parse_args()

    export_to_onnx(
        model_name=args.model,
        cache_dir=args.cache_dir,
        output_dir=args.output,
        quantize=args.quantize,
        opset_version=args.opset,
    )


if __name__ == "__main__":
    main()
