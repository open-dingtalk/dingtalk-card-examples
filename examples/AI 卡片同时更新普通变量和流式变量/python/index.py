import os
import time
import json
import logging
import argparse
from loguru import logger
from random import random
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


def convert_json_values_to_string(obj: dict) -> str:
    result = {}
    for key, value in obj.items():
        if isinstance(value, str):
            result[key] = value
        else:
            result[key] = json.dumps(value, ensure_ascii=False)
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
        card_template_id = "2b07a3e6-cdf4-4e8f-8bac-a4f6e3e19eff.schema"  # 该模板只用于测试使用，如需投入线上使用，请导入卡片模板 json 到自己的应用下
        content_key = "content"
        card_data = {
            content_key: "",
            "query": incoming_message.text.content,
            "preparations": [],
            "charts": [],
            "config": {"autoLayout": True},
        }
        card_instance = dingtalk_stream.AICardReplier(
            self.dingtalk_client, incoming_message
        )
        # 先投放卡片
        card_instance_id = card_instance.create_and_deliver_card(
            card_template_id, convert_json_values_to_string(card_data)
        )
        # 流式更新卡片
        try:
            # 更新成输入中状态
            card_instance.streaming(
                card_instance_id,
                content_key=content_key,
                content_value="",
                append=False,
                finished=False,
                failed=False,
            )
            # 更新卡片
            cardUpdateOptions = {
                "updateCardDataByKey": True,
                "updatePrivateDataByKey": True,
            }
            update_card_data = {"preparations": []}
            for action in [
                "正在理解需求",
                "正在生成 SQL",
                "正在执行 SQL",
                "正在生成图表",
            ]:
                update_card_data["preparations"].append({"name": action, "progress": 0})
                for progress in range(101):
                    if progress % 20 == 0:
                        update_card_data["preparations"][-1]["progress"] = progress
                        card_instance.put_card_data(
                            card_instance_id,
                            card_data=convert_json_values_to_string(update_card_data),
                            cardUpdateOptions=cardUpdateOptions,
                        )
                        time.sleep(random())
            fake_content_values = [
                "## 过去一个月的营收分析",
                "\n本报告分析了过去一个月的营收情况。",
                "包括每周营收的趋势图、",
                "按不同产品分类的柱状图、",
                "以及不同产品的总营收占比饼图。",
            ]
            content_value = ""
            for fake_content_value in fake_content_values:
                content_value += fake_content_value
                card_instance.streaming(
                    card_instance_id,
                    content_key=content_key,
                    content_value=content_value,
                    append=False,
                    finished=False,
                    failed=False,
                )
                time.sleep(random())
            card_instance.streaming(
                card_instance_id,
                content_key=content_key,
                content_value=content_value,
                append=False,
                finished=True,
                failed=False,
            )
            line_md = "# 过去一个月每周营收"
            line_chart = {
                "data": [
                    {"x": "0603", "y": 820, "type": "产品 A"},
                    {"x": "0610", "y": 932, "type": "产品 A"},
                    {"x": "0617", "y": 901, "type": "产品 A"},
                    {"x": "0624", "y": 934, "type": "产品 A"},
                    {"x": "0701", "y": 1290, "type": "产品 A"},
                    {"x": "0603", "y": 1498, "type": "产品 B"},
                    {"x": "0610", "y": 809, "type": "产品 B"},
                    {"x": "0617", "y": 728, "type": "产品 B"},
                    {"x": "0624", "y": 759, "type": "产品 B"},
                    {"x": "0701", "y": 995, "type": "产品 B"},
                    {"x": "0603", "y": 1249, "type": "产品 C"},
                    {"x": "0610", "y": 873, "type": "产品 C"},
                    {"x": "0617", "y": 1086, "type": "产品 C"},
                    {"x": "0624", "y": 908, "type": "产品 C"},
                    {"x": "0701", "y": 972, "type": "产品 C"},
                ],
                "type": "lineChart",
                "config": {},
            }
            bar_md = "# 不同产品的营收"
            bar_chart = {
                "type": "histogram",
                "data": [
                    {"x": "W1", "y": 820, "type": "产品 A"},
                    {"x": "W2", "y": 932, "type": "产品 A"},
                    {"x": "W3", "y": 901, "type": "产品 A"},
                    {"x": "W4", "y": 934, "type": "产品 A"},
                    {"x": "W5", "y": 1290, "type": "产品 A"},
                    {"x": "W1", "y": 1498, "type": "产品 B"},
                    {"x": "W2", "y": 809, "type": "产品 B"},
                    {"x": "W3", "y": 728, "type": "产品 B"},
                    {"x": "W4", "y": 759, "type": "产品 B"},
                    {"x": "W5", "y": 995, "type": "产品 B"},
                    {"x": "W1", "y": 1249, "type": "产品 C"},
                    {"x": "W2", "y": 873, "type": "产品 C"},
                    {"x": "W3", "y": 1086, "type": "产品 C"},
                    {"x": "W4", "y": 908, "type": "产品 C"},
                    {"x": "W5", "y": 972, "type": "产品 C"},
                ],
                "config": {},
            }
            pie_md = "# 不同产品的营收占比\n\n| 产品名称 | 营收 |\n| :-: | :-: |\n| 产品 A | 4877 |\n| 产品 B | 4789 |\n| 产品 C | 5088 |"
            pie_chart = {
                "type": "pieChart",
                "data": [
                    {"x": "产品 A", "y": 4877},
                    {"x": "产品 B", "y": 4789},
                    {"x": "产品 C", "y": 5088},
                ],
                "config": {"padding": [20, 30, 20, 30]},
            }
            update_card_data = []
            update_card_data.append({"markdown": line_md, "chart": line_chart})
            card_instance.put_card_data(
                card_instance_id,
                card_data=convert_json_values_to_string({"charts": update_card_data}),
                cardUpdateOptions=cardUpdateOptions,
            )
            time.sleep(random() * 2)

            update_card_data.append({"markdown": bar_md, "chart": bar_chart})
            card_instance.put_card_data(
                card_instance_id,
                card_data=convert_json_values_to_string({"charts": update_card_data}),
                cardUpdateOptions=cardUpdateOptions,
            )
            time.sleep(random() * 2)

            update_card_data.append({"markdown": pie_md, "chart": pie_chart})
            card_instance.put_card_data(
                card_instance_id,
                card_data=convert_json_values_to_string({"charts": update_card_data}),
                cardUpdateOptions=cardUpdateOptions,
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
