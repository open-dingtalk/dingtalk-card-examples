import os
import json
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
        card_template_id = "280f6d7a-63bc-4905-bf3f-4c6d95e5166b.schema"  # 该模板只用于测试使用，如需投入线上使用，请导入卡片模板 json 到自己的应用下
        # 卡片公有数据，非字符串类型的卡片数据参考文档：https://open.dingtalk.com/document/orgapp/instructions-for-filling-in-api-card-data
        card_data = {
            "form": {
                "fields": [
                    {
                        "name": "system_params_1",
                        "type": "TEXT",
                        "hidden": True,
                        "defaultValue": "asdf",
                    },
                    {
                        "name": "text",
                        "label": "必填文本输入",
                        "type": "TEXT",
                        "required": True,
                        "placeholder": "请输入文本",
                        "requiredMsg": "自定义必填错误提示",
                    },
                    {
                        "name": "text_optional",
                        "label": "非必填文本输入",
                        "type": "TEXT",
                        "placeholder": "请输入文本",
                    },
                    {
                        "name": "text_readonly",
                        "label": "非必填只读文本输入有默认值",
                        "type": "TEXT",
                        "readOnly": True,
                        "defaultValue": "文本默认值",
                    },
                    {
                        "name": "date",
                        "label": "必填日期选择",
                        "type": "DATE",
                        "required": True,
                        "placeholder": "请选择日期",
                    },
                    {
                        "name": "date_optional",
                        "label": "非必填日期选择",
                        "type": "DATE",
                        "placeholder": "请选择日期",
                    },
                    {
                        "name": "date_readonly",
                        "label": "非必填只读日期选择有默认值",
                        "type": "DATE",
                        "readOnly": True,
                        "defaultValue": "2024-05-27",
                    },
                    {
                        "name": "datetime",
                        "label": "必填日期时间选择",
                        "type": "DATETIME",
                        "required": True,
                        "placeholder": "请选择日期时间",
                    },
                    {
                        "name": "datetime_optional",
                        "label": "非必填日期时间选择",
                        "type": "DATETIME",
                        "placeholder": "请选择日期时间",
                    },
                    {
                        "name": "datetime_readonly",
                        "label": "非必填只读日期时间选择有默认值",
                        "type": "DATETIME",
                        "readOnly": True,
                        "defaultValue": "2024-05-27 12:00",
                    },
                    {
                        "name": "select",
                        "label": "必填单选",
                        "type": "SELECT",
                        "required": True,
                        "placeholder": "单选请选择",
                        "options": [
                            {"value": "1", "text": "选项1"},
                            {"value": "2", "text": "选项2"},
                            {"value": "3", "text": "选项3"},
                            {"value": "4", "text": "选项4"},
                        ],
                    },
                    {
                        "name": "select_optional",
                        "label": "非必填单选",
                        "type": "SELECT",
                        "placeholder": "单选请选择",
                        "options": [
                            {"value": "1", "text": "选项1"},
                            {"value": "2", "text": "选项2"},
                            {"value": "3", "text": "选项3"},
                            {"value": "4", "text": "选项4"},
                        ],
                    },
                    {
                        "name": "select_readonly",
                        "label": "非必填只读单选有默认值",
                        "type": "SELECT",
                        "readOnly": True,
                        "defaultValue": {"index": 3, "value": "4"},
                        "options": [
                            {"value": "1", "text": "选项1"},
                            {"value": "2", "text": "选项2"},
                            {"value": "3", "text": "选项3"},
                            {"value": "4", "text": "选项4"},
                        ],
                    },
                    {
                        "name": "multi_select",
                        "label": "必填多选",
                        "type": "MULTI_SELECT",
                        "required": True,
                        "placeholder": "多选请选择",
                        "options": [
                            {"value": "1", "text": "选项1"},
                            {"value": "2", "text": "选项2"},
                            {"value": "3", "text": "选项3"},
                            {"value": "4", "text": "选项4"},
                        ],
                    },
                    {
                        "name": "multi_select_optional",
                        "label": "非必填多选",
                        "type": "MULTI_SELECT",
                        "placeholder": "多选请选择",
                        "options": [
                            {"value": "1", "text": "选项1"},
                            {"value": "2", "text": "选项2"},
                            {"value": "3", "text": "选项3"},
                            {"value": "4", "text": "选项4"},
                        ],
                    },
                    {
                        "name": "multi_select_readonly",
                        "label": "非必填只读多选有默认值",
                        "type": "MULTI_SELECT",
                        "readOnly": True,
                        "defaultValue": {"index": [1, 3], "value": ["2", "4"]},
                        "options": [
                            {"value": "1", "text": "选项1"},
                            {"value": "2", "text": "选项2"},
                            {"value": "3", "text": "选项3"},
                            {"value": "4", "text": "选项4"},
                        ],
                    },
                    {"name": "checkbox", "label": "独立的复选框", "type": "CHECKBOX"},
                    {
                        "name": "checkbox_readonly",
                        "label": "只读独立的复选框",
                        "type": "CHECKBOX",
                        "readOnly": True,
                        "defaultValue": True,
                    },
                ]
            },
            "form_status": "normal",
            "form_btn_text": "提交",
            "title": content,
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

        user_private_data = {}

        card_private_data = incoming_message.content.get("cardPrivateData", {})
        params = card_private_data.get("params", {})

        form = params.get("form")
        current_form = params.get("current_form")
        if form and current_form:
            self.logger.info(f"form: {form}")
            for field in current_form.get("fields", []):
                submit_value = form.get(field["name"])
                if submit_value is not None:
                    field["defaultValue"] = submit_value
            user_private_data["form"] = current_form
            user_private_data["form_btn_text"] = "已提交"
            user_private_data["form_status"] = "disabled"

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
