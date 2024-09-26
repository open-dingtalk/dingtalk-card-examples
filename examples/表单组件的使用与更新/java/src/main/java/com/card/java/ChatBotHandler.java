package com.card.java;

import com.alibaba.fastjson.JSONObject;
import com.alibaba.fastjson.JSONArray;
import com.dingtalk.open.app.api.callback.OpenDingTalkCallbackListener;
import com.dingtalk.open.app.api.models.bot.ChatbotMessage;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import lombok.extern.slf4j.Slf4j;

import okhttp3.MediaType;
import okhttp3.OkHttpClient;
import okhttp3.Request;
import okhttp3.RequestBody;
import okhttp3.Response;

import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;
import java.util.UUID;

@Slf4j
@Component
public class ChatBotHandler implements OpenDingTalkCallbackListener<ChatbotMessage, Void> {

  @Autowired
  private AccessTokenService accessTokenService;

  @Autowired
  private JSONObjectUtils jsonObjectUtils;

  @Value("${openApiHost}")
  private String openApiHost;

  @Value("${dingtalk.app.client-id}")
  private String clientId;

  public static String genCardId(ChatbotMessage message) throws NoSuchAlgorithmException {
    String factor = message.getSenderId() + '_' + message.getSenderCorpId() + '_' + message.getConversationId() + '_'
        + message.getMsgId() + '_' + UUID.randomUUID().toString();
    MessageDigest digest = MessageDigest.getInstance("SHA-256");
    byte[] hash = digest.digest(factor.getBytes(StandardCharsets.UTF_8));
    StringBuilder hexString = new StringBuilder();
    for (byte b : hash) {
      String hex = Integer.toHexString(0xff & b);
      if (hex.length() == 1)
        hexString.append('0');
      hexString.append(hex);
    }
    return hexString.toString();
  }

  public String createAndDeliverCard(ChatbotMessage message, String cardTemplateId, JSONObject cardData,
      JSONObject options)
      throws IOException, InterruptedException, NoSuchAlgorithmException {
    String cardInstanceId = genCardId(message);

    boolean supportForward = (boolean) options.getOrDefault("supportForward", true);
    boolean atAll = (boolean) options.getOrDefault("alAll", false);
    boolean atSender = (boolean) options.getOrDefault("atSender", false);
    String callbackType = (String) options.getOrDefault("callbackType", "STREAM");
    String[] recipients = (String[]) options.getOrDefault("recipients", null);

    JSONObject restOptions = new JSONObject(options);
    restOptions.remove("cardTemplateId");
    restOptions.remove("cardData");
    restOptions.remove("callbackType");
    restOptions.remove("atSender");
    restOptions.remove("atAll");
    restOptions.remove("recipients");
    restOptions.remove("supportForward");

    // 构造创建并投放卡片的请求体
    JSONObject data = new JSONObject();
    data.put("cardTemplateId", cardTemplateId);
    data.put("outTrackId", cardInstanceId);
    data.put("callbackType", callbackType);

    // 构造 cardData
    JSONObject cardDataObject = new JSONObject();
    cardDataObject.put("cardParamMap", cardData);
    data.put("cardData", cardDataObject);

    // 构造 Model
    data.put("imGroupOpenSpaceModel", new JSONObject().fluentPut("supportForward", supportForward));
    data.put("imRobotOpenSpaceModel", new JSONObject().fluentPut("supportForward", supportForward));

    // 2：群聊, 1：单聊
    String conversationType = message.getConversationType();
    if (conversationType.equals("2")) {
      data.put("openSpaceId", "dtv1.card//IM_GROUP." + message.getConversationId());
      JSONObject imGroupOpenDeliverModel = new JSONObject();
      imGroupOpenDeliverModel.put("robotCode", clientId);
      if (atAll) {
        imGroupOpenDeliverModel.put("atUserIds", new JSONObject().fluentPut("@ALL", "@ALL"));
      } else if (atSender) {
        imGroupOpenDeliverModel.put("atUserIds",
            new JSONObject().fluentPut(message.getSenderStaffId(), message.getSenderNick()));
      }
      if (recipients != null) {
        imGroupOpenDeliverModel.put("recipients", recipients);
      }
      data.put("imGroupOpenDeliverModel", new JSONObject().fluentPut("robotCode", clientId));
    } else if (conversationType.equals("1")) {
      data.put("openSpaceId", "dtv1.card//IM_ROBOT." + message.getSenderStaffId());
      data.put("imRobotOpenDeliverModel", new JSONObject().fluentPut("spaceType", "IM_ROBOT"));
    }

    // 其余自定义参数
    for (String key : restOptions.keySet()) {
      data.put(key, restOptions.get(key));
    }

    String url = openApiHost + "/v1.0/card/instances/createAndDeliver";

    OkHttpClient client = new OkHttpClient();
    MediaType JSON = MediaType.get("application/json; charset=utf-8");
    RequestBody body = RequestBody.create(data.toJSONString(), JSON);

    Request request = new Request.Builder()
        .url(url)
        .addHeader("Content-Type", "application/json")
        .addHeader("Accept", "*/*")
        .addHeader("x-acs-dingtalk-access-token", accessTokenService.getAccessToken())
        .post(body)
        .build();

    try {
      Response response = client.newCall(request).execute();
      log.info("reply card: " + data.toJSONString());
      if (response.code() != 200) {
        log.error("reply card failed: " + response.code() + " " + response.body().string());
      }
    } catch (IOException e) {
      e.printStackTrace();
    }
    return cardInstanceId;
  }

