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

function getCurrentDateTime() {
  const now = new Date();
  const year = now.getFullYear();
  const month = String(now.getMonth() + 1).padStart(2, "0");
  const day = String(now.getDate()).padStart(2, "0");
  const hours = String(now.getHours()).padStart(2, "0");
  const minutes = String(now.getMinutes()).padStart(2, "0");
  const seconds = String(now.getSeconds()).padStart(2, "0");
  return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`;
}

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
  const cardTemplateId = "3d667b86-d30b-43ef-be8c-7fca37965210.schema"; // 该模板只用于测试使用，如需投入线上使用，请导入卡片模板 json 到自己的应用下
  // 卡片公有数据，非字符串类型的卡片数据参考文档：https://open.dingtalk.com/document/orgapp/instructions-for-filling-in-api-card-data
  const cardData: Record<string, any> = {
    title: content,
    joined: false,
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

const contentById: Record<string, any[]> = {};

// 卡片回传请求回调
const onCardCallback = async (event: DWClientDownStream) => {
  /**
   * 卡片事件回调文档：https://open.dingtalk.com/document/orgapp/event-callback-card
   */
  const message = JSON.parse(event.data);
  const userId = message.userId;
  console.log("card callback message: ", message);

  const cardInstanceId = message.outTrackId;
  const cardInstance = new CardReplier(client, message);

  const cardPrivateData = JSON.parse(message.content).cardPrivateData;
  const params = cardPrivateData.params;
  const currentContent = contentById[cardInstanceId] || [];
  const deleteUid = params.delete_uid;
  const userPrivateData: Record<string, any> = { uid: userId };
  let nextContent = [];
  if (deleteUid) {
    // 取消接龙
    userPrivateData.joined = false;
    nextContent = currentContent.filter((item: any) => item.uid !== deleteUid);
  } else {
    // 参与接龙
    userPrivateData.joined = true;
    const body: Record<string, any> = {
      timestamp: getCurrentDateTime(),
      uid: userId,
      remark: params.remark,
      nick: "",
      avatar: "",
    };
    const userInfo = await cardInstance.getUserInfoByUserId(userId);
    console.log(`get userInfo by userId ${userId}: ${userInfo}`);
    if (userInfo) {
      body.nick = userInfo.name || "";
      body.avatar = userInfo.avatar || "";
    }
    currentContent.push(body);
    nextContent = currentContent;
  }
  contentById[cardInstanceId] = nextContent;

  const cardUpdateOptions = {
    updateCardDataByKey: true,
    updatePrivateDataByKey: true,
  };

  // 更新接龙列表和参与状态
  const updateCardData = { content: nextContent };
  console.log(
    `update data: cardData.cardParamMap=${updateCardData}, userPrivateData.cardParamMap=${userPrivateData}`
  );
  client.socketCallBackResponse(event.headers.messageId, {
    cardData: { cardParamMap: convertJSONValuesToString(updateCardData) },
    userPrivateData: {
      cardParamMap: convertJSONValuesToString(userPrivateData),
    },
    cardUpdateOptions,
  });
};

client
  .registerCallbackListener(TOPIC_ROBOT, onBotMessage)
  .registerCallbackListener(TOPIC_CARD, onCardCallback)
  .connect();
