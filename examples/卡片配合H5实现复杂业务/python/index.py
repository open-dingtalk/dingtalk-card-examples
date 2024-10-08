import os
import json
import logging
import argparse
from loguru import logger
from urllib.parse import quote
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
        card_template_id = "b23d3b9d-1c9c-4a3b-82a8-744d475c483d.schema"  # 该模板只用于测试使用，如需投入线上使用，请导入卡片模板 json 到自己的应用下
        # 卡片公有数据，非字符串类型的卡片数据参考文档: https://open.dingtalk.com/document/orgapp/instructions-for-filling-in-api-card-data
        card_data = {
            "more_detail_url": f"dingtalk://dingtalkclient/page/link?pc_slide=true&url={quote('http://localhost:3000?page=detail&id=123')}",
            "evaluate_url": f"dingtalk://dingtalkclient/page/link?pc_slide=true&url={quote('http://localhost:3000?page=evaluate&id=123')}",
            "table": {
                "data": [
                    {
                        "uv": "324",
                        "pv": "433",
                        "rank": 1,
                        "appItem": {
                        "icon": "https://static.dingtalk.com/media/lALPDeC2uGvNwy3NArzNArw_700_700.png",
                        "name": "考勤打卡"
                        }
                    },
                    {
                        "uv": "350",
                        "pv": "354",
                        "rank": 2,
                        "appItem": {
                        "icon": "https://static.dingtalk.com/media/lALPDeC2uGvNwy3NArzNArw_700_700.png",
                        "name": "智能人事"
                        }
                    },
                    {
                        "uv": "189",
                        "pv": "322",
                        "rank": 3,
                        "appItem": {
                        "icon": "https://static.dingtalk.com/media/lALPDeC2uGvNwy3NArzNArw_700_700.png",
                        "name": "日志"
                        }
                    }
                ],
                "meta": [
                    {
                        "aliasName": "",
                        "dataType": "STRING",
                        "alias": "rank",
                        "weight": 10
                    },
                    {
                        "aliasName": "应用名",
                        "dataType": "MICROAPP",
                        "alias": "appItem",
                        "weight": 40
                    },
                    {
                        "aliasName": "点击次数",
                        "dataType": "STRING",
                        "alias": "pv",
                        "weight": 25
                    },
                    {
                        "aliasName": "点击人数",
                        "dataType": "STRING",
                        "alias": "uv",
                        "weight": 25
                    }
                ]
            }
        }

        card_instance = dingtalk_stream.CardReplier(
            self.dingtalk_client, incoming_message
        )
        # 创建并投放卡片: https://open.dingtalk.com/document/isvapp/create-and-deliver-cards
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
