import os
import json
import logging
import argparse
from loguru import logger
from random import randint
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


def convert_json_values_to_string(obj: dict) -> dict:
    result = {}
    for key, value in obj.items():
        if isinstance(value, str):
            result[key] = value
        else:
            result[key] = json.dumps(value, ensure_ascii=False)
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
        card_template_id = "fcc1df51-17bb-403f-aca9-65f1c6919129.schema"  # 该模板只用于测试使用，如需投入线上使用，请导入卡片模板 json 到自己的应用下
        # 卡片公有数据，非字符串类型的卡片数据参考文档：https://open.dingtalk.com/document/orgapp/instructions-for-filling-in-api-card-data
        card_data = {
            "last_message": "事件链演示",
            "markdown": "<font colorTokenV2=common_green1_color>动态显示的 markdown 内容</font>",
        }

        card_instance = dingtalk_stream.CardReplier(
            self.dingtalk_client, incoming_message
        )
        # 创建并投放卡片
        card_instance_id = card_instance.create_and_deliver_card(
            card_template_id,
            convert_json_values_to_string(card_data),
        )

        self.logger.info(f"reply card: {card_instance_id} {card_data}")

        return AckMessage.STATUS_OK, "OK"


class CardCallbackHandler(dingtalk_stream.CallbackHandler):
    def __init__(self, logger: logging.Logger = logger):
        super(dingtalk_stream.CallbackHandler, self).__init__()
        if logger:
            self.logger = logger

    async def process(self, callback: dingtalk_stream.CallbackMessage):
        """
        卡片事件回调文档：https://open.dingtalk.com/document/orgapp/event-callback-card
        """
        incoming_message = dingtalk_stream.CardCallbackMessage.from_dict(callback.data)
        self.logger.info(f"card callback message: {incoming_message.to_dict()}")

        update_card_data = {}
        user_private_data = {}

        card_private_data = incoming_message.content.get("cardPrivateData", {})
        params = card_private_data.get("params", {})
        variable = params.get("var")

        if variable == "pub_url":
            # 回传请求更新公有数据打开链接或弹窗提示
            update_card_data["pub_url"] = ""
            update_card_data["pub_url_msg"] = ""
            update_card_data["pub_url_status"] = (
                "success" if randint(0, 1) else "failed"
            )
            if update_card_data["pub_url_status"] == "success":
                update_card_data["pub_url"] = (
                    "dingtalk://dingtalkclient/page/link?web_wnd=workbench&pc_slide=true&hide_bar=true&url=https://www.dingtalk.com"
                )
            else:
                update_card_data["pub_url_msg"] = f"更新失败{randint(0, 100)}"
        elif variable == "pri_url":
            # 回传请求更新私有数据打开链接或弹窗提示
            user_private_data["pri_url"] = ""
            user_private_data["pri_url_msg"] = ""
            user_private_data["pri_url_status"] = (
                "success" if randint(0, 1) else "failed"
            )
            if user_private_data["pri_url_status"] == "success":
                user_private_data["pri_url"] = (
                    "dingtalk://dingtalkclient/page/link?web_wnd=workbench&pc_slide=true&hide_bar=true&url=https://github.com/open-dingtalk/dingtalk-card-examples"
                )
            else:
                user_private_data["pri_url_msg"] = f"更新失败{randint(0, 100)}"

        response = {
            "cardUpdateOptions": {
                "updateCardDataByKey": True,
                "updatePrivateDataByKey": True,
            },
            "cardData": {
                "cardParamMap": convert_json_values_to_string(update_card_data)
            },
            "userPrivateData": {
                "cardParamMap": convert_json_values_to_string(user_private_data),
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
