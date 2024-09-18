package com.card.java;

import lombok.extern.slf4j.Slf4j;
import com.alibaba.fastjson.JSON;
import com.alibaba.fastjson.JSONObject;
import com.alibaba.fastjson.JSONArray;
import com.card.java.models.CardCallbackMessage;

import org.springframework.beans.factory.annotation.Autowired;
import com.dingtalk.open.app.api.callback.OpenDingTalkCallbackListener;

import org.springframework.stereotype.Component;

import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;

@Slf4j
@Component
public class CardDynamicDataCallbackHandler implements OpenDingTalkCallbackListener<String, JSONObject> {

  @Autowired
  private JSONObjectUtils jsonObjectUtils;

  @Override
  public JSONObject execute(String messageString) {
    /**
     * 卡片事件回调文档：https://open.dingtalk.com/document/orgapp/event-callback-card
     */
    log.info("card callback message: " + messageString);

    CardCallbackMessage message = JSON.parseObject(messageString, CardCallbackMessage.class);
    String cardInstanceId = message.getOutTrackId();

    JSONObject response = new JSONObject();

    if (!ChatBotHandler.totalByCardInstanceId.containsKey(cardInstanceId)) {
      return response;
    }

    Integer total = ChatBotHandler.totalByCardInstanceId.get(cardInstanceId);
    Integer finished = ChatBotHandler.finishedByCardInstanceId.get(cardInstanceId) + 1;
    ChatBotHandler.finishedByCardInstanceId.put(cardInstanceId, finished);

    if (finished > total) {
      return response;
    }

    LocalDateTime now = LocalDateTime.now();
    DateTimeFormatter formatter = DateTimeFormatter.ofPattern("MM-dd HH:mm:ss");

    JSONObject cardData = new JSONObject();
    cardData.put("finished", finished);
    cardData.put("unfinished", total - finished);
    cardData.put("progress", (int) Math.round((double) finished / total * 100));
    cardData.put("update_at", now.format(formatter));

    response.put("dataSourceQueryResponses",
        new JSONArray().fluentAdd(new JSONObject()
            .fluentPut("data", JSON.toJSONString(jsonObjectUtils.convertJSONValuesToString(cardData)))
            .fluentPut("dynamicDataSourceId", ChatBotHandler.demoDynamicDataSourceId)
            .fluentPut("dynamicDataValueType", "OBJECT")));

    log.info("card callback response: " + JSON.toJSONString(response));
    return response;
  }
}
