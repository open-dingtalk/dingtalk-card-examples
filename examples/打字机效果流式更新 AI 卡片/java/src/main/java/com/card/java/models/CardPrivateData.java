package com.card.java.models;

import java.util.List;
import com.alibaba.fastjson.JSONObject;
import com.alibaba.fastjson.annotation.JSONField;

import lombok.Data;

@Data
public class CardPrivateData {
  @JSONField(name = "actionIds")
  private List<String> actionIds;
  @JSONField(name = "params")
  private JSONObject params;
}
