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
  const cardTemplateId = "280f6d7a-63bc-4905-bf3f-4c6d95e5166b.schema"; // 该模板只用于测试使用，如需投入线上使用，请导入卡片模板 json 到自己的应用下
  // 卡片公有数据，非字符串类型的卡片数据参考文档：https://open.dingtalk.com/document/orgapp/instructions-for-filling-in-api-card-data
  const cardData: Record<string, any> = {
    form: {
      fields: [
        {
          name: "system_params_1",
          type: "TEXT",
          hidden: true,
          defaultValue: "asdf",
        },
        {
          name: "text",
          label: "必填文本输入",
          type: "TEXT",
          required: true,
          placeholder: "请输入文本",
          requiredMsg: "自定义必填错误提示",
        },
        {
          name: "text_optional",
          label: "非必填文本输入",
          type: "TEXT",
          placeholder: "请输入文本",
        },
        {
          name: "text_readonly",
          label: "非必填只读文本输入有默认值",
          type: "TEXT",
          readOnly: true,
          defaultValue: "文本默认值",
        },
        {
          name: "date",
          label: "必填日期选择",
          type: "DATE",
          required: true,
          placeholder: "请选择日期",
        },
        {
          name: "date_optional",
          label: "非必填日期选择",
          type: "DATE",
          placeholder: "请选择日期",
        },
        {
          name: "date_readonly",
          label: "非必填只读日期选择有默认值",
          type: "DATE",
          readOnly: true,
          defaultValue: "2024-05-27",
        },
        {
          name: "datetime",
          label: "必填日期时间选择",
          type: "DATETIME",
          required: true,
          placeholder: "请选择日期时间",
        },
        {
          name: "datetime_optional",
          label: "非必填日期时间选择",
          type: "DATETIME",
          placeholder: "请选择日期时间",
        },
        {
          name: "datetime_readonly",
          label: "非必填只读日期时间选择有默认值",
          type: "DATETIME",
          readOnly: true,
          defaultValue: "2024-05-27 12:00",
        },
        {
          name: "select",
          label: "必填单选",
          type: "SELECT",
          required: true,
          placeholder: "单选请选择",
          options: [
            { value: "1", text: "选项1" },
            { value: "2", text: "选项2" },
            { value: "3", text: "选项3" },
            { value: "4", text: "选项4" },
          ],
        },
        {
          name: "select_optional",
          label: "非必填单选",
          type: "SELECT",
          placeholder: "单选请选择",
          options: [
            { value: "1", text: "选项1" },
            { value: "2", text: "选项2" },
            { value: "3", text: "选项3" },
            { value: "4", text: "选项4" },
          ],
        },
        {
          name: "select_readonly",
          label: "非必填只读单选有默认值",
          type: "SELECT",
          readOnly: true,
          defaultValue: { index: 3, value: "4" },
          options: [
            { value: "1", text: "选项1" },
            { value: "2", text: "选项2" },
            { value: "3", text: "选项3" },
            { value: "4", text: "选项4" },
          ],
        },
        {
          name: "multi_select",
          label: "必填多选",
          type: "MULTI_SELECT",
          required: true,
          placeholder: "多选请选择",
          options: [
            { value: "1", text: "选项1" },
            { value: "2", text: "选项2" },
            { value: "3", text: "选项3" },
            { value: "4", text: "选项4" },
          ],
        },
        {
          name: "multi_select_optional",
          label: "非必填多选",
          type: "MULTI_SELECT",
          placeholder: "多选请选择",
          options: [
            { value: "1", text: "选项1" },
            { value: "2", text: "选项2" },
            { value: "3", text: "选项3" },
            { value: "4", text: "选项4" },
          ],
        },
        {
          name: "multi_select_readonly",
          label: "非必填只读多选有默认值",
          type: "MULTI_SELECT",
          readOnly: true,
          defaultValue: { index: [1, 3], value: ["2", "4"] },
          options: [
            { value: "1", text: "选项1" },
            { value: "2", text: "选项2" },
            { value: "3", text: "选项3" },
            { value: "4", text: "选项4" },
          ],
        },
        { name: "checkbox", label: "独立的复选框", type: "CHECKBOX" },
        {
          name: "checkbox_readonly",
          label: "只读独立的复选框",
          type: "CHECKBOX",
          readOnly: true,
          defaultValue: true,
        },
      ],
    },
    form_status: "normal",
    form_btn_text: "提交",
    title: content,
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

  const userPrivateData: Record<string, any> = {};

  const cardPrivateData = JSON.parse(message.content).cardPrivateData;
  const params = cardPrivateData.params;

  const form = params.form
  const currentForm = params.current_form
  if (form && currentForm) {
    console.log("form: ", form)
    for (const field of (currentForm?.fields || [])) {
      const submitValue = form[field.name]
      if (submitValue != null) {
        field["defaultValue"] = submitValue
      }
    }
    userPrivateData["form"] = currentForm
    userPrivateData["form_btn_text"] = "已提交"
    userPrivateData["form_status"] = "disabled"
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
