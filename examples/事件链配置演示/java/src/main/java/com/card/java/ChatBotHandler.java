package com.card.java;

import com.alibaba.fastjson.JSONObject;
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
      String cardTemplateId = "fcc1df51-17bb-403f-aca9-65f1c6919129.schema"; // 该模板只用于测试使用，如需投入线上使用，请导入卡片模板 json 到自己的应用下
      // 卡片公有数据，非字符串类型的卡片数据参考文档：https://open.dingtalk.com/document/orgapp/instructions-for-filling-in-api-card-data
      JSONObject cardData = new JSONObject();
      cardData.put("last_message", "事件链演示");
      cardData.put("markdown", "<font colorTokenV2=common_green1_color>动态显示的 markdown 内容</font>");

      // 创建并投放卡片: https://open.dingtalk.com/document/orgapp/create-and-deliver-cards
      JSONObject options = new JSONObject();
      String cardInstanceId = createAndDeliverCard(message, cardTemplateId,
          jsonObjectUtils.convertJSONValuesToString(cardData), options);
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
}
