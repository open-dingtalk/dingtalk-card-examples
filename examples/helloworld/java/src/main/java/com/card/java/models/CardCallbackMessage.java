package com.card.java.models;

import com.alibaba.fastjson.annotation.JSONField;

public class CardCallbackMessage {
  @JSONField(name = "extension")
  private String extension;
  @JSONField(name = "corpId")
  private String corpId;
  @JSONField(name = "spaceType")
  private String spaceType;
  @JSONField(name = "userIdType")
  private Integer userIdType;
  @JSONField(name = "type")
  private String type;
  @JSONField(name = "userId")
  private String userId;
  @JSONField(name = "content")
  private String content; // This will be handled as a JSON string.
  @JSONField(name = "spaceId")
  private String spaceId;
  @JSONField(name = "outTrackId")
  private String outTrackId;
  @JSONField(name = "value")
  private String value; // This will be handled as a JSON string.

  public String getExtension() {
    return extension;
  }

  public void setExtension(String extension) {
    this.extension = extension;
  }

  public String getCorpId() {
    return corpId;
  }

  public void setCorpId(String corpId) {
    this.corpId = corpId;
  }

  public String getSpaceType() {
    return spaceType;
  }

  public void setSpaceType(String spaceType) {
    this.spaceType = spaceType;
  }

  public Integer getUserIdType() {
    return userIdType;
  }

  public void setUserIdType(Integer userIdType) {
    this.userIdType = userIdType;
  }

  public String getType() {
    return type;
  }

  public void setType(String type) {
    this.type = type;
  }

  public String getUserId() {
    return userId;
  }

  public void setUserId(String userId) {
    this.userId = userId;
  }

  public String getContent() {
    return content;
  }

  public void setContent(String content) {
    this.content = content;
  }

  public String getSpaceId() {
    return spaceId;
  }

  public void setSpaceId(String spaceId) {
    this.spaceId = spaceId;
  }

  public String getOutTrackId() {
    return outTrackId;
  }

  public void setOutTrackId(String outTrackId) {
    this.outTrackId = outTrackId;
  }

  public String getValue() {
    return value;
  }

  public void setValue(String value) {
    this.value = value;
  }
}
