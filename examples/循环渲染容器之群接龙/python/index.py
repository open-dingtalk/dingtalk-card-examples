import os
import json
import logging
import argparse
import requests
from loguru import logger
from datetime import datetime
from dingtalk_stream import AckMessage
import dingtalk_stream


content_by_id = {}  # 生产环境最好将当前内容持久化到数据库中


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


def convert_json_values_to_string(obj: dict) -> str:
    result = {}
    for key, value in obj.items():
        if isinstance(value, str):
            result[key] = value
        else:
            result[key] = json.dumps(value)
    return result


class CardBotHandler(dingtalk_stream.ChatbotHandler):
    def __init__(self, logger: logging.Logger = logger):
        super(dingtalk_stream.ChatbotHandler, self).__init__()
        if logger:
            self.logger = logger

    async def process(self, callback: dingtalk_stream.CallbackMessage):
        incoming_message = dingtalk_stream.ChatbotMessage.from_dict(callback.data)
        self.logger.info(f"收到消息：{incoming_message}")

        # 卡片模板 ID
        card_template_id = "3d667b86-d30b-43ef-be8c-7fca37965210.schema"  # 该模板只用于测试使用，如需投入线上使用，请导入卡片模板 json 到自己的应用下
        # 卡片公有数据，非字符串类型的卡片数据参考文档：https://open.dingtalk.com/document/orgapp/instructions-for-filling-in-api-card-data        
        card_data = {"title": incoming_message.text.content, "joined": False}

        card_instance = dingtalk_stream.CardReplier(
            self.dingtalk_client, incoming_message
        )
        # 创建并投放卡片
        card_instance_id = card_instance.create_and_deliver_card(
            card_template_id,
            card_data=convert_json_values_to_string(card_data),
            callback_type="STREAM",
            support_forward=True,
        )
        self.logger.info(f"reply card {card_instance_id} {card_data}")

        return AckMessage.STATUS_OK, "OK"


class CardCallbackHandler(dingtalk_stream.CallbackHandler):
    def __init__(self, logger: logging.Logger = logger):
        super(dingtalk_stream.CallbackHandler, self).__init__()
        if logger:
            self.logger = logger

    def get_userinfo_by_userid(self, user_id: str):
        """
        查询用户详情
        https://open.dingtalk.com/document/isvapp/query-user-details
        :param user_id: 用户的userId。
        """
        access_token = self.dingtalk_client.get_access_token()
        if not access_token:
            self.logger.error(
                "CardCallbackHandler.get_userinfo_by_userid failed, connot get userinfo"
            )
            return
        body = {"userid": user_id}
        url = "https://oapi.dingtalk.com/topapi/v2/user/get"
        try:
            response = requests.post(
                url, params={"access_token": access_token}, json=body
            )
            response.raise_for_status()
            result = response.json()
            if result.get("errcode") == 0:
                return result.get("result")
            self.logger.error(
                f"errcode: {result.get('errcode')}, errmsg: {result.get('errmsg')}"
            )
        except Exception as e:
            self.logger.error(
                "CardCallbackHandler.get_userinfo_by_userid failed, error=%s", e
            )
            return

    async def process(self, callback: dingtalk_stream.CallbackMessage):
        """
        卡片事件回调文档：https://open.dingtalk.com/document/orgapp/event-callback-card
        """
        incoming_message = dingtalk_stream.CardCallbackMessage.from_dict(callback.data)
        user_id = incoming_message.user_id
        self.logger.info(f"card callback message: {incoming_message.to_dict()}")

        card_instance_id = incoming_message.card_instance_id

        card_private_data = incoming_message.content.get("cardPrivateData", {})
        params = card_private_data.get("params", {})
        current_content = content_by_id.get(card_instance_id) or []
        delete_uid = params.get("delete_uid")
        user_private_data = {"uid": user_id}
        next_content = []
        if delete_uid:
            # 取消接龙
            user_private_data["joined"] = False
            next_content = [x for x in current_content if x.get("uid") != delete_uid]
        else:
            # 参与接龙
            user_private_data["joined"] = True
            body = {
                "timestamp": datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
                "uid": user_id,
                "remark": params.get("remark"),
                "nick": "",
                "avatar": "",
            }
            user_info = self.get_userinfo_by_userid(user_id)
            logger.info(f"get user_info by user_id {user_id}: {user_info}")
            if user_info:
                body["nick"] = user_info.get("name", "")
                body["avatar"] = user_info.get("avatar", "")
            current_content.append(body)
            next_content = current_content
        content_by_id[card_instance_id] = next_content

        cardUpdateOptions = {
            "updateCardDataByKey": True,
            "updatePrivateDataByKey": True,
        }

        # 更新接龙列表和参与状态
        update_card_data = {"content": next_content}
        card_instance = dingtalk_stream.CardReplier(
            self.dingtalk_client, incoming_message
        )
        card_instance.put_card_data(
            card_instance_id,
            card_data=convert_json_values_to_string(update_card_data),
            privateData={
                user_id: {
                    "cardParamMap": convert_json_values_to_string(user_private_data)
                }
            },
            cardUpdateOptions=cardUpdateOptions,
        )
        self.logger.info(
            f"update card: {card_instance_id} {update_card_data} {user_private_data}"
        )
        return AckMessage.STATUS_OK, {}


def main():
    options = define_options()

    credential = dingtalk_stream.Credential(options.client_id, options.client_secret)
    client = dingtalk_stream.DingTalkStreamClient(credential)
    client.register_callback_handler(
        dingtalk_stream.ChatbotMessage.TOPIC, CardBotHandler()
    )
    client.register_callback_handler(
        dingtalk_stream.CallbackHandler.TOPIC_CARD_CALLBACK, CardCallbackHandler()
    )
    client.start_forever()


if __name__ == "__main__":
    main()
