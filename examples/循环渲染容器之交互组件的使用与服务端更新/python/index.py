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


value_key_map = {
    "TEXT": "default_string",
    "DATE": "default_string",
    "DATETIME": "default_string",
    "SELECT": "default_number",
    "MULTI_SELECT": "default_number_array",
    "CHECKBOX": "default_boolean",
    "CHECKBOX_LIST": "checkbox_items",
    "CHECKBOX_LIST_MULTI": "checkbox_items",
}
form_fields_by_instance_id = {}


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
        card_template_id = "9f86e003-e65e-4680-bf4b-8df5958d9f17.schema"  # 该模板只用于测试使用，如需投入线上使用，请导入卡片模板 json 到自己的应用下
        # 卡片公有数据，非字符串类型的卡片数据参考文档：https://open.dingtalk.com/document/orgapp/instructions-for-filling-in-api-card-data
        form_fields = [
            {
                "type": "TEXT",
                "required": True,
                "name": "text_required",
                "label": "必填文本输入",
                "placeholder": "请输入文本",
                "default_string": "",
            },
            {
                "type": "TEXT",
                "name": "text",
                "label": "文本输入",
                "placeholder": "请输入文本",
                "default_string": "",
            },
            {
                "type": "DATE",
                "required": True,
                "name": "date_required",
                "label": "必填日期选择",
                "placeholder": "请选择日期",
            },
            {
                "type": "DATE",
                "name": "date",
                "label": "日期选择",
                "placeholder": "请选择日期",
                "default_string": "2024-06-06",
            },
            {
                "type": "DATETIME",
                "required": True,
                "name": "datetime_required",
                "label": "必填日期时间选择",
                "placeholder": "请选择日期时间",
            },
            {
                "type": "DATETIME",
                "name": "datetime",
                "label": "日期时间选择",
                "placeholder": "请选择日期时间",
                "default_string": "2024-06-06 12:00",
            },
            {
                "type": "SELECT",
                "required": True,
                "name": "select_required",
                "label": "必填单选下拉框",
                "placeholder": "单选请选择",
                "options": [
                    {"value": 1, "text": {"zh_CN": "选项 1"}},
                    {"value": 2, "text": {"zh_CN": "选项 2"}},
                    {"value": 3, "text": {"zh_CN": "选项 3"}},
                    {"value": 4, "text": {"zh_CN": "选项 4"}},
                ],
            },
            {
                "type": "SELECT",
                "name": "select",
                "label": "单选下拉框",
                "placeholder": "单选请选择",
                "default_number": 1,
                "options": [
                    {"value": 1, "text": {"zh_CN": "选项 1"}},
                    {"value": 2, "text": {"zh_CN": "选项 2"}},
                    {"value": 3, "text": {"zh_CN": "选项 3"}},
                    {"value": 4, "text": {"zh_CN": "选项 4"}},
                ],
            },
            {
                "type": "MULTI_SELECT",
                "required": True,
                "name": "multi_select",
                "label": "必填多选下拉框",
                "placeholder": "多选请选择",
                "default_number_array": [0, 2],
                "options": [
                    {"value": 1, "text": {"zh_CN": "选项 1"}},
                    {"value": 2, "text": {"zh_CN": "选项 2"}},
                    {"value": 3, "text": {"zh_CN": "选项 3"}},
                    {"value": 4, "text": {"zh_CN": "选项 4"}},
                ],
            },
            {
                "type": "CHECKBOX_LIST",
                "required": True,
                "name": "checkbox_list",
                "label": "必填单选列表",
                "checkbox_items": [
                    {
                        "value": 0,
                        "text": "选项 0",
                        "checked": False,
                        "name": "checkbox_list",
                        "type": "CHECKBOX_LIST",
                    },
                    {
                        "value": 1,
                        "text": "选项 1",
                        "checked": False,
                        "name": "checkbox_list",
                        "type": "CHECKBOX_LIST",
                    },
                    {
                        "value": 2,
                        "text": "选项 2",
                        "checked": False,
                        "name": "checkbox_list",
                        "type": "CHECKBOX_LIST",
                    },
                    {
                        "value": 3,
                        "text": "选项 3",
                        "checked": False,
                        "name": "checkbox_list",
                        "type": "CHECKBOX_LIST",
                    },
                ],
            },
            {
                "type": "CHECKBOX_LIST_MULTI",
                "required": True,
                "name": "checkbox_list_multi",
                "label": "必填多选列表",
                "checkbox_items": [
                    {
                        "value": 0,
                        "text": "选项 0",
                        "checked": False,
                        "name": "checkbox_list_multi",
                        "type": "CHECKBOX_LIST_MULTI",
                    },
                    {
                        "value": 1,
                        "text": "选项 1",
                        "checked": True,
                        "name": "checkbox_list_multi",
                        "type": "CHECKBOX_LIST_MULTI",
                    },
                    {
                        "value": 2,
                        "text": "选项 2",
                        "checked": False,
                        "name": "checkbox_list_multi",
                        "type": "CHECKBOX_LIST_MULTI",
                    },
                    {
                        "value": 3,
                        "text": "选项 3",
                        "checked": True,
                        "name": "checkbox_list_multi",
                        "type": "CHECKBOX_LIST_MULTI",
                    },
                ],
            },
            {
                "type": "CHECKBOX",
                "name": "checkbox",
                "label": "复选框",
            },
            {
                "type": "CHECKBOX",
                "name": "checkbox_default_true",
                "label": "复选框默认勾选",
                "default_boolean": True,
            },
        ]
        card_data = {
            "form_fields": form_fields,
            "form_status": "normal",
            "button_text": "提交",
            "title": content,
            "err_msg": "",
        }

        card_instance = dingtalk_stream.CardReplier(
            self.dingtalk_client, incoming_message
        )
        # 创建并投放卡片: https://open.dingtalk.com/document/orgapp/create-and-deliver-cards
        card_instance_id = card_instance.create_and_deliver_card(
            card_template_id,
            convert_json_values_to_string(card_data),
        )
        form_fields_by_instance_id[card_instance_id] = form_fields

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
        action_id = card_private_data.get("actionIds", [""])[0]
        params = card_private_data.get("params", {})

        submit_form_fields = params.get("submit_form_fields", [])

        if submit_form_fields:
            # 提交表单，做必填校验，响应错误提示或者响应提交成功处理
            required_error_labels = []
            for form_field in submit_form_fields:
                form_field_type = form_field.get("type")
                if form_field_type not in value_key_map:
                    user_private_data["err_msg"] = (
                        f"无效的表单类型「{form_field_type}」"
                    )
                    break
                form_field_value = form_field.get(value_key_map[form_field_type])
                form_field_label = form_field.get("label")
                form_field_required = form_field.get("required", False)
                if form_field_type in ["CHECKBOX_LIST", "CHECKBOX_LIST_MULTI"]:
                    if form_field_required and (
                        len(
                            list(
                                filter(
                                    lambda x: x.get("checked"),
                                    form_field_value or [],
                                ),
                            )
                        )
                        == 0
                    ):
                        required_error_labels.append(form_field_label)
                elif form_field_type == "SELECT":
                    if form_field_required and not (
                        isinstance(form_field_value, int) and form_field_value >= 0
                    ):
                        required_error_labels.append(form_field_label)
                else:
                    if form_field_required and not form_field_value:
                        required_error_labels.append(form_field_label)

            if user_private_data.get("err_msg") is None:
                if required_error_labels:
                    user_private_data["err_msg"] = (
                        f"请填写必填项「{', '.join(required_error_labels)}」"
                    )
                else:
                    user_private_data["form_status"] = "disabled"
                    user_private_data["button_text"] = "已提交"
        else:
            # 更新表单项
            update_name = params.get("name")
            self.logger.info(f"update name={update_name}")
            if update_name:
                card_instance_id = incoming_message.card_instance_id
                form_fields = form_fields_by_instance_id.get(card_instance_id, [])
                for form_field in form_fields:
                    if form_field.get("name") == update_name:
                        update_type = params.get("type")
                        update_key = value_key_map.get(update_type)
                        update_value = ""
                        if params.get("remove") and update_type == "multiSelect":
                            update_type = "MULTI_SELECT"
                            update_name = action_id
                            update_key = value_key_map.get(update_type)
                            remove_index = params.get(action_id, {}).get("index")
                            update_value = [
                                v
                                for v in form_field.get(update_key, [])
                                if v != remove_index
                            ]
                        elif update_type == "CHECKBOX_LIST":
                            update_value = [
                                {
                                    **value,
                                    "checked": value.get("value")
                                    == params.get("value"),
                                }
                                for value in form_field.get(update_key, [])
                            ]
                        elif update_type == "CHECKBOX_LIST_MULTI":
                            # 对选中的 checkbox checked 取反
                            update_value = [
                                {
                                    **value,
                                    "checked": (
                                        not value.get("checked")
                                        if value.get("value") == params.get("value")
                                        else value.get("checked")
                                    ),
                                }
                                for value in form_field.get(update_key, [])
                            ]
                        else:
                            update_value = params.get(update_name)
                            if update_type in ["SELECT", "MULTI_SELECT"]:
                                update_value = update_value.get("index")
                            elif update_type == "CHECKBOX":
                                update_value = not form_field.get(update_key)
                        self.logger.info(
                            f"update name={update_name}, type={update_type}, key={update_key}, value={update_value}"
                        )
                        form_field[update_key] = update_value
                user_private_data["form_fields"] = form_fields
            else:
                user_private_data["err_msg"] = "服务异常"

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
