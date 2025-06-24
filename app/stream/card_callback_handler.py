import json
from loguru import logger
from app.stream.utils import convert_json_values_to_string
from dingtalk_stream import (
    CallbackHandler,
    CallbackMessage,
    CardCallbackMessage,
    AckMessage,
)


class CardCallbackHandler(CallbackHandler):
    def __init__(self, logger=logger):
        super(CallbackHandler, self).__init__()
        if logger:
            self.logger = logger

    async def process(self, callback: CallbackMessage):
        incoming_message = CardCallbackMessage.from_dict(callback.data)
        card_private_data = incoming_message.content.get("cardPrivateData", {})
        params = card_private_data.get("params", {})
        self.logger.info(f"received callback params: {params}")

        action = params.get("action")

        card_data = {}
        user_private_data = {}

        if action == "submit_form":
            # 处理审批表单提交
            form_info = params.get("form_info", {})
            form_data = params.get("form_data", {})
            self.logger.success(
                f"审批表单提交：{json.dumps(form_data, ensure_ascii=False)}。审批相关信息：{json.dumps(form_info, ensure_ascii=False)}"
            )
            form_info["submitBtnText"] = "已提交"
            form_info["submitBtnStatus"] = "disabled"
            card_data["formInfo"] = form_info
            card_data["formData"] = {
                **form_data,
                "typeIndex": form_data.get("type", {}).get("index", -1),
            }

        if action == "submit_dislike":
            # 处理点踩提交
            self.logger.success(
                f"点踩原因：{'、'.join(params.get('dislike_reason', []))}。自定义点踩原因：{params.get('custom_dislike_reason', '')}"
            )
            card_data["submitted"] = True

        cardUpdateOptions = {
            "updateCardDataByKey": True,
            "updatePrivateDataByKey": True,
        }

        response = {
            "cardUpdateOptions": cardUpdateOptions,
            "cardData": {
                "cardParamMap": convert_json_values_to_string(card_data),
            },
            "userPrivateData": {
                "cardParamMap": convert_json_values_to_string(user_private_data)
            },
        }
        self.logger.info(f"response: {response}")
        return AckMessage.STATUS_OK, response
