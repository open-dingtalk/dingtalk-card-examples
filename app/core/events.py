from typing import Callable

from fastapi import FastAPI
from loguru import logger
from app.core.config import AppSettings
from app.stream.events import run_dingtalk_stream


def create_start_app_handler(
    app: FastAPI,
    settings: AppSettings,
) -> Callable:
    async def start_app() -> None:
        client = await run_dingtalk_stream(settings)
        app.state.dingtalk_client = client

    return start_app


def create_stop_app_handler(app: FastAPI) -> Callable:
    @logger.catch
    async def stop_app() -> None:
        pass

    return stop_app
