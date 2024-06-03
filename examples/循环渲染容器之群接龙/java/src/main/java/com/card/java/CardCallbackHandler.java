package com.card.java;

import lombok.extern.slf4j.Slf4j;
import com.alibaba.fastjson.JSON;
import com.alibaba.fastjson.JSONArray;
import com.alibaba.fastjson.JSONObject;
import com.card.java.models.CardCallbackMessage;
import com.card.java.models.CardPrivateData;
import com.card.java.models.CardPrivateDataWrapper;

import okhttp3.MediaType;
import okhttp3.OkHttpClient;
import okhttp3.Request;
import okhttp3.RequestBody;
import okhttp3.Response;

import java.io.IOException;
import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.security.NoSuchAlgorithmException;

import org.springframework.beans.factory.annotation.Autowired;
import com.dingtalk.open.app.api.callback.OpenDingTalkCallbackListener;

import org.springframework.stereotype.Component;

@Slf4j
@Component
public class CardCallbackHandler implements OpenDingTalkCallbackListener<String, JSONObject> {

  @Autowired
  private JSONObjectUtils jsonObjectUtils;

  @Autowired
  private AccessTokenService accessTokenService;

  @Autowired
  private ChatBotHandler chatbotHandler;

  private final JSONObject contentById = new JSONObject();

  public JSONObject getUserInfoByUserId(String userId)
      throws IOException {
    String accessToken = accessTokenService.getAccessToken();
    String url = "https://oapi.dingtalk.com/topapi/v2/user/get?access_token=" + accessToken;
    JSONObject data = new JSONObject();
    data.put("userid", userId);
    OkHttpClient client = new OkHttpClient();
    RequestBody body = RequestBody.create(data.toJSONString(), MediaType.get("application/json; charset=utf-8"));

    Request request = new Request.Builder()
        .url(url)
        .addHeader("Content-Type", "application/json")
        .addHeader("Accept", "*/*")
        .post(body)
        .build();

    try {
      Response response = client.newCall(request).execute();
      String responseStr = response.body().string();
      JSONObject responseJson = JSON.parseObject(responseStr);
      log.info("get userinfo by userid: " + responseStr);
      if (responseJson != null) {
        if (responseJson.getInteger("errcode") == 0) {
          return responseJson.getJSONObject("result");
        } else {
          log.error("get userinfo by userid failed: " + responseJson.getString("errcode") + " "
              + responseJson.getString("errmsg"));
          return null;
        }
      }
      return null;
    } catch (IOException e) {
      e.printStackTrace();
      throw e;
    }
  }

  @Override
  public JSONObject execute(String messageString) {
    log.info("card callback message: " + messageString);
    try {
      CardCallbackMessage message = JSON.parseObject(messageString, CardCallbackMessage.class);
      String userId = message.getUserId();
      String outTrackId = message.getOutTrackId();
      CardPrivateDataWrapper content = JSON.parseObject(message.getContent(), CardPrivateDataWrapper.class);
      CardPrivateData cardPrivateData = content.getCardPrivateData();
      JSONObject params = cardPrivateData.getParams();

      JSONArray currentContent = contentById.getJSONArray(outTrackId);
      if (currentContent == null) {
        currentContent = new JSONArray();
      }
      log.info("contentById: " + contentById.toJSONString());
      JSONObject userPrivateData = new JSONObject();
      userPrivateData.put("uid", userId);
      JSONArray nextContent = new JSONArray();
      String deleteUid = params.getString("delete_uid");

      if (deleteUid != null) {
        // 取消接龙
        userPrivateData.put("joined", false);
        for (int i = 0; i < currentContent.size(); i++) {
          JSONObject item = currentContent.getJSONObject(i);
          String uid = item.getString("uid");
          if (!deleteUid.equals(uid)) {
            nextContent.add(item);
          }
        }
      } else {
        // 参与接龙
        userPrivateData.put("joined", true);
        JSONObject body = new JSONObject();
        LocalDateTime now = LocalDateTime.now();
        DateTimeFormatter formatter = DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss");
        String formattedNow = now.format(formatter);
        body.put("timestamp", formattedNow);
        body.put("uid", userId);
        body.put("remark", params.getString("remark"));
        body.put("nick", "");
        body.put("avatar", "");
        JSONObject userInfo = getUserInfoByUserId(userId);
        if (userInfo != null) {
          body.put("nick", userInfo.getString("name"));
          body.put("avatar", userInfo.getString("avatar"));
        }
        currentContent.add(body);
        nextContent = currentContent;
      }
      contentById.put(outTrackId, nextContent);

      // 更新接龙列表和参与状态
      JSONObject cardUpdateOptions = new JSONObject();
      cardUpdateOptions.put("updateCardDataByKey", true);
      cardUpdateOptions.put("updatePrivateDataByKey", true);

      JSONObject updateCardData = new JSONObject();
      updateCardData.put("content", nextContent);
      JSONObject updateOptions = new JSONObject();
      updateOptions.put("cardUpdateOptions", cardUpdateOptions);
      updateOptions.put("privateData", new JSONObject().fluentPut(userId,
          new JSONObject().fluentPut("cardParamMap", jsonObjectUtils.convertJSONValuesToString(userPrivateData))));

      chatbotHandler.updateCard(outTrackId, jsonObjectUtils.convertJSONValuesToString(updateCardData), updateOptions);
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
