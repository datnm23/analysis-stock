from fastapi import APIRouter

router = APIRouter(tags=["health"])

# These will be set by main.py
_model = None


def set_model(model):
    global _model
    _model = model


def get_model():
    return _model


@router.get("/health")
async def health_check():
    """Health check endpoint."""
    model = get_model()

    return {
        "status": "healthy" if model and model.is_loaded else "unhealthy",
        "model_loaded": model.is_loaded if model else False,
        "model_name": "vinai/phobert-base",
        "memory_usage_mb": model.get_memory_usage() if model else 0
    }


@router.get("/ready")
async def readiness_check():
    """Readiness check for Kubernetes."""
    model = get_model()

    if not model or not model.is_loaded:
        return {"ready": False}

    return {"ready": True}
