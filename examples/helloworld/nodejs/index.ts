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
  const cardTemplateId = "2c278d79-fc0b-41b4-b14e-8b8089dc08e8.schema"; // 该模板只用于测试使用，如需投入线上使用，请导入卡片模板 json 到自己的应用下
  // 卡片公有数据，非字符串类型的卡片数据参考文档：https://open.dingtalk.com/document/orgapp/instructions-for-filling-in-api-card-data
  const cardData: Record<string, any> = {
    markdown: content,
    submitted: false,
    title: "钉钉互动卡片",
    tag: "标签",
  };

  const cardInstance = new CardReplier(client, message);
  // 创建并投放卡片
  const cardInstanceId = await cardInstance.createAndDeliverCard({
    cardTemplateId,
    cardData: convertJSONValuesToString(cardData),
  });

  console.log("reply card: ", cardInstanceId, cardData);

  // 更新卡片
  setTimeout(() => {
    const updateCardData: Record<string, any> = { tag: "更新后的标签" };
    cardInstance.putCardData({
      cardInstanceId,
      cardData: convertJSONValuesToString(updateCardData),
      cardUpdateOptions: {
        updateCardDataByKey: true,
      },
    });
    console.log("update card: ", cardInstanceId, updateCardData);
  }, 2000);

  client.socketCallBackResponse(event.headers.messageId, EventAck.SUCCESS);
};

// 卡片回传请求回调
const onCardCallback = async (event: DWClientDownStream) => {
  /**
   * 卡片事件回调文档：https://open.dingtalk.com/document/orgapp/event-callback-card
   */
  const message = JSON.parse(event.data);
  console.log("card callback message: ", message);

  const userPrivateData: Record<string, any> = {};

  const cardPrivateData = JSON.parse(message.content).cardPrivateData;
  const params = cardPrivateData.params;
  const local_input = params.local_input;

  if (local_input != null) {
    userPrivateData.private_input = local_input;
    userPrivateData.submitted = true;
  }

  const cardUpdateOptions = {
    updateCardDataByKey: true,
    updatePrivateDataByKey: true,
  };
  const response = {
    cardUpdateOptions,
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
