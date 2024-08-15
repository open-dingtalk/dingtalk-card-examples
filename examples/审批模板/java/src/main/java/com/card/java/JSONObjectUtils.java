package com.card.java;

import com.alibaba.fastjson.JSON;
import com.alibaba.fastjson.JSONObject;
import org.springframework.stereotype.Component;

@Component
public class JSONObjectUtils {
  public JSONObject convertJSONValuesToString(JSONObject obj) {
    JSONObject result = new JSONObject();
    for (String key : obj.keySet()) {
      Object value = obj.get(key);
      if (value instanceof String) {
        result.put(key, value);
      } else {
        // 将非字符串类型的值转换成JSON字符串
        result.put(key, JSON.toJSONString(value));
      }
    }
    return result;
  }
}
