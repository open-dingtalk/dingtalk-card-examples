package com.card.java;

import com.alibaba.fastjson.JSONArray;
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
import java.util.Random;
import java.util.concurrent.TimeUnit;

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

  public void streaming(
      String cardInstanceId,
      String contentKey,
      String contentValue,
      Boolean isFull,
      Boolean isFinalize,
      Boolean isError) {

    JSONObject data = new JSONObject().fluentPut("outTrackId", cardInstanceId);
    data.put("key", contentKey);
    data.put("content", contentValue);
    data.put("isFull", isFull);
    data.put("isFinalize", isFinalize);
    data.put("isError", isError);
    data.put("guid", UUID.randomUUID().toString());

    String url = openApiHost + "/v1.0/card/streaming";

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
      log.info("streaming update card: " + data.toJSONString());
      if (response.code() != 200) {
        log.error("streaming update card failed: " + response.code() + " " + response.body().string());
      }
    } catch (IOException e) {
      e.printStackTrace();
    }
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
      String cardTemplateId = "2b07a3e6-cdf4-4e8f-8bac-a4f6e3e19eff.schema"; // 该模板只用于测试使用，如需投入线上使用，请导入卡片模板 json 到自己的应用下
      String contentKey = "content";
      JSONObject cardData = new JSONObject();
      cardData.put(contentKey, "");
      cardData.put("query", receivedMessage);
      cardData.put("preparations", new JSONArray());
      cardData.put("charts", new JSONArray());
      cardData.put("config", new JSONObject().fluentPut("autoLayout", true));
      // 先投放卡片
      JSONObject options = new JSONObject();
      String cardInstanceId = createAndDeliverCard(message, cardTemplateId,
          jsonObjectUtils.convertJSONValuesToString(cardData), options);
      // 流式更新卡片
      try {
        // 更新成输入中状态
        streaming(cardInstanceId, contentKey, "", true, false, false);
        // 更新卡片
        JSONObject cardUpdateOptions = new JSONObject()
            .fluentPut("updateCardDataByKey", true)
            .fluentPut("updatePrivateDataByKey", true);
        JSONObject updateOptions = new JSONObject();
        updateOptions.put("cardUpdateOptions", cardUpdateOptions);
        JSONObject updateCardData = new JSONObject();
        JSONArray preparations = new JSONArray();
        for (String action : new String[] { "正在理解需求", "正在生成 SQL", "正在执行 SQL", "正在生成图表" }) {
          JSONObject step = new JSONObject();
          step.put("name", action);
          step.put("progress", 0);
          preparations.add(step);
          for (int progress = 0; progress <= 100; progress += 20) {
            step.put("progress", progress);
            updateCardData.put("preparations", preparations);
            updateCard(cardInstanceId, jsonObjectUtils.convertJSONValuesToString(updateCardData), updateOptions);
            Thread.sleep(new Random().nextInt(1001));
          }
        }
        String[] fakeContentValues = new String[] {
            "## 过去一个月的营收分析",
            "\n本报告分析了过去一个月的营收情况。",
            "包括每周营收的趋势图、",
            "按不同产品分类的柱状图、",
            "以及不同产品的总营收占比饼图。"
        };
        StringBuilder contentValue = new StringBuilder();
        for (String fakeContentValue : fakeContentValues) {
          contentValue.append(fakeContentValue);
          streaming(cardInstanceId, contentKey, contentValue.toString(), true, false, false);
          Thread.sleep(new Random().nextInt(1001));
        }
        streaming(cardInstanceId, contentKey, contentValue.toString(), true, true, false);

        JSONObject updateChartCardData = new JSONObject();
        JSONArray charts = new JSONArray();

        JSONObject line = new JSONObject();
        String lineMd = "# 过去一个月每周营收";
        line.put("markdown", lineMd);
        JSONObject lineChart = new JSONObject();
        lineChart.put("type", "lineChart");
        lineChart.put("config", new JSONObject());
        lineChart.put("data", new JSONArray()
            .fluentAdd(createChartData("0603", 820, "产品 A"))
            .fluentAdd(createChartData("0610", 932, "产品 A"))
            .fluentAdd(createChartData("0617", 901, "产品 A"))
            .fluentAdd(createChartData("0624", 934, "产品 A"))
            .fluentAdd(createChartData("0701", 1290, "产品 A"))
            .fluentAdd(createChartData("0603", 1498, "产品 B"))
            .fluentAdd(createChartData("0610", 809, "产品 B"))
            .fluentAdd(createChartData("0617", 728, "产品 B"))
            .fluentAdd(createChartData("0624", 759, "产品 B"))
            .fluentAdd(createChartData("0701", 995, "产品 B"))
            .fluentAdd(createChartData("0603", 1249, "产品 C"))
            .fluentAdd(createChartData("0610", 873, "产品 C"))
            .fluentAdd(createChartData("0617", 1086, "产品 C"))
            .fluentAdd(createChartData("0624", 908, "产品 C"))
            .fluentAdd(createChartData("0701", 972, "产品 C")));
        line.put("chart", lineChart);
        charts.add(line);
        updateChartCardData.put("charts", charts);
        updateCard(cardInstanceId, jsonObjectUtils.convertJSONValuesToString(updateChartCardData), updateOptions);
        Thread.sleep(new Random().nextInt(2001));

        JSONObject bar = new JSONObject();
        String barMd = "# 不同产品的营收";
        bar.put("markdown", barMd);
        JSONObject barChart = new JSONObject();
        barChart.put("type", "histogram");
        barChart.put("config", new JSONObject());
        barChart.put("data", new JSONArray()
            .fluentAdd(createChartData("W1", 820, "产品 A"))
            .fluentAdd(createChartData("W2", 932, "产品 A"))
            .fluentAdd(createChartData("W3", 901, "产品 A"))
            .fluentAdd(createChartData("W4", 934, "产品 A"))
            .fluentAdd(createChartData("W5", 1290, "产品 A"))
            .fluentAdd(createChartData("W1", 1498, "产品 B"))
            .fluentAdd(createChartData("W2", 809, "产品 B"))
            .fluentAdd(createChartData("W3", 728, "产品 B"))
            .fluentAdd(createChartData("W4", 759, "产品 B"))
            .fluentAdd(createChartData("W5", 995, "产品 B"))
            .fluentAdd(createChartData("W1", 1249, "产品 C"))
            .fluentAdd(createChartData("W2", 873, "产品 C"))
            .fluentAdd(createChartData("W3", 1086, "产品 C"))
            .fluentAdd(createChartData("W4", 908, "产品 C"))
            .fluentAdd(createChartData("W5", 972, "产品 C")));
        bar.put("chart", barChart);
        charts.add(bar);
        updateChartCardData.put("charts", charts);
        updateCard(cardInstanceId, jsonObjectUtils.convertJSONValuesToString(updateChartCardData), updateOptions);
        Thread.sleep(new Random().nextInt(2001));

        JSONObject pie = new JSONObject();
        String pieMd = "# 不同产品的营收占比\n\n| 产品名称 | 营收 |\n| :-: | :-: |\n| 产品 A | 4877 |\n| 产品 B | 4789 |\n| 产品 C | 5088 |";
        pie.put("markdown", pieMd);
        JSONObject pieChart = new JSONObject();
        pieChart.put("type", "pieChart");
        int[] padding = { 20, 30, 20, 30 };
        pieChart.put("config", new JSONObject().fluentPut("padding", padding));
        pieChart.put("data", new JSONArray()
            .fluentAdd(createChartData("产品 A", 4877, null))
            .fluentAdd(createChartData("产品 B", 4789, null))
            .fluentAdd(createChartData("产品 C", 5088, null)));
        pie.put("chart", pieChart);
        charts.add(pie);
        updateChartCardData.put("charts", charts);
        updateCard(cardInstanceId, jsonObjectUtils.convertJSONValuesToString(updateChartCardData), updateOptions);
      } catch (Exception e) {
        e.printStackTrace();
        streaming(cardInstanceId, contentKey, "", true, false, true);
      }
    } catch (NoSuchAlgorithmException | IOException e) {
      e.printStackTrace();
    } catch (InterruptedException e) {
      e.printStackTrace();
      Thread.currentThread().interrupt();
    }

    return null;
  }

  private JSONObject createChartData(String x, int y, String type) {
    JSONObject chartData = new JSONObject().fluentPut("x", x).fluentPut("y", y);
    if (type != null) {
      chartData.put("type", type);
    }
    return chartData;
  }
}
