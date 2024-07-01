package com.card.java;

import lombok.extern.slf4j.Slf4j;
import com.alibaba.fastjson.JSON;
import com.alibaba.fastjson.JSONArray;
import com.alibaba.fastjson.JSONObject;
import com.card.java.models.CardCallbackMessage;
import com.card.java.models.CardPrivateData;
import com.card.java.models.CardPrivateDataWrapper;

import java.lang.reflect.Array;
import java.lang.reflect.Field;
import java.util.function.Predicate;
import java.util.stream.Collectors;
import java.util.ArrayList;
import java.util.Collection;
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

  private static Map<String, String> valueKeyMap = new HashMap<>();

  static {
    valueKeyMap.put("TEXT", "default_string");
    valueKeyMap.put("DATE", "default_string");
    valueKeyMap.put("DATETIME", "default_string");
    valueKeyMap.put("SELECT", "default_number");
    valueKeyMap.put("MULTI_SELECT", "default_number_array");
    valueKeyMap.put("CHECKBOX", "default_boolean");
    valueKeyMap.put("CHECKBOX_LIST", "checkbox_items");
    valueKeyMap.put("CHECKBOX_LIST_MULTI", "checkbox_items");
  }

  @Override
  public JSONObject execute(String messageString) {
    /**
     * 卡片事件回调文档：https://open.dingtalk.com/document/orgapp/event-callback-card
     */
    log.info("card callback message: " + messageString);
    JSONObject userPrivateData = new JSONObject();

    CardCallbackMessage message = JSON.parseObject(messageString, CardCallbackMessage.class);
    String outTrackId = message.getOutTrackId();
    CardPrivateDataWrapper content = JSON.parseObject(message.getContent(), CardPrivateDataWrapper.class);
    CardPrivateData cardPrivateData = content.getCardPrivateData();
    JSONObject params = cardPrivateData.getParams();

    JSONArray submitFormFields = (JSONArray) params.get("submit_form_fields");

    if (submitFormFields != null && !submitFormFields.isEmpty()) {
      // 提交表单，做必填校验，响应错误提示或者响应提交成功处理
      List<String> requiredErrorLabels = new ArrayList<>();
      for (int i = 0; i < submitFormFields.size(); i++) {
        JSONObject formField = submitFormFields.getJSONObject(i);
        String formFieldType = formField.getString("type");
        if (!valueKeyMap.containsKey(formFieldType)) {
          userPrivateData.put("err_msg", "无效的表单类型「" + formFieldType + "」");
          break;
        }
        String formFieldLabel = formField.getString("label");
        Boolean formFieldRequired = formField.getBooleanValue("required");
        Object formFieldValue = formField.get(valueKeyMap.get(formFieldType));

        if (formFieldType.equals("CHECKBOX_LIST") || formFieldType.equals("CHECKBOX_LIST_MULTI")) {
          Predicate<JSONObject> checkedFilter = (JSONObject x) -> x.getBooleanValue("checked");
          long checkedCount = formField.getJSONArray(valueKeyMap.get(formFieldType)).stream()
              .map(JSONObject.class::cast).filter(checkedFilter).count();
          if (formFieldRequired && checkedCount == 0) {
            requiredErrorLabels.add(formFieldLabel);
          }
        } else if (formFieldType.equals("SELECT")) {
          if (formFieldRequired && !(formFieldValue instanceof Integer && (Integer) formFieldValue >= 0)) {
            requiredErrorLabels.add(formFieldLabel);
          }
        } else {
          try {
            if (formFieldRequired && isEmpty(formFieldValue)) {
              requiredErrorLabels.add(formFieldLabel);
            }
          } catch (IllegalAccessException e) {
            e.printStackTrace();
          }
        }
      }

      if (!userPrivateData.containsKey("err_msg")) {
        if (requiredErrorLabels.isEmpty()) {
          userPrivateData.put("form_status", "disabled");
          userPrivateData.put("button_text", "已提交");
        } else {
          userPrivateData.put("err_msg", "请填写必填项「" + String.join(", ", requiredErrorLabels) + "」");
        }
      }
    } else {
      String updateName = params.getString("name");
      if (updateName == null) {
        userPrivateData.put("err_msg", "服务异常");
      } else {
        log.info("cardInstanceId: " + outTrackId);
        JSONArray formFields = ChatBotHandler.formFieldsByInstanceId.get(outTrackId);
        for (int i = 0; i < formFields.size(); i++) {
          JSONObject formField = formFields.getJSONObject(i);
          if (formField.getString("name").equals(updateName)) {
            Boolean remove = false;
            try {
              remove = !isEmpty(params.getString("remove"));
            } catch (IllegalAccessException e) {
              e.printStackTrace();
            }

            String updateType = params.getString("type");
            String updateKey = valueKeyMap.get(updateType);
            Object updateValue = "";
            if (remove && updateType.equals("multiSelect")) {
              updateType = "MULTI_SELECT";
              updateKey = valueKeyMap.get(updateType);
              int removeIndex = params.getJSONObject(updateName).getIntValue("index");
              updateValue = formField.getJSONArray(updateKey).toJavaList(Integer.class).stream()
                  .filter(v -> v != removeIndex).collect(Collectors.toList());
            } else if (updateType.equals("CHECKBOX_LIST")) {
              updateValue = formField.getJSONArray(updateKey).stream()
                  .map(obj -> ((JSONObject) obj).fluentPut("checked",
                      ((JSONObject) obj).getIntValue("value") == params.getIntValue("value")))
                  .collect(Collectors.toCollection(JSONArray::new));
            } else if (updateType.equals("CHECKBOX_LIST_MULTI")) {
              updateValue = formField.getJSONArray(updateKey).stream().map(obj -> {
                JSONObject json = (JSONObject) obj;
                boolean checked = json.getIntValue("value") == params.getIntValue("value")
                    ? !json.getBooleanValue("checked")
                    : json.getBooleanValue("checked");
                return json.fluentPut("checked", checked);
              }).collect(Collectors.toCollection(JSONArray::new));
            } else {
              updateValue = params.get(updateName);
              if (updateType.equals("SELECT")) {
                updateValue = ((JSONObject) updateValue).getIntValue("index");
              } else if (updateType.equals("MULTI_SELECT")) {
                updateValue = ((JSONObject) updateValue).getJSONArray("index");
              } else if (updateType.equals("CHECKBOX")) {
                updateValue = !formField.getBooleanValue(updateKey);
              }
            }
            formField.put(updateKey, updateValue);
          }
        }
        userPrivateData.put("form_fields", formFields);
      }
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

  public static boolean isEmpty(Object value) throws IllegalAccessException {
    if (value == null) {
      return true;
    }

    if (value instanceof String) {
      return ((String) value).isEmpty();
    }

    if (value.getClass().isArray()) {
      return Array.getLength(value) == 0;
    }

    if (value instanceof Collection) {
      return ((Collection<?>) value).isEmpty();
    }

    if (value instanceof Map) {
      return ((Map<?, ?>) value).isEmpty();
    }

    if (value instanceof Boolean) {
      return !((Boolean) value);
    }

    if (value instanceof Number) {
      if (value instanceof Float || value instanceof Double) {
        return ((Number) value).doubleValue() == 0;
      }
      return ((Number) value).longValue() == 0;
    }

    for (Field field : value.getClass().getDeclaredFields()) {
      field.setAccessible(true);
      Object fieldValue = field.get(value);
      if (!isEmpty(fieldValue)) {
        return false;
      }
    }
    return true;
  }
}
