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
  const cardTemplateId = "db56f2c2-f609-4878-9a34-46f6a0194a73.schema"; // 该模板只用于测试使用，如需投入线上使用，请导入卡片模板 json 到自己的应用下
  // 卡片公有数据，非字符串类型的卡片数据参考文档：https://open.dingtalk.com/document/orgapp/instructions-for-filling-in-api-card-data
  const cardData: Record<string, any> = {
    lastMessage: "审批",
    title: "朱小志提交的财务报销",
    type: "差旅费",
    amount: "1000元",
    reason: "出差费用",
    createTime: "2023-10-10 10:10:10",
    status: "",
  };
  const cardInstance = new CardReplier(client, message);
  // 创建并投放卡片
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

  const updateCardData: Record<string, any> = {};  // 更新公有数据
  const userPrivateData: Record<string, any> = {};  // 更新触发回传请求事件的人的私有数据

  const cardPrivateData = JSON.parse(message.content).cardPrivateData;
  const params = cardPrivateData.params;
  const action = params.action;

  if (action === "agree" || action === "reject") {
    updateCardData.status = action;
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
