import os
import json
import time
import logging
import argparse
from loguru import logger
from dingtalk_stream import AckMessage
import dingtalk_stream


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


def convertJSONValuesToString(obj: dict) -> str:
    result = {}
    for key, value in obj.items():
        if isinstance(value, str):
            result[key] = value
        else:
            result[key] = json.dumps(value)
    return result


class ChatBotHandler(dingtalk_stream.ChatbotHandler):
    def __init__(self, logger: logging.Logger = logger):
        super(dingtalk_stream.ChatbotHandler, self).__init__()
        if logger:
            self.logger = logger

    async def process(self, callback: dingtalk_stream.CallbackMessage):
        incoming_message = dingtalk_stream.ChatbotMessage.from_dict(callback.data)
        content = (incoming_message.text.content or "").strip()
        self.logger.info(f"received message: {content}")

        # 卡片模板 ID
        card_template_id = "2c278d79-fc0b-41b4-b14e-8b8089dc08e8.schema"  # 该模板只用于测试使用，如需投入线上使用，请导入卡片模板 json 到自己的应用下
        # 卡片公有数据，非字符串类型的卡片数据参考文档：https://open.dingtalk.com/document/orgapp/instructions-for-filling-in-api-card-data
        card_data = {
            "markdown": content,
            "submitted": False,
            "title": "钉钉互动卡片",
            "tag": "标签",
        }

        card_instance = dingtalk_stream.CardReplier(
            self.dingtalk_client, incoming_message
        )
        # 创建并投放卡片
        card_instance_id = card_instance.create_and_deliver_card(
            card_template_id,
            convertJSONValuesToString(card_data),
        )

        self.logger.info(f"reply card: {card_instance_id} {card_data}")

        # 更新卡片
        time.sleep(2)
        update_card_data = {"tag": "更新后的标签"}
        card_instance.put_card_data(
            card_instance_id,
            convertJSONValuesToString(update_card_data),
            cardUpdateOptions={"updateCardDataByKey": True},
        )
        self.logger.info(f"update card: {card_instance_id} {update_card_data}")

        return AckMessage.STATUS_OK, "OK"


class CardCallbackHandler(dingtalk_stream.CallbackHandler):
    def __init__(self, logger: logging.Logger = logger):
        super(dingtalk_stream.CallbackHandler, self).__init__()
        if logger:
            self.logger = logger

    async def process(self, callback: dingtalk_stream.CallbackMessage):
        incoming_message = dingtalk_stream.CardCallbackMessage.from_dict(callback.data)
        self.logger.info(f"card callback message: {incoming_message.to_dict()}")

        user_private_data = {}

        card_private_data = incoming_message.content.get("cardPrivateData", {})
        params = card_private_data.get("params", {})
        local_input = params.get("local_input")

        if local_input is not None:
            user_private_data["private_input"] = local_input
            user_private_data["submitted"] = True

        cardUpdateOptions = {
            "updateCardDataByKey": True,
            "updatePrivateDataByKey": True,
        }

        response = {
            "cardUpdateOptions": cardUpdateOptions,
            "userPrivateData": {
                "cardParamMap": convertJSONValuesToString(user_private_data),
            },
        }

        self.logger.info(f"card callback response: {response}")
        return AckMessage.STATUS_OK, response


def main():
    options = define_options()

    credential = dingtalk_stream.Credential(options.client_id, options.client_secret)
    client = dingtalk_stream.DingTalkStreamClient(credential)
    client.register_callback_handler(
        dingtalk_stream.ChatbotMessage.TOPIC, ChatBotHandler()
    )
    client.register_callback_handler(
        dingtalk_stream.CallbackHandler.TOPIC_CARD_CALLBACK, CardCallbackHandler()
    )
    client.start_forever()


if __name__ == "__main__":
    main()
