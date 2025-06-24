from loguru import logger
from http import HTTPStatus
from typing import Callable
from dashscope import Generation
from app.stream.utils import convert_json_values_to_string
from dingtalk_stream import (
    ChatbotHandler,
    ChatbotMessage,
    CallbackMessage,
    AICardReplier,
    AckMessage,
)


async def call_with_stream(request_content: str, callback: Callable[[str], None]):
    messages = [{"role": "user", "content": request_content}]
    responses = Generation.call(
        Generation.Models.qwen_turbo,
        messages=messages,
        result_format="message",
        stream=True,
        incremental_output=True,
    )
    full_content = ""
    length = 0
    for response in responses:
        if response.status_code == HTTPStatus.OK:
            full_content += response.output.choices[0]["message"]["content"]
            full_content_length = len(full_content)
            if full_content_length - length > 20:
                await callback(full_content)
                logger.debug(
                    f"调用流式更新接口更新内容：current_length: {length}, next_length: {full_content_length}"
                )
                length = full_content_length
        else:
            raise Exception(
                f"Request id: {response.request_id}, Status code: {response.status_code}, error code: {response.code}, error message: {response.message}"
            )
    await callback(full_content)
    logger.info(
        f"Request Content: {request_content}\nFull response: {full_content}\nFull response length: {len(full_content)}"
    )
    return full_content


class CardBotHandler(ChatbotHandler):
    def __init__(self, logger=logger):
        super(ChatbotHandler, self).__init__()
        if logger:
            self.logger = logger

    async def process(self, callback: CallbackMessage):
        incoming_message = ChatbotMessage.from_dict(callback.data)
        self.logger.info(incoming_message.sender_staff_id)  # 给机器人发消息的用户 userId
        self.logger.info(incoming_message.conversation_id)  # 机器人在群里收到消息的群 cid
        content = (incoming_message.text.content or "").strip()
        self.logger.info(f"received message: {content}")

        card_template_id = "d70f026e-7148-4479-b089-8dcf60289b9d.schema"  # 导出的模板在 app/card_templates 目录下的 「智能回复.json」
        content_key = "content"
        card_data = {content_key: "", "config": {"autoLayout": True}}
        card_instance = AICardReplier(self.dingtalk_client, incoming_message)

        # 先投放卡片: https://open.dingtalk.com/document/orgapp/create-and-deliver-cards
        card_instance_id = await card_instance.async_create_and_deliver_card(
            card_template_id,
            convert_json_values_to_string(card_data),
        )

        # 再流式更新卡片: https://open.dingtalk.com/document/isvapp/api-streamingupdate
        async def callback(content_value: str):
            return await card_instance.async_streaming(
                card_instance_id,
                content_key=content_key,
                content_value=content_value,
                append=False,
                finished=False,
                failed=False,
            )

        try:
            full_content_value = await call_with_stream(
                incoming_message.text.content, callback
            )
            await card_instance.async_streaming(
                card_instance_id,
                content_key=content_key,
                content_value=full_content_value,
                append=False,
                finished=True,
                failed=False,
            )
        except Exception as e:
            self.logger.exception(e)
            await card_instance.async_streaming(
                card_instance_id,
                content_key=content_key,
                content_value=full_content_value,
                append=False,
                finished=False,
                failed=True,
            )

        self.logger.info(f"reply ai card {card_template_id} {card_instance_id}")

        return AckMessage.STATUS_OK, "OK"