  @Override
  public Void execute(ChatbotMessage message) {

    String receivedMessage = message.getText().getContent().trim();
    log.info("received message: " + receivedMessage);

    try {
      // 卡片模板 ID
      String cardTemplateId = "280f6d7a-63bc-4905-bf3f-4c6d95e5166b.schema"; // 该模板只用于测试使用，如需投入线上使用，请导入卡片模板 json 到自己的应用下
      // 卡片公有数据，非字符串类型的卡片数据参考文档：https://open.dingtalk.com/document/orgapp/instructions-for-filling-in-api-card-data
      JSONObject cardData = new JSONObject();
      cardData.put("title", receivedMessage);
      cardData.put("form_status", "normal");
      cardData.put("form_btn_text", "提交");

      JSONArray fields = new JSONArray();
      fields.add(createField("system_params_1", "TEXT", null, "asdf", true, false, false, null, null));
      fields.add(createField("text", "TEXT", "必填文本输入", null, false, false, true, "请输入文本", "自定义必填错误提示"));
      fields.add(createField("text_optional", "TEXT", "非必填文本输入", null, false, false, false, "请输入文本", null));
      fields.add(createField("text_readonly", "TEXT", "非必填只读文本输入有默认值", "文本默认值", false, true, false, null, null));
      fields.add(createField("date", "DATE", "必填日期选择", null, false, false, true, "请选择日期", null));
      fields.add(createField("date_optional", "DATE", "非必填日期选择", null, false, false, false, "请选择日期", null));
      fields.add(createField("date_readonly", "DATE", "非必填只读日期选择有默认值", "2024-05-27", false, true, false, null, null));
      fields.add(createField("datetime", "DATETIME", "必填日期时间选择", null, false, false, true, "请选择日期时间", null));
      fields.add(createField("datetime_optional", "DATETIME", "非必填日期时间选择", null, false, false, false, "请选择日期时间", null));
      fields.add(createField("datetime_readonly", "DATETIME", "非必填只读日期时间选择有默认值", "2024-05-27 12:00", false, true, false,
          null, null));

      fields.add(createFieldWithOptions("select", "SELECT", "必填单选", null, false, false, true, "单选请选择", null,
          new String[] { "1", "2", "3", "4" }, new String[] { "选项1", "选项2", "选项3", "选项4" }));
      fields.add(createFieldWithOptions("select_optional", "SELECT", "非必填单选", null, false, false, false, "单选请选择", null,
          new String[] { "1", "2", "3", "4" }, new String[] { "选项1", "选项2", "选项3", "选项4" }));

      JSONObject selectReadonlyDefaultValue = new JSONObject();
      selectReadonlyDefaultValue.put("index", 3);
      selectReadonlyDefaultValue.put("value", "4");
      fields.add(createFieldWithOptions("select_readonly", "SELECT", "非必填只读单选有默认值", selectReadonlyDefaultValue, false,
          true, false, null, null,
          new String[] { "1", "2", "3", "4" }, new String[] { "选项1", "选项2", "选项3", "选项4" }));

      fields.add(createFieldWithOptions("multi_select", "MULTI_SELECT", "必填多选", null, false, false, true, "多选请选择", null,
          new String[] { "1", "2", "3", "4" }, new String[] { "选项1", "选项2", "选项3", "选项4" }));
      fields.add(
          createFieldWithOptions("multi_select_optional", "MULTI_SELECT", "非必填多选", null, false, false, false, "多选请选择", null,
              new String[] { "1", "2", "3", "4" }, new String[] { "选项1", "选项2", "选项3", "选项4" }));

      JSONObject multiSelectReadonlyDefaultValue = new JSONObject();
      multiSelectReadonlyDefaultValue.put("index", new int[] { 1, 3 });
      multiSelectReadonlyDefaultValue.put("value", new String[] { "2", "4" });
      fields.add(createFieldWithOptions("multi_select_readonly", "MULTI_SELECT", "非必填只读多选有默认值", multiSelectReadonlyDefaultValue,
          false, true, false, null, null,
          new String[] { "1", "2", "3", "4" }, new String[] { "选项1", "选项2", "选项3", "选项4" }));

      fields.add(createField("checkbox", "CHECKBOX", "独立的复选框", null, false, false, false, null, null));
      fields.add(createField("checkbox_readonly", "CHECKBOX", "只读独立的复选框", true, false, true, false, null, null));

      JSONObject form = new JSONObject();
      form.put("fields", fields);
      cardData.put("form", form);

      // 创建并投放卡片: https://open.dingtalk.com/document/orgapp/create-and-deliver-cards
      JSONObject options = new JSONObject();
      createAndDeliverCard(message, cardTemplateId, jsonObjectUtils.convertJSONValuesToString(cardData), options);

    } catch (NoSuchAlgorithmException e) {
      e.printStackTrace();
    } catch (IOException e) {
      e.printStackTrace();
    } catch (InterruptedException e) {
      e.printStackTrace();
      Thread.currentThread().interrupt();
    }

    return null;
  }

