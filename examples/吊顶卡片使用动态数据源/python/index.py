import os
import json
import time
import random
import logging
import argparse
from loguru import logger
from datetime import datetime
from dingtalk_stream import AckMessage
import dingtalk_stream


demo_dynamic_data_source_id = "demo_dynamic_data_source_id"

total_by_card_instance_id = {}
finished_by_card_instance_id = {}


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
        card_template_id = "c36a2fbe-ff53-44ac-a91d-dedbe3654306.schema"  # 该模板只用于测试使用，如需投入线上使用，请导入卡片模板 json 到自己的应用下
        # 卡片公有数据，非字符串类型的卡片数据参考文档：https://open.dingtalk.com/document/orgapp/instructions-for-filling-in-api-card-data
        title = f"{random.randint(1, 12)}月迭代"
        total = random.randint(100, 200)
        finished = 0
        card_data = {
            "title": title,
            "total": total,
            "finished": finished,
            "unfinished": total - finished,
            "progress": 0,
            "update_at": "",
        }

        pullConfig = {
            # 仅拉取一次数据
            "pullStrategy": "ONCE",
        }

        if content.upper() == "RENDER":
            # 每次卡片渲染时拉取数据
            pullConfig = {"pullStrategy": "RENDER"}

        if content.upper() == "INTERVAL":
            pullConfig = {
                # 每隔 10 秒拉取一次数据
                "pullStrategy": "INTERVAL",
                "interval": 10,
                "timeUnit": "SECONDS",
            }

        openDynamicDataConfig = {
            "dynamicDataSourceConfigs": [
                {
                    "dynamicDataSourceId": demo_dynamic_data_source_id,
                    "pullConfig": pullConfig,
                }
            ]
        }

        card_instance = dingtalk_stream.CardReplier(
            self.dingtalk_client, incoming_message
        )
        # 创建并投放卡片: https://open.dingtalk.com/document/orgapp/create-and-deliver-cards
        card_instance_id = card_instance.create_and_deliver_card(
            card_template_id,
            convert_json_values_to_string(card_data),
            openSpaceId=f"dtv1.card//ONE_BOX.{incoming_message.conversation_id}",
            topOpenSpaceModel={"spaceType": "ONE_BOX"},
            topOpenDeliverModel={"expiredTimeMillis": int(time.time() + 60 * 3) * 1000},  # 3 分钟后过期
            openDynamicDataConfig=openDynamicDataConfig,
        )

        total_by_card_instance_id[card_instance_id] = total
        finished_by_card_instance_id[card_instance_id] = finished

        self.logger.info(
            f"reply card: {card_instance_id} {card_data} {openDynamicDataConfig}"
        )

        return AckMessage.STATUS_OK, "OK"


# 卡片动态数据源回调文档：https://open.dingtalk.com/document/isvapp/dynamic-data-source
class CardDynamicDataHandler(dingtalk_stream.CallbackHandler):
    def __init__(self, logger: logging.Logger = logger):
        super(dingtalk_stream.CallbackHandler, self).__init__()
        if logger:
            self.logger = logger

    async def process(self, callback: dingtalk_stream.CallbackMessage):
        incoming_message = dingtalk_stream.CardCallbackMessage.from_dict(callback.data)
        self.logger.info(
            f"card dynamic data callback message: {incoming_message.to_dict()}"
        )

        card_instance_id = incoming_message.card_instance_id

        if card_instance_id not in total_by_card_instance_id:
            return AckMessage.STATUS_OK, "OK"

        total = total_by_card_instance_id[card_instance_id]
        finished_by_card_instance_id[card_instance_id] += 1
        finished = finished_by_card_instance_id[card_instance_id]
        if finished > total:
            return AckMessage.STATUS_OK, "OK"

        card_data = {
            "finished": finished,
            "unfinished": total - finished,
            "progress": round(finished / total * 100),
            "update_at": datetime.now().strftime("%m-%d %H:%M:%S"),
        }

        response = {
            "dataSourceQueryResponses": [
                {
                    # 动态数据源的 data 会同时被更新到公有数据、私有数据上
                    "data": json.dumps(card_data),
                    "dynamicDataSourceId": demo_dynamic_data_source_id,
                    "dynamicDataValueType": "OBJECT",
                }
            ]
        }

        self.logger.info(f"card dynamic data response: {response}")
        return AckMessage.STATUS_OK, response


def main():
    options = define_options()

    credential = dingtalk_stream.Credential(options.client_id, options.client_secret)
    client = dingtalk_stream.DingTalkStreamClient(credential)
    client.register_callback_handler(
        dingtalk_stream.ChatbotMessage.TOPIC, ChatBotHandler()
    )
    TOPIC_DYNAMIC_DATA = "/v1.0/card/dynamicData/get"
    client.register_callback_handler(TOPIC_DYNAMIC_DATA, CardDynamicDataHandler())
    client.start_forever()


if __name__ == "__main__":
    main()
