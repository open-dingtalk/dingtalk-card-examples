package com.card.java;

import lombok.extern.slf4j.Slf4j;
import com.alibaba.fastjson.JSON;
import com.alibaba.fastjson.JSONObject;
import com.card.java.models.CardCallbackMessage;
import com.card.java.models.CardPrivateData;
import com.card.java.models.CardPrivateDataWrapper;

import org.springframework.beans.factory.annotation.Autowired;
import com.dingtalk.open.app.api.callback.OpenDingTalkCallbackListener;

import org.springframework.stereotype.Component;

import java.util.Random;

@Slf4j
@Component
public class CardCallbackHandler implements OpenDingTalkCallbackListener<String, JSONObject> {

  @Autowired
  private JSONObjectUtils jsonObjectUtils;

  @Override
  public JSONObject execute(String messageString) {
    /**
     * 卡片事件回调文档：https://open.dingtalk.com/document/orgapp/event-callback-card
     */
    log.info("card callback message: " + messageString);
    JSONObject updateCardData = new JSONObject(); // 更新公有数据
    JSONObject userPrivateData = new JSONObject(); // 更新触发回传请求事件的人的私有数据

    CardCallbackMessage message = JSON.parseObject(messageString, CardCallbackMessage.class);
    CardPrivateDataWrapper content = JSON.parseObject(message.getContent(), CardPrivateDataWrapper.class);
    CardPrivateData cardPrivateData = content.getCardPrivateData();
    JSONObject params = cardPrivateData.getParams();
    String action = params.getString("action");

    if (action.equals("agree") || action.equals("reject")) {
      updateCardData.put("status", action);
    }

    JSONObject cardUpdateOptions = new JSONObject();
    cardUpdateOptions.put("updateCardDataByKey", true);
    cardUpdateOptions.put("updatePrivateDataByKey", true);

    JSONObject response = new JSONObject();
    response.put("cardUpdateOptions", cardUpdateOptions);
    response.put("cardData",
        new JSONObject().fluentPut("cardParamMap", jsonObjectUtils.convertJSONValuesToString(updateCardData)));
    response.put("userPrivateData",
        new JSONObject().fluentPut("cardParamMap", jsonObjectUtils.convertJSONValuesToString(userPrivateData)));

    log.info("card callback response: " + JSON.toJSONString(response));
    return response;
  }
}
