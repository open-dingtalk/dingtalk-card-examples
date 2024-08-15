package com.card.java.models;

import com.alibaba.fastjson.annotation.JSONField;
import lombok.Data;

@Data
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
}
