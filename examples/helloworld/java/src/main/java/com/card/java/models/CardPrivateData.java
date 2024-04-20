package com.card.java.models;

import java.util.List;
import com.alibaba.fastjson.JSONObject;
import com.alibaba.fastjson.annotation.JSONField;

public class CardPrivateData {
  @JSONField(name = "actionIds")
  private List<String> actionIds;
  @JSONField(name = "params")
  private JSONObject params;

  public List<String> getActionIds() {
    return actionIds;
  }

  public void setActionIds(List<String> actionIds) {
    this.actionIds = actionIds;
  }

  public JSONObject getParams() {
    return params;
  }

  public void setParams(JSONObject params) {
    this.params = params;
  }
}
