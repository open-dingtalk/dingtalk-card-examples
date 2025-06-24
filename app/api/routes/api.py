from fastapi import APIRouter

from app.api.routes import ping
from app.api.routes import notice

router = APIRouter()
router.include_router(ping.router, tags=["ping"], prefix="/ping")
router.include_router(notice.router, tags=["notice"], prefix="/notice")

