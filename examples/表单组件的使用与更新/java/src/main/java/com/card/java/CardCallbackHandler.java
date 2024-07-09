package com.card.java;

import lombok.extern.slf4j.Slf4j;
import com.alibaba.fastjson.JSON;
import com.alibaba.fastjson.JSONArray;
import com.alibaba.fastjson.JSONObject;
import com.card.java.models.CardCallbackMessage;
import com.card.java.models.CardPrivateData;
import com.card.java.models.CardPrivateDataWrapper;

import org.springframework.beans.factory.annotation.Autowired;
import com.dingtalk.open.app.api.callback.OpenDingTalkCallbackListener;

import org.springframework.stereotype.Component;

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
    JSONObject userPrivateData = new JSONObject();

    CardCallbackMessage message = JSON.parseObject(messageString, CardCallbackMessage.class);
    CardPrivateDataWrapper content = JSON.parseObject(message.getContent(), CardPrivateDataWrapper.class);
    CardPrivateData cardPrivateData = content.getCardPrivateData();
    JSONObject params = cardPrivateData.getParams();

    JSONObject form = params.getJSONObject("form");
    JSONObject currentForm = params.getJSONObject("current_form");

    if (form != null && currentForm != null) {
      log.info("form: " + form);
      JSONArray fields = currentForm.getJSONArray("fields");
      for (int i = 0; i < fields.size(); i++) {
        JSONObject field = fields.getJSONObject(i);
        String fieldName = field.getString("name");
        Object submitValue = form.get(fieldName);
        if (submitValue != null) {
          field.put("defaultValue", submitValue);
        }
      }

      userPrivateData.put("form", currentForm);
      userPrivateData.put("form_btn_text", "已提交");
      userPrivateData.put("form_status", "disabled");
    }

    JSONObject cardUpdateOptions = new JSONObject();
    cardUpdateOptions.put("updateCardDataByKey", true);
    cardUpdateOptions.put("updatePrivateDataByKey", true);

    JSONObject response = new JSONObject();
    response.put("cardUpdateOptions", cardUpdateOptions);
    response.put("userPrivateData",
        new JSONObject().fluentPut("cardParamMap", jsonObjectUtils.convertJSONValuesToString(userPrivateData)));

    log.info("card callback response: " + JSON.toJSONString(response));
    return response;
  }
}
