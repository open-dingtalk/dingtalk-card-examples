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

const valueKeyMap = {
  TEXT: "default_string",
  DATE: "default_string",
  DATETIME: "default_string",
  SELECT: "default_number",
  MULTI_SELECT: "default_number_array",
  CHECKBOX: "default_boolean",
  CHECKBOX_LIST: "checkbox_items",
  CHECKBOX_LIST_MULTI: "checkbox_items",
};
const formFieldsByInstanceId = {};

// 机器人接收消息回调
const onBotMessage = async (event: DWClientDownStream) => {
  const message = JSON.parse(event.data);
  const content = (message?.text?.content || "").trim();
  console.log("received message: ", content);

  // 卡片模板 ID
  const cardTemplateId = "9f86e003-e65e-4680-bf4b-8df5958d9f17.schema"; // 该模板只用于测试使用，如需投入线上使用，请导入卡片模板 json 到自己的应用下
  // 卡片公有数据，非字符串类型的卡片数据参考文档：https://open.dingtalk.com/document/orgapp/instructions-for-filling-in-api-card-data
  const formFields = [
    {
      type: "TEXT",
      required: true,
      name: "text_required",
      label: "必填文本输入",
      placeholder: "请输入文本",
      default_string: "",
    },
    {
      type: "TEXT",
      name: "text",
      label: "文本输入",
      placeholder: "请输入文本",
      default_string: "",
    },
    {
      type: "DATE",
      required: true,
      name: "date_required",
      label: "必填日期选择",
      placeholder: "请选择日期",
    },
    {
      type: "DATE",
      name: "date",
      label: "日期选择",
      placeholder: "请选择日期",
      default_string: "2024-06-06",
    },
    {
      type: "DATETIME",
      required: true,
      name: "datetime_required",
      label: "必填日期时间选择",
      placeholder: "请选择日期时间",
    },
    {
      type: "DATETIME",
      name: "datetime",
      label: "日期时间选择",
      placeholder: "请选择日期时间",
      default_string: "2024-06-06 12:00",
    },
    {
      type: "SELECT",
      required: true,
      name: "select_required",
      label: "必填单选下拉框",
      placeholder: "单选请选择",
      options: [
        { value: 1, text: { zh_CN: "选项 1" } },
        { value: 2, text: { zh_CN: "选项 2" } },
        { value: 3, text: { zh_CN: "选项 3" } },
        { value: 4, text: { zh_CN: "选项 4" } },
      ],
    },
    {
      type: "SELECT",
      name: "select",
      label: "单选下拉框",
      placeholder: "单选请选择",
      default_number: 1,
      options: [
        { value: 1, text: { zh_CN: "选项 1" } },
        { value: 2, text: { zh_CN: "选项 2" } },
        { value: 3, text: { zh_CN: "选项 3" } },
        { value: 4, text: { zh_CN: "选项 4" } },
      ],
    },
    {
      type: "MULTI_SELECT",
      required: true,
      name: "multi_select",
      label: "必填多选下拉框",
      placeholder: "多选请选择",
      default_number_array: [0, 2],
      options: [
        { value: 1, text: { zh_CN: "选项 1" } },
        { value: 2, text: { zh_CN: "选项 2" } },
        { value: 3, text: { zh_CN: "选项 3" } },
        { value: 4, text: { zh_CN: "选项 4" } },
      ],
    },
    {
      type: "CHECKBOX_LIST",
      required: true,
      name: "checkbox_list",
      label: "必填单选列表",
      checkbox_items: [
        {
          value: 0,
          text: "选项 0",
          checked: false,
          name: "checkbox_list",
          type: "CHECKBOX_LIST",
        },
        {
          value: 1,
          text: "选项 1",
          checked: false,
          name: "checkbox_list",
          type: "CHECKBOX_LIST",
        },
        {
          value: 2,
          text: "选项 2",
          checked: false,
          name: "checkbox_list",
          type: "CHECKBOX_LIST",
        },
        {
          value: 3,
          text: "选项 3",
          checked: false,
          name: "checkbox_list",
          type: "CHECKBOX_LIST",
        },
      ],
    },
    {
      type: "CHECKBOX_LIST_MULTI",
      required: true,
      name: "checkbox_list_multi",
      label: "必填多选列表",
      checkbox_items: [
        {
          value: 0,
          text: "选项 0",
          checked: false,
          name: "checkbox_list_multi",
          type: "CHECKBOX_LIST_MULTI",
        },
        {
          value: 1,
          text: "选项 1",
          checked: true,
          name: "checkbox_list_multi",
          type: "CHECKBOX_LIST_MULTI",
        },
        {
          value: 2,
          text: "选项 2",
          checked: false,
          name: "checkbox_list_multi",
          type: "CHECKBOX_LIST_MULTI",
        },
        {
          value: 3,
          text: "选项 3",
          checked: true,
          name: "checkbox_list_multi",
          type: "CHECKBOX_LIST_MULTI",
        },
      ],
    },
    { type: "CHECKBOX", name: "checkbox", label: "复选框" },
    {
      type: "CHECKBOX",
      name: "checkbox_default_true",
      label: "复选框默认勾选",
      default_boolean: true,
    },
  ];
  const cardData: Record<string, any> = {
    form_fields: formFields,
    form_status: "normal",
    button_text: "提交",
    title: content,
    err_msg: "",
  };

  const cardInstance = new CardReplier(client, message);
  // 创建并投放卡片: https://open.dingtalk.com/document/orgapp/create-and-deliver-cards
  const cardInstanceId = await cardInstance.createAndDeliverCard({
    cardTemplateId,
    cardData: convertJSONValuesToString(cardData),
  });
  formFieldsByInstanceId[cardInstanceId] = formFields;

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
  const actionId = cardPrivateData.actionIds[0] || "";
  const params = cardPrivateData.params;

  const submitFormFields = params.submit_form_fields || [];

  if (submitFormFields.length > 0) {
    // 提交表单，做必填校验，响应错误提示或者响应提交成功处理
    const requiredErrorLabels: string[] = [];
    for (const formField of submitFormFields) {
      const formFieldType = formField.type;
      if (valueKeyMap[formFieldType] == null) {
        userPrivateData.err_msg = `无效的表单类型「${formFieldType}」`;
        break;
      }
      const formFieldValue = formField[valueKeyMap[formFieldType]];
      const formFieldLabel = formField.label;
      if (["CHECKBOX_LIST", "CHECKBOX_LIST_MULTI"].includes(formFieldType)) {
        if (
          formField.required &&
          (formFieldValue || []).filter((x: any) => !!x.checked).length === 0
        ) {
          requiredErrorLabels.push(formFieldLabel);
        }
      } else if (formFieldType === "SELECT") {
        if (
          formField.required &&
          !(typeof formFieldValue === "number" && formFieldValue >= 0)
        ) {
          requiredErrorLabels.push(formFieldLabel);
        }
      } else {
        const formFieldValueTruly = Array.isArray(formFieldValue)
          ? formFieldValue.length > 0
          : !!formFieldValue;
        if (formField.required && !formFieldValueTruly) {
          requiredErrorLabels.push(formFieldLabel);
        }
      }
    }

    if (userPrivateData.err_msg == null) {
      if (requiredErrorLabels.length > 0) {
        userPrivateData.err_msg = `请填写必填项「${requiredErrorLabels.join(
          ", "
        )}」`;
      } else {
        userPrivateData.form_status = "disabled";
        userPrivateData.button_text = "已提交";
      }
    }
  } else {
    // 更新表单项
    let updateName = params.name;
    console.log(`update name=${updateName}`);
    if (updateName) {
      const cardInstanceId = message.outTrackId;
      const formFields = formFieldsByInstanceId[cardInstanceId] || [];
      for (const formField of formFields) {
        if (formField.name === updateName) {
          let updateType: string = params.type;
          let updateKey: string = valueKeyMap[updateType];
          let updateValue: any;
          if (params.remove && updateType === "multiSelect") {
            updateType = "MULTI_SELECT";
            updateName = actionId;
            updateKey = valueKeyMap[updateType];
            const removeIndex = params[actionId]?.index;
            updateValue = (formField[updateKey] || []).filter(
              (x: number) => x !== removeIndex
            );
          } else if (updateType === "CHECKBOX_LIST") {
            updateValue = (formField[updateKey] || []).map((x: any) => ({
              ...x,
              checked: x.value === params.value,
            }));
          } else if (updateType === "CHECKBOX_LIST_MULTI") {
            updateValue = (formField[updateKey] || []).map((x: any) => ({
              ...x,
              checked: x.value === params.value ? !x.checked : x.checked,
            }));
          } else {
            updateValue = params[updateName];
            if (["SELECT", "MULTI_SELECT"].includes(updateType)) {
              updateValue = updateValue.index;
            } else if (updateType === "CHECKBOX") {
              updateValue = !formField[updateKey];
            }
          }
          console.log(
            `update name=${updateName}, type=${updateType}, key=${updateKey}, value=${updateValue}`
          );
          formField[updateKey] = updateValue;
        }
      }
      userPrivateData.form_fields = formFields;
    } else {
      userPrivateData.err_msg = "服务异常";
    }
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