  private static JSONObject createField(String name, String type, String label, Object defaultValue, boolean hidden,
      boolean readOnly, boolean required, String placeholder, String requiredMsg) {
    JSONObject field = new JSONObject();
    field.put("name", name);
    field.put("type", type);
    if (label != null)
      field.put("label", label);
    if (defaultValue != null)
      field.put("defaultValue", defaultValue);
    if (hidden)
      field.put("hidden", true);
    if (readOnly)
      field.put("readOnly", true);
    if (required)
      field.put("required", true);
    if (placeholder != null)
      field.put("placeholder", placeholder);
    if (requiredMsg != null)
      field.put("requiredMsg", requiredMsg);
    return field;
  }

  private static JSONObject createFieldWithOptions(String name, String type, String label, Object defaultValue,
      boolean hidden,
      boolean readOnly, boolean required, String placeholder, String requiredMsg, String[] values, String[] texts) {
    JSONObject field = createField(name, type, label, defaultValue, hidden, readOnly, required, placeholder,
        requiredMsg);
    JSONArray options = new JSONArray();
    for (int i = 0; i < values.length; i++) {
      JSONObject option = new JSONObject();
      option.put("value", values[i]);
      option.put("text", texts[i]);
      options.add(option);
    }
    field.put("options", options);
    return field;
  }
}
