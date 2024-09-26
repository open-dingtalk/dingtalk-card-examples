import {
  DWClient,
  DWClientDownStream,
  EventAck,
  TOPIC_ROBOT,
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

const demoDynamicDataSourceId = "demo_dynamic_data_source_id";

const totalByCardInstanceId: Record<string, number> = {};
const finishedByCardInstanceId: Record<string, number> = {};

const getRandInt = (start: number, end: number) => {
  return Math.floor(Math.random() * (end - start + 1)) + start;
};

const getFormatTimestamp = (): string => {
  const date = new Date(); // 获取当前时间
  const month = String(date.getMonth() + 1).padStart(2, "0"); // 月份从0开始，所以加1
  const day = String(date.getDate()).padStart(2, "0");
  const hours = String(date.getHours()).padStart(2, "0");
  const minutes = String(date.getMinutes()).padStart(2, "0");
  const seconds = String(date.getSeconds()).padStart(2, "0");

  // 格式化字符串
  return `${month}-${day} ${hours}:${minutes}:${seconds}`;
};

// 机器人接收消息回调
const onBotMessage = async (event: DWClientDownStream) => {
  const message = JSON.parse(event.data);
  const content = (message?.text?.content || "").trim();
  console.log("received message: ", content, message);

  // 卡片模板 ID
  const cardTemplateId = "c36a2fbe-ff53-44ac-a91d-dedbe3654306.schema"; // 该模板只用于测试使用，如需投入线上使用，请导入卡片模板 json 到自己的应用下
  // 卡片公有数据，非字符串类型的卡片数据参考文档：https://open.dingtalk.com/document/orgapp/instructions-for-filling-in-api-card-data
  const title = `${getRandInt(1, 12)}月迭代`;
  const total = getRandInt(100, 200);
  const finished = 0;
  const cardData: Record<string, any> = {
    title: title,
    total: total,
    finished: finished,
    unfinished: total - finished,
    progress: 0,
    update_at: "",
  };

  let pullConfig: any = {
    // 仅拉取一次数据
    pullStrategy: "ONCE",
  };

  if (content.toUpperCase() == "RENDER") {
    // 每次卡片渲染时拉取数据
    pullConfig = { pullStrategy: "RENDER" };
  }

  if (content.toUpperCase() == "INTERVAL") {
    pullConfig = {
      // 每隔 10 秒拉取一次数据
      pullStrategy: "INTERVAL",
      interval: 10,
      timeUnit: "SECONDS",
    };
  }

  const openDynamicDataConfig = {
    dynamicDataSourceConfigs: [
      {
        dynamicDataSourceId: demoDynamicDataSourceId,
        pullConfig: pullConfig,
      },
    ],
  };

  const cardInstance = new CardReplier(client, message);
  // 创建并投放卡片: https://open.dingtalk.com/document/orgapp/create-and-deliver-cards
  const cardInstanceId = await cardInstance.createAndDeliverCard({
    cardTemplateId,
    cardData: convertJSONValuesToString(cardData),
    openSpaceId: `dtv1.card//ONE_BOX.${message.conversationId}`,
    topOpenSpaceModel: { spaceType: "ONE_BOX" },
    topOpenDeliverModel: {
      expiredTimeMillis: Date.now() + 1000 * 60 * 3, // 3 分钟后过期
    },
    openDynamicDataConfig: openDynamicDataConfig,
  });

  totalByCardInstanceId[cardInstanceId] = total;
  finishedByCardInstanceId[cardInstanceId] = finished;

  console.log("reply card: ", cardInstanceId, cardData);

  client.socketCallBackResponse(event.headers.messageId, EventAck.SUCCESS);
};

// 卡片动态数据源回调
const onCardDynamicDataCallback = async (event: DWClientDownStream) => {
  /**
   * 卡片动态数据源回调文档：https://open.dingtalk.com/document/isvapp/dynamic-data-source
   */
  const message = JSON.parse(event.data);
  console.log("card callback message: ", message);

  const cardInstanceId = message.outTrackId;

  if (totalByCardInstanceId[cardInstanceId] != null) {
    const total = totalByCardInstanceId[cardInstanceId];
    finishedByCardInstanceId[cardInstanceId] += 1;
    const finished = finishedByCardInstanceId[cardInstanceId];

    if (finished <= total) {
      const cardData = {
        finished,
        unfinished: total - finished,
        progress: Math.round((finished / total) * 100),
        update_at: getFormatTimestamp(),
      };

      const response = {
        dataSourceQueryResponses: [
          {
            data: JSON.stringify(cardData),
            dynamicDataSourceId: demoDynamicDataSourceId,
            dynamicDataValueType: "OBJECT",
          },
        ],
      };

      console.log("card callback response: ", response);
      client.socketCallBackResponse(event.headers.messageId, response);
    }
  }
};

const TOPIC_DYNAMIC_DATA = "/v1.0/card/dynamicData/get";
client
  .registerCallbackListener(TOPIC_ROBOT, onBotMessage)
  .registerCallbackListener(TOPIC_DYNAMIC_DATA, onCardDynamicDataCallback)
  .connect();
