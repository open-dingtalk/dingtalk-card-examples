package com.card.java;

import lombok.extern.slf4j.Slf4j;
import com.alibaba.fastjson.JSON;
import com.alibaba.fastjson.JSONArray;
import com.alibaba.fastjson.JSONObject;
import com.card.java.models.CardCallbackMessage;
import com.card.java.models.CardPrivateData;
import com.card.java.models.CardPrivateDataWrapper;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

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
    log.info("card callback message: " + messageString);
    JSONObject userPrivateData = new JSONObject();

    CardCallbackMessage message = JSON.parseObject(messageString, CardCallbackMessage.class);
    CardPrivateDataWrapper content = JSON.parseObject(message.getContent(), CardPrivateDataWrapper.class);
    CardPrivateData cardPrivateData = content.getCardPrivateData();
    JSONObject params = cardPrivateData.getParams();

    Map<String, String> requiredFields = new HashMap<>();
    requiredFields.put("input", "文本输入");
    requiredFields.put("select", "下拉单选");
    requiredFields.put("multiSelect", "下拉多选");
    requiredFields.put("date", "日期选择");
    requiredFields.put("datetime", "日期时间选择");
    requiredFields.put("singleCheckbox", "单选列表");
    requiredFields.put("multiCheckbox", "多选列表");

    String input = (String) params.get("input");
    if (input != null && !input.isEmpty()) {
      userPrivateData.put("input", input);
      requiredFields.remove("input");
    }

    String date = (String) params.get("date");
    if (date != null && !date.isEmpty()) {
      userPrivateData.put("date", date);
      requiredFields.remove("date");
    }

    String datetime = (String) params.get("datetime");
    if (datetime != null && !datetime.isEmpty()) {
      userPrivateData.put("datetime", datetime);
      requiredFields.remove("datetime");
    }

    Object selectObj = params.get("select");
    if (selectObj instanceof String && ((String) selectObj).isEmpty()) {
      selectObj = null;
    }
    if (selectObj instanceof JSONObject) {
      JSONObject select = (JSONObject) selectObj;
      if (select.containsKey("index")) {
        userPrivateData.put("selectIndex", select.get("index"));
        requiredFields.remove("select");
      }
    }

    Object multiSelectObj = params.get("multiSelect");
    if (multiSelectObj instanceof String && ((String) multiSelectObj).isEmpty()) {
      multiSelectObj = null;
    }
    if (multiSelectObj instanceof JSONObject) {
      JSONObject multiSelect = (JSONObject) multiSelectObj;
      if (multiSelect.containsKey("index") && multiSelect.getJSONArray("index").size() > 0) {
        userPrivateData.put("multiSelectIndexes", multiSelect.get("index"));
        requiredFields.remove("multiSelect");
      }
    }

    Object checkboxObj = params.get("checkbox");
    if (checkboxObj instanceof String && ((String) checkboxObj).isEmpty()) {
      checkboxObj = null;
    }
    if (checkboxObj instanceof Boolean) {
      Boolean checkbox = (Boolean) checkboxObj;
      userPrivateData.put("checkbox", checkbox);
    }

    Object singleCheckboxObj = params.get("singleCheckbox");
    List<JSONObject> singleCheckboxItems = (List<JSONObject>) params.get("singleCheckboxItems");
    if (singleCheckboxObj instanceof String && ((String) singleCheckboxObj).isEmpty()) {
      singleCheckboxObj = null;
    } else if (singleCheckboxObj instanceof String) {
      try {
        singleCheckboxObj = Integer.parseInt((String) singleCheckboxObj);
      } catch (NumberFormatException e) {
        singleCheckboxObj = null;
      }
    }
    if (singleCheckboxObj instanceof Integer && singleCheckboxItems != null) {
      Integer singleCheckbox = (Integer) singleCheckboxObj;
      for (JSONObject item : singleCheckboxItems) {
        Boolean checked = singleCheckbox.equals(item.get("value"));
        item.put("checked", checked);
      }
      userPrivateData.put("singleCheckboxItems", singleCheckboxItems);
      requiredFields.remove("singleCheckbox");
    }

    Object multiCheckboxObj = params.get("multiCheckbox");
    List<JSONObject> multiCheckboxItems = (List<JSONObject>) params.get("multiCheckboxItems");
    if (multiCheckboxObj instanceof String && ((String) multiCheckboxObj).isEmpty()) {
      multiCheckboxObj = null;
    } else if (multiCheckboxObj instanceof JSONArray && multiCheckboxItems != null) {
      JSONArray multiCheckboxArray = (JSONArray) multiCheckboxObj;
      JSONArray multiCheckbox = new JSONArray();
      for (Object element : multiCheckboxArray) {
        if (element instanceof String) {
          try {
            multiCheckbox.add(Integer.parseInt((String) element));
          } catch (NumberFormatException ignored) {
            //
          }
        } else if (element instanceof Integer) {
          multiCheckbox.add(element);
        }
      }
      if (multiCheckbox.size() > 0) {
        for (JSONObject item : multiCheckboxItems) {
          Boolean checked = multiCheckbox.contains(item.get("value"));
          item.put("checked", checked);
        }
        userPrivateData.put("multiCheckboxItems", multiCheckboxItems);
        requiredFields.remove("multiCheckbox");
      }
    }

    if (!requiredFields.isEmpty()) {
      String errorMessage = "表单未填写完整，" + String.join("、", requiredFields.values()) + " 是必填项";
      log.error(errorMessage);
      userPrivateData.clear();
      userPrivateData.put("errMsg", errorMessage);
    } else {
      userPrivateData.put("submitBtnText", "已提交");
      userPrivateData.put("submitBtnStatus", "disabled");
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
