"""
PhoBERT Fine-Tuning Pipeline for Vietnamese Stock Sentiment

Fine-tunes vinai/phobert-base on labeled Vietnamese financial text data
for 3-class sentiment classification (positive/negative/neutral).

Data format (CSV):
    text,label
    "Cổ phiếu VNM tăng mạnh",positive
    "Thị trường lao dốc",negative

Usage:
    python -m scripts.fine_tune --data data/labeled_sentiment.csv
    python -m scripts.fine_tune --data data/labeled_sentiment.csv --epochs 5 --lr 2e-5
"""

import argparse
import json
import logging
import os
import sys
from pathlib import Path

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
logger = logging.getLogger(__name__)

LABEL_MAP = {"negative": 0, "neutral": 1, "positive": 2}
REVERSE_LABEL = {v: k for k, v in LABEL_MAP.items()}


def load_dataset(csv_path: str, test_split: float = 0.15):
    """Load CSV dataset and split into train/test."""
    import pandas as pd
    from sklearn.model_selection import train_test_split

    df = pd.read_csv(csv_path)

    # Validate columns
    if "text" not in df.columns or "label" not in df.columns:
        raise ValueError("CSV must have 'text' and 'label' columns")

    # Map labels
    df["label_id"] = df["label"].str.lower().map(LABEL_MAP)
    unknown = df[df["label_id"].isna()]
    if len(unknown) > 0:
        logger.warning("Dropping %d rows with unknown labels: %s", len(unknown), unknown["label"].unique())
        df = df.dropna(subset=["label_id"])

    df["label_id"] = df["label_id"].astype(int)

    train_df, test_df = train_test_split(
        df, test_size=test_split, stratify=df["label_id"], random_state=42
    )

    logger.info(
        "Dataset: %d total (%d train, %d test). Distribution: %s",
        len(df), len(train_df), len(test_df),
        df["label"].value_counts().to_dict(),
    )
    return train_df, test_df


def create_torch_dataset(df, tokenizer, max_length=256):
    """Create a PyTorch Dataset from a DataFrame."""
    import torch
    from torch.utils.data import Dataset

    class SentimentDataset(Dataset):
        def __init__(self, texts, labels, tokenizer, max_length):
            self.encodings = tokenizer(
                texts.tolist(),
                max_length=max_length,
                padding="max_length",
                truncation=True,
                return_tensors="pt",
            )
            self.labels = torch.tensor(labels.tolist(), dtype=torch.long)

        def __len__(self):
            return len(self.labels)

        def __getitem__(self, idx):
            return {
                "input_ids": self.encodings["input_ids"][idx],
                "attention_mask": self.encodings["attention_mask"][idx],
                "labels": self.labels[idx],
            }

    return SentimentDataset(df["text"], df["label_id"], tokenizer, max_length)


