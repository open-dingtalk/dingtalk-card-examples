package com.card.java.models;

import com.alibaba.fastjson.annotation.JSONField;
import lombok.Data;

@Data
public class CardPrivateDataWrapper {
  @JSONField(name = "cardPrivateData")
  private CardPrivateData cardPrivateData;
}
