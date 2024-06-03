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
  const cardTemplateId = "737cda86-7a7f-4d83-ba07-321e6933be12.schema"; // 该模板只用于测试使用，如需投入线上使用，请导入卡片模板 json 到自己的应用下
  // 卡片公有数据，非字符串类型的卡片数据参考文档：https://open.dingtalk.com/document/orgapp/instructions-for-filling-in-api-card-data
  const cardData: Record<string, any> = {
    lastMessage: "交互组件本地更新卡片",
    submitBtnStatus: "normal",
    submitBtnText: "提交",
    input: "",
    selectIndex: -1,
    multiSelectIndexes: [],
    date: "",
    datetime: "",
    checkbox: false,
    singleCheckboxItems: [
      { value: 0, text: "单选复选框选项 1" },
      { value: 1, text: "单选复选框选项 2" },
      { value: 2, text: "单选复选框选项 3" },
      { value: 3, text: "单选复选框选项 4" },
    ],
    multiCheckboxItems: [
      { value: 0, text: "多选复选框选项 1" },
      { value: 1, text: "多选复选框选项 2" },
      { value: 2, text: "多选复选框选项 3" },
      { value: 3, text: "多选复选框选项 4" },
    ],
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
  const message = JSON.parse(event.data);
  console.log("card callback message: ", message);

  let userPrivateData: Record<string, any> = {};
  const requiredFields = {
    input: "文本输入",
    select: "下拉单选",
    multiSelect: "下拉多选",
    date: "日期选择",
    datetime: "日期时间选择",
    singleCheckbox: "单选列表",
    multiCheckbox: "多选列表",
  };

  const cardPrivateData = JSON.parse(message.content).cardPrivateData;
  const params = cardPrivateData.params;

  let input = params.input;
  if (typeof input === "string" && input) {
    userPrivateData["input"] = input;
    delete requiredFields["input"];
  }

  let date = params.date;
  if (typeof date === "string" && date) {
    userPrivateData["date"] = date;
    delete requiredFields["date"];
  }

  let datetime = params.datetime;
  if (typeof datetime === "string" && datetime) {
    userPrivateData["datetime"] = datetime;
    delete requiredFields["datetime"];
  }

  let select = params.select;
  if (typeof select === "object" && "index" in select) {
    userPrivateData["selectIndex"] = select.index;
    delete requiredFields["select"];
  }

  let multiSelect = params.multiSelect;
  if (
    typeof multiSelect === "object" &&
    "index" in multiSelect &&
    multiSelect.index.length > 0
  ) {
    userPrivateData["multiSelectIndexes"] = multiSelect.index;
    delete requiredFields["multiSelect"];
  }

  let checkbox = params.checkbox;
  if (typeof checkbox === "boolean") {
    userPrivateData["checkbox"] = checkbox;
  }

  let singleCheckbox = params.singleCheckbox;
  let singleCheckboxItems = params.singleCheckboxItems;
  if (
    typeof singleCheckbox === "number" &&
    Array.isArray(singleCheckboxItems)
  ) {
    userPrivateData["singleCheckboxItems"] = singleCheckboxItems.map(
      (item) => ({
        ...item,
        checked: singleCheckbox === item.value,
      })
    );
    delete requiredFields["singleCheckbox"];
  }

  let multiCheckbox = params.multiCheckbox;
  let multiCheckboxItems = params.multiCheckboxItems;
  if (
    Array.isArray(multiCheckbox) &&
    multiCheckbox.length > 0 &&
    Array.isArray(multiCheckboxItems)
  ) {
    userPrivateData["multiCheckboxItems"] = multiCheckboxItems.map((item) => ({
      ...item,
      checked: multiCheckbox.includes(item.value),
    }));
    delete requiredFields["multiCheckbox"];
  }

  if (Object.keys(requiredFields).length) {
    const errMsg = `表单未填写完整，${Object.values(requiredFields).join(
      "、"
    )} 是必填项`;
    console.error(errMsg);
    userPrivateData = {
      errMsg,
    };
  } else {
    userPrivateData["submitBtnText"] = "已提交";
    userPrivateData["submitBtnStatus"] = "disabled";
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
