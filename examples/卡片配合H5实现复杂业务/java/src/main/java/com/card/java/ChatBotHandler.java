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

import java.net.URLEncoder;
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

  public void updateCard(String cardInstanceId, JSONObject cardData, JSONObject options)
      throws IOException, InterruptedException, NoSuchAlgorithmException {
    JSONObject data = new JSONObject().fluentPut("outTrackId", cardInstanceId);
    // 构造 cardData
    JSONObject cardDataObject = new JSONObject();
    cardDataObject.put("cardParamMap", cardData);
    data.put("cardData", cardDataObject);

    for (String key : options.keySet()) {
      data.put(key, options.get(key));
    }

    String url = openApiHost + "/v1.0/card/instances";

    OkHttpClient client = new OkHttpClient();
    MediaType JSON = MediaType.get("application/json; charset=utf-8");
    RequestBody body = RequestBody.create(data.toJSONString(), JSON);

    Request request = new Request.Builder()
        .url(url)
        .addHeader("Content-Type", "application/json")
        .addHeader("Accept", "*/*")
        .addHeader("x-acs-dingtalk-access-token", accessTokenService.getAccessToken())
        .put(body)
        .build();

    try {
      Response response = client.newCall(request).execute();
      log.info("update card: " + data.toJSONString());
      if (response.code() != 200) {
        log.error("update card failed: " + response.code() + " " + response.body().string());
      }
    } catch (IOException e) {
      e.printStackTrace();
    }
  }

  @Override
  public Void execute(ChatbotMessage message) {

    String receivedMessage = message.getText().getContent().trim();
    log.info("received message: " + receivedMessage);

    try {
      // 卡片模板 ID
      String cardTemplateId = "b23d3b9d-1c9c-4a3b-82a8-744d475c483d.schema"; // 该模板只用于测试使用，如需投入线上使用，请导入卡片模板 json 到自己的应用下
      // 卡片公有数据，非字符串类型的卡片数据参考文档：https://open.dingtalk.com/document/orgapp/instructions-for-filling-in-api-card-data
      JSONObject cardData = new JSONObject();
      cardData.put("evaluate_done", false);
      JSONObject table = new JSONObject();
      JSONArray meta = new JSONArray();
      meta.add(new JSONObject().fluentPut("aliasName", "").fluentPut("dataType", "STRING").fluentPut("alias", "rank")
          .fluentPut("weight", 10));
      meta.add(
          new JSONObject().fluentPut("aliasName", "应用名").fluentPut("dataType", "MICROAPP").fluentPut("alias", "appItem")
              .fluentPut("weight", 40));
      meta.add(new JSONObject().fluentPut("aliasName", "点击次数").fluentPut("dataType", "STRING").fluentPut("alias", "pv")
          .fluentPut("weight", 25));
      meta.add(new JSONObject().fluentPut("aliasName", "点击人数").fluentPut("dataType", "STRING").fluentPut("alias", "uv")
          .fluentPut("weight", 25));
      JSONArray data = new JSONArray();
      data.add(new JSONObject().fluentPut("uv", "324").fluentPut("pv", "433").fluentPut("rank", 1).fluentPut("appItem",
          new JSONObject().fluentPut("icon", "https://static.dingtalk.com/media/lALPDeC2uGvNwy3NArzNArw_700_700.png")
              .fluentPut("name", "考勤打卡")));
      data.add(new JSONObject().fluentPut("uv", "350").fluentPut("pv", "354").fluentPut("rank", 2).fluentPut("appItem",
          new JSONObject().fluentPut("icon", "https://static.dingtalk.com/media/lALPDeC2uGvNwy3NArzNArw_700_700.png")
              .fluentPut("name", "智能人事")));
      data.add(new JSONObject().fluentPut("uv", "189").fluentPut("pv", "322").fluentPut("rank", 3).fluentPut("appItem",
          new JSONObject().fluentPut("icon", "https://static.dingtalk.com/media/lALPDeC2uGvNwy3NArzNArw_700_700.png")
              .fluentPut("name", "日志")));
      table.put("data", data);
      table.put("meta", meta);
      cardData.put("table", table);

      // 创建并投放卡片: https://open.dingtalk.com/document/isvapp/create-and-deliver-cards
      JSONObject options = new JSONObject();
      String cardInstanceId = createAndDeliverCard(message, cardTemplateId,
          jsonObjectUtils.convertJSONValuesToString(cardData), options);

      // 更新卡片:
      // https://open.dingtalk.com/document/orgapp/interactive-card-update-interface
      JSONObject updateCardData = new JSONObject();
      String moreDetailUrl = String.format("dingtalk://dingtalkclient/page/link?pc_slide=true&url=%s",
          URLEncoder.encode("http://localhost:3000?page=detail&id=" + cardInstanceId, "UTF-8"));
      String evaluateUrl = String.format("dingtalk://dingtalkclient/page/link?pc_slide=true&url=%s",
          URLEncoder.encode("http://localhost:3000?page=evaluate&id=" + cardInstanceId, "UTF-8"));
      updateCardData.put("more_detail_url", moreDetailUrl);
      updateCardData.put("evaluate_url", evaluateUrl);
      JSONObject updateOptions = new JSONObject();
      JSONObject cardUpdateOptions = new JSONObject();
      cardUpdateOptions.put("updateCardDataByKey", true);
      cardUpdateOptions.put("updatePrivateDataByKey", true);
      updateOptions.put("cardUpdateOptions", cardUpdateOptions);
      updateCard(cardInstanceId, jsonObjectUtils.convertJSONValuesToString(updateCardData), updateOptions);

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
