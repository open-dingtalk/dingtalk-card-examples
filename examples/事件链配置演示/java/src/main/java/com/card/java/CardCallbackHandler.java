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
    JSONObject updateCardData = new JSONObject();
    JSONObject userPrivateData = new JSONObject();

    CardCallbackMessage message = JSON.parseObject(messageString, CardCallbackMessage.class);
    CardPrivateDataWrapper content = JSON.parseObject(message.getContent(), CardPrivateDataWrapper.class);
    CardPrivateData cardPrivateData = content.getCardPrivateData();
    JSONObject params = cardPrivateData.getParams();
    String variable = params.getString("var");

    if (variable.equals("pub_url")) {
      updateCardData.put("pub_url", "");
      updateCardData.put("pub_url_msg", "");
      Random rand = new Random();
      int status = rand.nextInt(2);
      if (status == 0) {
        updateCardData.put("pub_url_status", "failed");
        updateCardData.put("pub_url_msg", String.format("更新失败%s", (int)rand.nextInt(101)));
      } else {
        updateCardData.put("pub_url_status", "success");
        updateCardData.put("pub_url", "dingtalk://dingtalkclient/page/link?web_wnd=workbench&pc_slide=true&hide_bar=true&url=https://www.dingtalk.com");
      }
    } else if (variable.equals("pri_url")) {
      userPrivateData.put("pri_url", "");
      userPrivateData.put("pri_url_msg", "");
      Random rand = new Random();
      int status = rand.nextInt(2);
      if (status == 0) {
        userPrivateData.put("pri_url_status", "failed");
        userPrivateData.put("pri_url_msg", String.format("更新失败%s", (int)rand.nextInt(101)));
      } else {
        userPrivateData.put("pri_url_status", "success");
        userPrivateData.put("pri_url", "dingtalk://dingtalkclient/page/link?web_wnd=workbench&pc_slide=true&hide_bar=true&url=https://github.com/open-dingtalk/dingtalk-card-examples");
      }
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
