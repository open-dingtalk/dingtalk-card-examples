import os
import logging
import argparse
from loguru import logger
from dingtalk_stream import AckMessage
import dingtalk_stream

from http import HTTPStatus
from dashscope import Generation

from typing import Callable


def define_options():
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "--client_id",
        dest="client_id",
        default=os.getenv("DINGTALK_APP_CLIENT_ID"),
        help="app_key or suite_key from https://open-dev.digntalk.com",
    )
    parser.add_argument(
        "--client_secret",
        dest="client_secret",
        default=os.getenv("DINGTALK_APP_CLIENT_SECRET"),
        help="app_secret or suite_secret from https://open-dev.digntalk.com",
    )
    options = parser.parse_args()
    return options


def call_with_stream(request_content: str, callback: Callable[[str], None]):
    messages = [{"role": "user", "content": request_content}]
    responses = Generation.call(
        Generation.Models.qwen_turbo,
        messages=messages,
        result_format="message",  # set the result to be "message" format.
        stream=True,  # set stream output.
        incremental_output=True,  # get streaming output incrementally.
    )
    full_content = ""  # with incrementally we need to merge output.
    for response in responses:
        if response.status_code == HTTPStatus.OK:
            full_content += response.output.choices[0]["message"]["content"]
            callback(full_content)
        else:
            raise Exception(
                f"Request id: {response.request_id}, Status code: {response.status_code}, error code: {response.code}, error message: {response.message}"
            )

    logger.info(f"Request Content: {request_content}, Full response:\n {full_content}")
    return full_content


class CardBotHandler(dingtalk_stream.ChatbotHandler):
    def __init__(self, logger: logging.Logger = logger):
        super(dingtalk_stream.ChatbotHandler, self).__init__()
        if logger:
            self.logger = logger

    async def process(self, callback: dingtalk_stream.CallbackMessage):
        incoming_message = dingtalk_stream.ChatbotMessage.from_dict(callback.data)
        self.logger.info(f"收到消息：{incoming_message}")

        if incoming_message.message_type != "text":
            self.reply_text("俺只看得懂文字喔~", incoming_message)
            return AckMessage.STATUS_OK, "OK"

        # 卡片模板 ID
        card_template_id = "8aebdfb9-28f4-4a98-98f5-396c3dde41a0.schema"  # 该模板只用于测试使用，如需投入线上使用，请导入卡片模板 json 到自己的应用下
        content_key = "content"
        card_data = {content_key: ""}
        card_instance = dingtalk_stream.AICardReplier(
            self.dingtalk_client, incoming_message
        )
        # 先投放卡片
        card_instance_id = card_instance.create_and_deliver_card(
            card_template_id, card_data
        )
        # 再流式更新卡片
        try:
            full_content_value = call_with_stream(
                incoming_message.text.content,
                lambda content_value: card_instance.streaming(
                    card_instance_id,
                    content_key=content_key,
                    content_value=content_value,
                    append=False,
                    finished=False,
                    failed=False,
                ),
            )
            card_instance.streaming(
                card_instance_id,
                content_key=content_key,
                content_value=full_content_value,
                append=False,
                finished=True,
                failed=False,
            )
        except Exception as e:
            self.logger.exception(e)
            card_instance.streaming(
                card_instance_id,
                content_key=content_key,
                content_value="",
                append=False,
                finished=False,
                failed=True,
            )

        return AckMessage.STATUS_OK, "OK"


def main():
    options = define_options()

    credential = dingtalk_stream.Credential(options.client_id, options.client_secret)
    client = dingtalk_stream.DingTalkStreamClient(credential)
    client.register_callback_handler(
        dingtalk_stream.ChatbotMessage.TOPIC, CardBotHandler()
    )
    client.start_forever()


if __name__ == "__main__":
    main()
