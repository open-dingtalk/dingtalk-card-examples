import {
  DWClient,
  DWClientDownStream,
  EventAck,
  TOPIC_ROBOT,
} from "dingtalk-stream";
import { URL } from "url";
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
  const cardTemplateId = "b23d3b9d-1c9c-4a3b-82a8-744d475c483d.schema"; // 该模板只用于测试使用，如需投入线上使用，请导入卡片模板 json 到自己的应用下
  // 卡片公有数据，非字符串类型的卡片数据参考文档：https://open.dingtalk.com/document/orgapp/instructions-for-filling-in-api-card-data
  const cardData: Record<string, any> = {
    evaluate_done: false,
    table: {
      data: [
        {
          uv: "324",
          pv: "433",
          rank: 1,
          appItem: {
            icon: "https://static.dingtalk.com/media/lALPDeC2uGvNwy3NArzNArw_700_700.png",
            name: "考勤打卡",
          },
        },
        {
          uv: "350",
          pv: "354",
          rank: 2,
          appItem: {
            icon: "https://static.dingtalk.com/media/lALPDeC2uGvNwy3NArzNArw_700_700.png",
            name: "智能人事",
          },
        },
        {
          uv: "189",
          pv: "322",
          rank: 3,
          appItem: {
            icon: "https://static.dingtalk.com/media/lALPDeC2uGvNwy3NArzNArw_700_700.png",
            name: "日志",
          },
        },
      ],
      meta: [
        {
          aliasName: "",
          dataType: "STRING",
          alias: "rank",
          weight: 10,
        },
        {
          aliasName: "应用名",
          dataType: "MICROAPP",
          alias: "appItem",
          weight: 40,
        },
        {
          aliasName: "点击次数",
          dataType: "STRING",
          alias: "pv",
          weight: 25,
        },
        {
          aliasName: "点击人数",
          dataType: "STRING",
          alias: "uv",
          weight: 25,
        },
      ],
    },
  };

  const cardInstance = new CardReplier(client, message);
  // 创建并投放卡片: https://open.dingtalk.com/document/isvapp/create-and-deliver-cards
  const cardInstanceId = await cardInstance.createAndDeliverCard({
    cardTemplateId,
    cardData: convertJSONValuesToString(cardData),
  });

  console.log("reply card: ", cardInstanceId, cardData);

  // 更新卡片: https://open.dingtalk.com/document/orgapp/interactive-card-update-interface

  const updateCardData: Record<string, any> = {
    more_detail_url: `dingtalk://dingtalkclient/page/link?pc_slide=true&url=${encodeURIComponent(
      `http://localhost:3000?page=detail&id=${cardInstanceId}`
    )}`,
    evaluate_url: `dingtalk://dingtalkclient/page/link?pc_slide=true&url=${encodeURIComponent(
      `http://localhost:3000?page=evaluate&id=${cardInstanceId}`
    )}`,
  };
  cardInstance.putCardData({
    cardInstanceId,
    cardData: convertJSONValuesToString(updateCardData),
    cardUpdateOptions: {
      updateCardDataByKey: true,
      updatePrivateDataByKey: true,
    },
  });
  console.log("update card: ", cardInstanceId, updateCardData);

  client.socketCallBackResponse(event.headers.messageId, EventAck.SUCCESS);
};

client.registerCallbackListener(TOPIC_ROBOT, onBotMessage).connect();
