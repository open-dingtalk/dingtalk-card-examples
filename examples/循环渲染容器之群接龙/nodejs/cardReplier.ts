import { DWClient } from "dingtalk-stream";
import axios from "axios";
import { v1 as uuidv1 } from "uuid";
import * as crypto from "crypto";

const DINGTALK_OPENAPI_ENDPOINT: string = "https://api.dingtalk.com";

interface IncomingMessage {
  senderPlatform: string;
  conversationId: string;
  chatbotCorpId: string;
  chatbotUserId: string;
  msgId: string;
  senderNick: string;
  isAdmin: boolean;
  senderStaffId: string;
  sessionWebhookExpiredTime: number;
  createAt: number;
  senderCorpId: string;
  conversationType: string;
  senderId: string;
  sessionWebhook: string;
  text: {
    content: string;
  };
  robotCode: string;
  msgtype: string;
}

interface CreateAndDeliverCardOptions {
  cardTemplateId: string;
  cardData: Record<string, string>;
  callbackType?: string;
  atSender?: boolean;
  atAll?: boolean;
  recipients?: string[];
  supportForward?: boolean;
  [key: string]: any;
}

interface PutCardDataOptions {
  cardInstanceId: string;
  cardData: Record<string, string>;
  [key: string]: any;
}

class CardReplier {
  dingtalkClient: DWClient;
  incomingMessage: IncomingMessage;

  constructor(dingtalkClient: DWClient, incomingMessage: IncomingMessage) {
    this.dingtalkClient = dingtalkClient;
    this.incomingMessage = incomingMessage;
  }

  static genCardId(msg: IncomingMessage): string {
    const factor: string = `${msg.senderId}_${msg.senderCorpId}_${
      msg.conversationId
    }_${msg.msgId}_${uuidv1()}`;
    const hash = crypto.createHash("sha256");
    hash.update(factor);
    return hash.digest("hex");
  }

  static getRequestHeader(accessToken: string): object {
    return {
      "Content-Type": "application/json",
      Accept: "*/*",
      "x-acs-dingtalk-access-token": accessToken,
    };
  }

  async createAndDeliverCard(
    options: CreateAndDeliverCardOptions
  ): Promise<string> {
    /**
     * 创建并投放卡片
     * https://open.dingtalk.com/document/orgapp/create-and-deliver-cards
     */
    const accessToken = await this.dingtalkClient.getAccessToken();
    if (!accessToken) {
      console.error(
        "CardResponder.sendCard failed, cannot get dingtalk access token"
      );
      return "";
    }

    const {
      cardTemplateId,
      cardData,
      callbackType = "STREAM",
      atSender = false,
      atAll = false,
      recipients = null,
      supportForward = true,
      ...restOptions
    } = options;

    const cardInstanceId = CardReplier.genCardId(this.incomingMessage);

    let data: any = {
      cardTemplateId,
      outTrackId: cardInstanceId,
      callbackType,
      cardData: {
        cardParamMap: cardData,
      },
      imGroupOpenSpaceModel: {
        supportForward,
      },
      imRobotOpenSpaceModel: {
        supportForward,
      },
    };

    // 2：群聊, 1：单聊
    if (this.incomingMessage.conversationType === "2") {
      data.openSpaceId = `dtv1.card//IM_GROUP.${this.incomingMessage.conversationId}`;
      data.imGroupOpenDeliverModel = {
        robotCode: this.dingtalkClient.config.clientId,
      };
      if (atAll) {
        data.imGroupOpenDeliverModel.atUserIds = { "@ALL": "@ALL" };
      } else if (atSender) {
        data.imGroupOpenDeliverModel.atUserIds = {
          [this.incomingMessage.senderStaffId]: this.incomingMessage.senderNick,
        };
      }
      if (Array.isArray(recipients)) {
        data.imGroupOpenDeliverModel.recipients = recipients;
      }
    } else if (this.incomingMessage.conversationType === "1") {
      data.openSpaceId = `dtv1.card//IM_ROBOT.${this.incomingMessage.senderStaffId}`;
      data.imRobotOpenDeliverModel = {
        spaceType: "IM_ROBOT",
      };
    }

    const url =
      DINGTALK_OPENAPI_ENDPOINT + "/v1.0/card/instances/createAndDeliver";
    try {
      const response = await axios.post(
        url,
        { ...data, ...restOptions },
        {
          headers: CardReplier.getRequestHeader(accessToken),
        }
      );
      if (response.data.success) {
        console.log(response.data);
        return response.data.result.outTrackId;
      } else {
        console.error("create and deliver card failed: ", response.data);
      }
    } catch (error) {
      console.error("craete and deliver card error: ", error);
      return "";
    }

    return "";
  }

  async putCardData(options: PutCardDataOptions): Promise<void> {
    /**
     * 更新卡片内容
     * https://open.dingtalk.com/document/orgapp/interactive-card-update-interface
     */
    const accessToken = await this.dingtalkClient.getAccessToken();
    if (!accessToken) {
      console.error(
        "CardResponder.sendCard failed, cannot get dingtalk access token"
      );
      return;
    }

    const { cardInstanceId, cardData, ...restOptions } = options;
    const data: any = {
      outTrackId: cardInstanceId,
      cardData: {
        cardParamMap: cardData,
      },
      ...restOptions,
    };
    const url = DINGTALK_OPENAPI_ENDPOINT + "/v1.0/card/instances";
    try {
      const response = await axios.put(url, data, {
        headers: CardReplier.getRequestHeader(accessToken),
      });
      if (response.data.success) {
        console.log(response.data);
        return;
      } else {
        console.error("update card failed: ", response.data);
      }
    } catch (error) {
      console.error("update card error: ", error);
    }
    return;
  }

  async getUserInfoByUserId(userid: string): Promise<any> {
    /**
     * 查询用户详情
     * https://open.dingtalk.com/document/isvapp/query-user-details
     */
    const accessToken = await this.dingtalkClient.getAccessToken();
    if (!accessToken) {
      console.error(
        "CardResponder.sendCard failed, cannot get dingtalk access token"
      );
      return;
    }
    const data = { userid };
    const url = "https://oapi.dingtalk.com/topapi/v2/user/get";
    try {
      const response = await axios.post(url, data, {
        params: { access_token: accessToken },
      });
      if (response.data.errcode === 0) {
        return response.data.result;
      }
      console.error(
        `get user info by userid failed, errcode: ${response.data.errcode}, errmsg: ${response.data.errmsg}`
      );
    } catch (error) {
      console.error("get user info by userid error: ", error);
      return;
    }
  }
}

export default CardReplier;
