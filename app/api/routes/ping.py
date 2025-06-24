from fastapi import APIRouter

router = APIRouter()


@router.get("", response_model=str, name="ping")
async def ping() -> str:
    return "pong"