def fine_tune(
    data_path: str,
    model_name: str = "vinai/phobert-base",
    cache_dir: str = ".cache",
    output_dir: str = "models/fine_tuned",
    epochs: int = 3,
    batch_size: int = 16,
    learning_rate: float = 2e-5,
    max_length: int = 256,
    warmup_ratio: float = 0.1,
):
    """Run the fine-tuning pipeline."""
    try:
        import torch
        from transformers import (
            AutoModelForSequenceClassification,
            AutoTokenizer,
            Trainer,
            TrainingArguments,
        )
        from sklearn.metrics import classification_report, accuracy_score
        import numpy as np
    except ImportError as e:
        logger.error("Missing dependency: %s. Install: pip install torch transformers scikit-learn pandas", e)
        sys.exit(1)

    # ---- Load data ----
    train_df, test_df = load_dataset(data_path)

    if len(train_df) < 100:
        logger.warning("⚠️  Training set has only %d samples. Recommend ≥ 5,000 for good results.", len(train_df))

    # ---- Load model & tokenizer ----
    logger.info("Loading %s …", model_name)
    tokenizer = AutoTokenizer.from_pretrained(model_name, cache_dir=cache_dir)
    model = AutoModelForSequenceClassification.from_pretrained(
        model_name,
        cache_dir=cache_dir,
        num_labels=3,
        id2label=REVERSE_LABEL,
        label2id=LABEL_MAP,
    )

    # Apply slang preprocessing if available
    try:
        from app.services.sentiment_analyzer import SentimentAnalyzer
        from app.models.phobert import PhoBERTSentiment

        phobert = PhoBERTSentiment(model_name=model_name, cache_dir=cache_dir)
        analyzer = SentimentAnalyzer(phobert)
        logger.info("Applying slang dict preprocessing to training data …")
        train_df["text"] = train_df["text"].apply(analyzer.preprocess_text)
        test_df["text"] = test_df["text"].apply(analyzer.preprocess_text)
    except Exception as e:
        logger.warning("Could not apply slang preprocessing: %s", e)

    # ---- Create datasets ----
    train_dataset = create_torch_dataset(train_df, tokenizer, max_length)
    test_dataset = create_torch_dataset(test_df, tokenizer, max_length)

    # ---- Training config ----
    os.makedirs(output_dir, exist_ok=True)

    training_args = TrainingArguments(
        output_dir=output_dir,
        eval_strategy="epoch",
        save_strategy="epoch",
        learning_rate=learning_rate,
        per_device_train_batch_size=batch_size,
        per_device_eval_batch_size=batch_size * 2,
        num_train_epochs=epochs,
        warmup_ratio=warmup_ratio,
        weight_decay=0.01,
        load_best_model_at_end=True,
        metric_for_best_model="accuracy",
        greater_is_better=True,
        logging_steps=50,
        save_total_limit=2,
        fp16=torch.cuda.is_available(),
        report_to="none",
    )

    def compute_metrics(eval_pred):
        logits, labels = eval_pred
        predictions = np.argmax(logits, axis=-1)
        return {"accuracy": accuracy_score(labels, predictions)}

    # ---- Train ----
    trainer = Trainer(
        model=model,
        args=training_args,
        train_dataset=train_dataset,
        eval_dataset=test_dataset,
        compute_metrics=compute_metrics,
    )

    logger.info("🚀 Starting fine-tuning: %d epochs, lr=%.1e, batch=%d", epochs, learning_rate, batch_size)
    trainer.train()

    # ---- Evaluate ----
    logger.info("Evaluating on test set …")
    predictions = trainer.predict(test_dataset)
    preds = np.argmax(predictions.predictions, axis=-1)
    labels = predictions.label_ids

    report = classification_report(
        labels, preds,
        target_names=["negative", "neutral", "positive"],
        output_dict=True,
    )
    logger.info("\n%s", classification_report(labels, preds, target_names=["negative", "neutral", "positive"]))

    # ---- Save ----
    best_dir = os.path.join(output_dir, "best")
    model.save_pretrained(best_dir)
    tokenizer.save_pretrained(best_dir)

    # Save metrics
    metrics_path = os.path.join(output_dir, "metrics.json")
    with open(metrics_path, "w") as f:
        json.dump({
            "accuracy": float(report["accuracy"]),
            "macro_f1": float(report["macro avg"]["f1-score"]),
            "per_class": {
                k: {
                    "precision": float(report[k]["precision"]),
                    "recall": float(report[k]["recall"]),
                    "f1": float(report[k]["f1-score"]),
                    "support": int(report[k]["support"]),
                }
                for k in ["negative", "neutral", "positive"]
            },
            "train_samples": len(train_df),
            "test_samples": len(test_df),
            "epochs": epochs,
            "learning_rate": learning_rate,
        }, f, indent=2)

    logger.info("✅ Fine-tuning complete!")
    logger.info("   Model saved: %s", best_dir)
    logger.info("   Metrics saved: %s", metrics_path)
    logger.info("   Accuracy: %.2f%%", report["accuracy"] * 100)
    logger.info("   Macro F1: %.2f%%", report["macro avg"]["f1-score"] * 100)

    return best_dir


def main():
    parser = argparse.ArgumentParser(description="Fine-tune PhoBERT for VN stock sentiment")
    parser.add_argument("--data", required=True, help="Path to labeled CSV (text,label)")
    parser.add_argument("--model", default="vinai/phobert-base", help="Base model")
    parser.add_argument("--cache-dir", default=".cache")
    parser.add_argument("--output", default="models/fine_tuned")
    parser.add_argument("--epochs", type=int, default=3)
    parser.add_argument("--batch-size", type=int, default=16)
    parser.add_argument("--lr", type=float, default=2e-5)
    parser.add_argument("--max-length", type=int, default=256)

    args = parser.parse_args()

    fine_tune(
        data_path=args.data,
        model_name=args.model,
        cache_dir=args.cache_dir,
        output_dir=args.output,
        epochs=args.epochs,
        batch_size=args.batch_size,
        learning_rate=args.lr,
        max_length=args.max_length,
    )


if __name__ == "__main__":
    main()
