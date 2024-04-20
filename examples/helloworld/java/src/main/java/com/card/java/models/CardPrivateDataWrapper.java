package com.card.java.models;

import com.alibaba.fastjson.annotation.JSONField;

public class CardPrivateDataWrapper {
  @JSONField(name = "cardPrivateData")
  private CardPrivateData cardPrivateData;

  public CardPrivateData getCardPrivateData() {
    return cardPrivateData;
  }

  public void setCardPrivateData(CardPrivateData cardPrivateData) {
    this.cardPrivateData = cardPrivateData;
  }
}
