from loguru import logger
from pydantic import BaseModel
from dingtalk_stream import DingTalkStreamClient, ChatbotMessage, CardReplier
from fastapi import APIRouter, Request, HTTPException, Body
from app.stream.utils import convert_json_values_to_string
from app.core.config import get_app_settings

router = APIRouter()
app_settings = get_app_settings()


class NoticeResponse(BaseModel):
    success: bool
    err_msg: str = ""


class ButtonType(BaseModel):
    text: str
    url: str


class ChannelNewLaunchedBody(BaseModel):
    new_launched_id: str = "111"
    new_launched_name: str = "钉钉卡片平台"
    new_launched_sub_id: str = "111333"
    new_launceed_sub_name: str = "如何搭建 AI 智能回复卡片"
    channel_id: str = "222"
    channel_name: str = "钉钉低代码"


class LiveBeginningBody(BaseModel):
    live_id: str = "111"
    live_name: str = "通义千问图生视频模型更新！"
    live_beginning_time: str = "今天下午16:00"


card_template_id = "25df7976-dc87-46de-aba3-2795fca4a399.schema"  # 导出的模板在 app/card_templates 目录下的 「多 Tab 页.json」
common_card_data = {
    "formInfo": {
        "title": "审批",
        "tag": "请假",
        "nianjia_days1": 5,
        "tiaoxiu_days2": 2,
        "submitBtnText": "提交",
        "submitBtnStatus": "normal",
    },
    "metricsTitle": "昨日（20250625）钉瓜瓜运营数据推送",
    "metrics": [
        {"text": "卡片发送量", "count": 10086, "unit": "次"},
        {"text": "峰值发送量", "count": 300, "unit": "次/秒"},
        {"text": "直播次数", "count": 2, "unit": "次"},
        {"text": "直播峰值人数", "count": 1024, "unit": "个"},
    ],
}
fake_incoming_message_dict = {
    "conversationId": "",  # 群聊需要
    "senderNick": "小钉",  # 群聊 at 人时需要
    "senderStaffId": app_settings.user_id,  # 单聊需要
    "senderCorpId": "",
    "conversationType": "1",  # 2: 群聊, 1: 单聊
    "senderId": "",
}


def gen_two_btns(btn1: ButtonType, btn2: ButtonType):
    return [
        {
            "text": btn1.text,
            "color": "blue",
            "action": {
                "type": "openLink",
                "params": {"url": btn1.url},
            },
        },
        {
            "text": btn2.text,
            "color": "gray",
            "action": {
                "type": "openLink",
                "params": {"url": btn2.url},
            },
        },
    ]


@router.post(
    "/channel_new_course",
    response_model=NoticeResponse,
    name="notice:channel-new-course",
)
async def channel_new_wiki(
    request: Request, body: ChannelNewLaunchedBody = Body(...)
) -> NoticeResponse:
    """
    课程更新卡片通知
    """
    app = request.app
    client: DingTalkStreamClient = getattr(app.state, "dingtalk_client", None)
    if not client:
        raise HTTPException(status_code=500, detail="dingtalk client not initialized")

    # 构造一个 incoming_message 对象
    incoming_message = ChatbotMessage.from_dict(
        {**fake_incoming_message_dict, "senderStaffId": app_settings.user_id}
    )
    card_data = {
        **common_card_data,
        "config": {"autoLayout": True},
        "title": "📚课程更新提醒！",
        "content": f"Hi，{incoming_message.sender_nick}同学：<br>你报名的课程[「{body.new_launched_name}」](https://www.dingtalk.com?id={body.new_launched_id})下，新增了一节内容[「{body.new_launceed_sub_name}」](https://www.dingtalk.com?id={body.new_launched_sub_id})。<br>快来继续学习吧！",
        "btns": gen_two_btns(
            ButtonType(
                text="👉继续学习👈",
                url=f"https://www.dingtalk.com?new_launched_sub_id={body.new_launched_sub_id}",
            ),
            ButtonType(
                text="解锁更多精品课程",
                url=f"https://www.dingtalk.com?channel_id={body.channel_id}",
            ),
        ),
    }
    card_instance = CardReplier(client, incoming_message)

    card_instance_id = await card_instance.async_create_and_deliver_card(
        card_template_id,
        convert_json_values_to_string(card_data),
        imRobotOpenSpaceModel={
            "supportForward": False,
            "lastMessageI18n": {"ZH_CN": "课程更新提醒"},
        },
    )

    logger.info(f"reply card {card_template_id} {card_instance_id} {card_data}")

    return NoticeResponse(success=True)


@router.post(
    "/live_beginning",
    response_model=NoticeResponse,
    name="notice:live-beginning",
)
async def channel_new_wiki(
    request: Request, body: LiveBeginningBody = Body(...)
) -> NoticeResponse:
    """
    预约的直播开始提醒
    """
    app = request.app
    client: DingTalkStreamClient = getattr(app.state, "dingtalk_client", None)
    if not client:
        raise HTTPException(status_code=500, detail="dingtalk client not initialized")

    # 构造一个 incoming_message 对象
    incoming_message = ChatbotMessage.from_dict(
        {**fake_incoming_message_dict, "senderStaffId": app_settings.user_id}
    )
    card_data = {
        **common_card_data,
        "config": {"autoLayout": True},
        "title": "⏰直播开始提醒：你预约的直播快开始啦！",
        "content": f"Hi，{incoming_message.sender_nick}同学：<br>你预约的直播「{body.live_name}」将在<font colorTokenV2=common_green1_color>{body.live_beginning_time}</font>正式开播，别忘记准时来听哦！<br>快来钉瓜瓜学习吧！",
        "btns": gen_two_btns(
            ButtonType(
                text="查看直播👈",
                url=f"https://www.dingtalk.com?live_id={body.live_id}",
            ),
            ButtonType(
                text="解锁更多精品课程",
                url=f"https://www.dingtalk.com",
            ),
        ),
    }
    card_instance = CardReplier(client, incoming_message)

    card_instance_id = await card_instance.async_create_and_deliver_card(
        card_template_id,
        convert_json_values_to_string(card_data),
        imRobotOpenSpaceModel={
            "supportForward": False,
            "lastMessageI18n": {"ZH_CN": "预约直播提醒"},
        },
    )

    logger.info(f"reply card {card_template_id} {card_instance_id} {card_data}")

    return NoticeResponse(success=True)
