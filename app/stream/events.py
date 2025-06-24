import asyncio
from dingtalk_stream import Credential, ChatbotMessage, CallbackHandler, DingTalkStreamClient
from app.stream.chatbot_handler import CardBotHandler
from app.stream.card_callback_handler import CardCallbackHandler
from app.core.config import AppSettings


async def run_dingtalk_stream(settings: AppSettings):
    credential = Credential(
        settings.client_id, settings.client_secret.get_secret_value()
    )
    client = DingTalkStreamClient(credential)
    client.register_callback_handler(ChatbotMessage.TOPIC, CardBotHandler())
    client.register_callback_handler(
        CallbackHandler.TOPIC_CARD_CALLBACK, CardCallbackHandler()
    )
    asyncio.create_task(client.start())
    return client
