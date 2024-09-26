import {
  DWClient,
  DWClientDownStream,
  EventAck,
  TOPIC_ROBOT,
  TOPIC_CARD,
} from "dingtalk-stream";
import { program } from "commander";
import CardReplier from "./cardReplier";

program
  .requiredOption(
    "--clientId <Client ID>",
    "your client id, AppKey or SuiteKey",
    process.env.DINGTALK_APP_CLIENT_ID
  )
  .requiredOption(
    "--clientSecret <Client Secret>",
    "your client secret, AppSecret or SuiteSecret",
    process.env.DINGTALK_APP_CLIENT_SECRET
  )
  .parse();
const options = program.opts();

// 给日志加上时间戳前缀
const originalLog = console.log;
const originalError = console.error;
console.log = (...args: any) => {
  const timestamp = new Date().toISOString();
  originalLog(`[${timestamp}]`, ...args);
};
console.error = (...args: any) => {
  const timestamp = new Date().toISOString();
  originalError(`[${timestamp}]`, ...args);
};

const client = new DWClient({
  clientId: options.clientId,
  clientSecret: options.clientSecret,
  debug: false, // 调试模式，开启后可以看到更多详细日志
});

// 中断程序前先关闭 websocket 连接
const handleExit = () => {
  client.disconnect();
  process.exit(0);
};
process.on("SIGINT", handleExit);
process.on("SIGTERM", handleExit);

const convertJSONValuesToString = (obj: Record<string, any>) => {
  const newObj: Record<string, string> = {};
  for (const key in obj) {
    const value = obj[key];
    if (obj.hasOwnProperty(key) && value != null) {
      if (typeof value === "string") {
        newObj[key] = value;
      } else {
        newObj[key] = JSON.stringify(value);
      }
    }
  }
  return newObj;
};

// 机器人接收消息回调
const onBotMessage = async (event: DWClientDownStream) => {
  const message = JSON.parse(event.data);
  const content = (message?.text?.content || "").trim();
  console.log("received message: ", content);

  // 卡片模板 ID
  const cardTemplateId = "fcc1df51-17bb-403f-aca9-65f1c6919129.schema"; // 该模板只用于测试使用，如需投入线上使用，请导入卡片模板 json 到自己的应用下
  // 卡片公有数据，非字符串类型的卡片数据参考文档：https://open.dingtalk.com/document/orgapp/instructions-for-filling-in-api-card-data
  const cardData: Record<string, any> = {
    last_message: "事件链演示",
    markdown:
      "<font colorTokenV2=common_green1_color>动态显示的 markdown 内容</font>",
  };

  const cardInstance = new CardReplier(client, message);
  // 创建并投放卡片: https://open.dingtalk.com/document/orgapp/create-and-deliver-cards
  const cardInstanceId = await cardInstance.createAndDeliverCard({
    cardTemplateId,
    cardData: convertJSONValuesToString(cardData),
  });

  console.log("reply card: ", cardInstanceId, cardData);

  client.socketCallBackResponse(event.headers.messageId, EventAck.SUCCESS);
};

// 卡片回传请求回调
const onCardCallback = async (event: DWClientDownStream) => {
  /**
   * 卡片事件回调文档：https://open.dingtalk.com/document/orgapp/event-callback-card
   */
  const message = JSON.parse(event.data);
  console.log("card callback message: ", message);

  const updateCardData: Record<string, any> = {};
  const userPrivateData: Record<string, any> = {};

  const cardPrivateData = JSON.parse(message.content).cardPrivateData;
  const params = cardPrivateData.params;
  const variable = params.var;

  if (variable === "pub_url") {
    updateCardData.pub_url = "";
    updateCardData.pub_url_msg = "";
    updateCardData.pub_url_status = Math.round(Math.random())
      ? "success"
      : "failed";
    if (updateCardData.pub_url_status === "success") {
      updateCardData.pub_url =
        "dingtalk://dingtalkclient/page/link?web_wnd=workbench&pc_slide=true&hide_bar=true&url=https://www.dingtalk.com";
    } else {
      updateCardData.pub_url_msg = `更新失败${Math.floor(Math.random() * 101)}`;
    }
  } else if (variable === "pri_url") {
    userPrivateData.pri_url = "";
    userPrivateData.pri_url_msg = "";
    userPrivateData.pri_url_status = Math.round(Math.random())
      ? "success"
      : "failed";
    if (userPrivateData.pri_url_status === "success") {
      userPrivateData.pri_url =
        "dingtalk://dingtalkclient/page/link?web_wnd=workbench&pc_slide=true&hide_bar=true&url=https://github.com/open-dingtalk/dingtalk-card-examples";
    } else {
      userPrivateData.pri_url_msg = `更新失败${Math.floor(
        Math.random() * 101
      )}`;
    }
  }

  const cardUpdateOptions = {
    updateCardDataByKey: true,
    updatePrivateDataByKey: true,
  };
  const response = {
    cardUpdateOptions,
    cardData: {
      cardParamMap: convertJSONValuesToString(updateCardData),
    },
    userPrivateData: {
      cardParamMap: convertJSONValuesToString(userPrivateData),
    },
  };

  console.log("card callback response: ", response);
  client.socketCallBackResponse(event.headers.messageId, response);
};

client
  .registerCallbackListener(TOPIC_ROBOT, onBotMessage)
  .registerCallbackListener(TOPIC_CARD, onCardCallback)
  .connect();
