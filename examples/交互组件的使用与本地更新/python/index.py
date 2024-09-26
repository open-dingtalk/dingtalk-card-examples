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
        card_template_id = "737cda86-7a7f-4d83-ba07-321e6933be12.schema"  # 该模板只用于测试使用，如需投入线上使用，请导入卡片模板 json 到自己的应用下
        # 卡片公有数据，非字符串类型的卡片数据参考文档：https://open.dingtalk.com/document/orgapp/instructions-for-filling-in-api-card-data
        card_data = {
            "lastMessage": "交互组件本地更新卡片",
            "submitBtnStatus": "normal",
            "submitBtnText": "提交",
            "input": "",
            "selectIndex": -1,
            "multiSelectIndexes": [],
            "date": "",
            "datetime": "",
            "checkbox": False,
            "singleCheckboxItems": [
                {"value": 0, "text": "单选复选框选项 1"},
                {"value": 1, "text": "单选复选框选项 2"},
                {"value": 2, "text": "单选复选框选项 3"},
                {"value": 3, "text": "单选复选框选项 4"},
            ],
            "multiCheckboxItems": [
                {"value": 0, "text": "多选复选框选项 1"},
                {"value": 1, "text": "多选复选框选项 2"},
                {"value": 2, "text": "多选复选框选项 3"},
                {"value": 3, "text": "多选复选框选项 4"},
            ],
        }

        card_instance = dingtalk_stream.CardReplier(
            self.dingtalk_client, incoming_message
        )
        # 创建并投放卡片: https://open.dingtalk.com/document/orgapp/create-and-deliver-cards
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
        required_fields = {
            "input": "文本输入",
            "select": "下拉单选",
            "multiSelect": "下拉多选",
            "date": "日期选择",
            "datetime": "日期时间选择",
            "singleCheckbox": "单选列表",
            "multiCheckbox": "多选列表",
        }

        card_private_data = incoming_message.content.get("cardPrivateData", {})
        params = card_private_data.get("params", {})

        input_ = params.get("input")
        if isinstance(input_, str) and input_:
            user_private_data["input"] = input_
            required_fields.pop("input")

        date = params.get("date")
        if isinstance(date, str) and date:
            user_private_data["date"] = date
            if date:
                required_fields.pop("date")

        datetime = params.get("datetime")
        if isinstance(datetime, str) and datetime:
            user_private_data["datetime"] = datetime
            required_fields.pop("datetime")

        select = params.get("select")
        if isinstance(select, dict) and "index" in select:
            user_private_data["selectIndex"] = select.get("index")
            required_fields.pop("select")

        multi_select = params.get("multiSelect")
        if isinstance(multi_select, dict) and "index" in multi_select and multi_select.get("index"):
            user_private_data["multiSelectIndexes"] = multi_select.get("index")
            required_fields.pop("multiSelect")

        checkbox = params.get("checkbox")
        if isinstance(checkbox, bool):
            user_private_data["checkbox"] = checkbox

        single_checkbox = params.get("singleCheckbox")
        single_checkbox_items = params.get("singleCheckboxItems")
        if isinstance(single_checkbox, int) and isinstance(single_checkbox_items, list):
            user_private_data["singleCheckboxItems"] = [
                {**item, "checked": single_checkbox == item.get("value")}
                for item in single_checkbox_items
            ]
            required_fields.pop("singleCheckbox")

        multi_checkbox = params.get("multiCheckbox")
        multi_checkbox_items = params.get("multiCheckboxItems")
        if isinstance(multi_checkbox, list) and multi_checkbox and isinstance(multi_checkbox_items, list):
            user_private_data["multiCheckboxItems"] = [
                {**item, "checked": item.get("value") in multi_checkbox}
                for item in multi_checkbox_items
            ]
            required_fields.pop("multiCheckbox")

        # 这里可以做必填校验，决定是否完成提交
        if required_fields:
            error_message = (
                f"表单未填写完整，{'、'.join(required_fields.values())} 是必填项"
            )
            self.logger.error(error_message)
            user_private_data = {"errMsg": error_message}
        else:
            user_private_data["submitBtnText"] = "已提交"
            user_private_data["submitBtnStatus"] = "disabled"

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
