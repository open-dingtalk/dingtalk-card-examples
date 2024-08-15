package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dingtalkcard_1_0 "github.com/alibabacloud-go/dingtalk/card_1_0"
	dingtalkim_1_0 "github.com/alibabacloud-go/dingtalk/im_1_0"
	dingtalkoauth2_1_0 "github.com/alibabacloud-go/dingtalk/oauth2_1_0"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/google/uuid"
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/card"
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/chatbot"
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/client"
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/logger"
)

type DingTalkClient struct {
	ClientID     string
	clientSecret string
	imClient     *dingtalkim_1_0.Client
	oauthClient  *dingtalkoauth2_1_0.Client
	cardClient   *dingtalkcard_1_0.Client
}

var (
	dingtalkClient *DingTalkClient = nil
)

func NewDingTalkClient(clientId, clientSecret string) *DingTalkClient {
	config := &openapi.Config{}
	config.Protocol = tea.String("https")
	config.RegionId = tea.String("central")
	oauthClient, _ := dingtalkoauth2_1_0.NewClient(config)
	imClient, _ := dingtalkim_1_0.NewClient(config)
	cardClient, _ := dingtalkcard_1_0.NewClient(config)
	return &DingTalkClient{
		ClientID:     clientId,
		clientSecret: clientSecret,
		oauthClient:  oauthClient,
		imClient:     imClient,
		cardClient:   cardClient,
	}
}

func (c *DingTalkClient) GetAccessToken() (string, error) {
	request := &dingtalkoauth2_1_0.GetAccessTokenRequest{
		AppKey:    tea.String(c.ClientID),
		AppSecret: tea.String(c.clientSecret),
	}
	response, tryErr := func() (_resp *dingtalkoauth2_1_0.GetAccessTokenResponse, _e error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				_e = r
			}
		}()
		_resp, _err := c.oauthClient.GetAccessToken(request)
		if _err != nil {
			return nil, _err
		}

		return _resp, nil
	}()
	if tryErr != nil {
		return "", tryErr
	}
	return *response.Body.AccessToken, nil
}

func (c *DingTalkClient) SendCard(request *dingtalkcard_1_0.CreateAndDeliverRequest) (*dingtalkcard_1_0.CreateAndDeliverResponse, error) {
	accessToken, err := c.GetAccessToken()
	if err != nil {
		return nil, err
	}
	headers := &dingtalkcard_1_0.CreateAndDeliverHeaders{}
	headers.XAcsDingtalkAccessToken = tea.String(accessToken)

	resp, tryErr := func() (resp *dingtalkcard_1_0.CreateAndDeliverResponse, _e error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				_e = r
			}
		}()
		result, _err := c.cardClient.CreateAndDeliverWithOptions(request, headers, &util.RuntimeOptions{})
		if _err != nil {
			return nil, _err
		}

		return result, nil
	}()
	if tryErr != nil {
		var sdkError = &tea.SDKError{}
		if _t, ok := tryErr.(*tea.SDKError); ok {
			sdkError = _t
		} else {
			sdkError.Message = tea.String(tryErr.Error())
		}
		if !tea.BoolValue(util.Empty(sdkError.Code)) && !tea.BoolValue(util.Empty(sdkError.Message)) {
			logger.GetLogger().Errorf("CreateAndDeliverWithOptions failed, clientId=%s, err=%+v", c.ClientID, sdkError)
		}
		return nil, tryErr
	}

	return resp, nil
}

func (c *DingTalkClient) UpdateCard(request *dingtalkcard_1_0.UpdateCardRequest) (*dingtalkcard_1_0.UpdateCardResponse, error) {
	accessToken, err := c.GetAccessToken()
	if err != nil {
		return nil, err
	}
	headers := &dingtalkcard_1_0.UpdateCardHeaders{}
	headers.XAcsDingtalkAccessToken = tea.String(accessToken)

	resp, tryErr := func() (resp *dingtalkcard_1_0.UpdateCardResponse, _e error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				_e = r
			}
		}()
		result, _err := c.cardClient.UpdateCardWithOptions(request, headers, &util.RuntimeOptions{})
		if _err != nil {
			return nil, _err
		}

		return result, nil
	}()
	if tryErr != nil {
		var sdkError = &tea.SDKError{}
		if _t, ok := tryErr.(*tea.SDKError); ok {
			sdkError = _t
		} else {
			sdkError.Message = tea.String(tryErr.Error())
		}
		if !tea.BoolValue(util.Empty(sdkError.Code)) && !tea.BoolValue(util.Empty(sdkError.Message)) {
			logger.GetLogger().Errorf("UpdateCardWithOptions failed, clientId=%s, err=%+v", c.ClientID, sdkError)
		}
		return nil, tryErr
	}

	return resp, nil
}

func OnChatBotMessageReceived(ctx context.Context, data *chatbot.BotCallbackDataModel) ([]byte, error) {
	content := strings.TrimSpace(data.Text.Content)
	logger.GetLogger().Infof("received message: %v", content)

	// 卡片模板 ID
	CARD_TEMPLATE_ID := "fcc1df51-17bb-403f-aca9-65f1c6919129.schema" // 该模板只用于测试使用，如需投入线上使用，请导入卡片模板 json 到自己的应用下
	// 卡片公有数据，非字符串类型的卡片数据参考文档：https://open.dingtalk.com/document/orgapp/instructions-for-filling-in-api-card-data
	cardData := &dingtalkcard_1_0.CreateAndDeliverRequestCardData{
		CardParamMap: make(map[string]*string),
	}
	cardData.CardParamMap["last_message"] = tea.String("事件链演示")
	cardData.CardParamMap["markdown"] = tea.String("<font colorTokenV2=common_green1_color>动态显示的 markdown 内容</font>")
	imGroupOpenSpaceModel := &dingtalkcard_1_0.CreateAndDeliverRequestImGroupOpenSpaceModel{
		SupportForward: tea.Bool(true),
	}
	imGroupOpenDeliverModel := &dingtalkcard_1_0.CreateAndDeliverRequestImGroupOpenDeliverModel{
		Extension: make(map[string]*string),
		RobotCode: tea.String(dingtalkClient.ClientID),
	}
	imRobotOpenSpaceModel := &dingtalkcard_1_0.CreateAndDeliverRequestImRobotOpenSpaceModel{
		SupportForward: tea.Bool(true),
	}
	imRobotOpenDeliverModel := &dingtalkcard_1_0.CreateAndDeliverRequestImRobotOpenDeliverModel{
		Extension: make(map[string]*string),
		RobotCode: tea.String(dingtalkClient.ClientID),
		SpaceType: tea.String("IM_ROBOT"),
	}
	u, _ := uuid.NewUUID()
	outTrackId := u.String()

	sendCardRequest := &dingtalkcard_1_0.CreateAndDeliverRequest{
		UserIdType:     tea.Int32(1), // 1（默认）：userid模式；2：unionId模式;
		CardTemplateId: tea.String(CARD_TEMPLATE_ID),
		OutTrackId:     tea.String(outTrackId),
		CallbackType:   tea.String("STREAM"), // 采用 Stream 模式接收回调事件
		CardData:       cardData,
	}
	if data.ConversationType == "2" {
		// 群聊
		sendCardRequest.OpenSpaceId = tea.String(fmt.Sprintf("dtv1.card//IM_GROUP.%s", data.ConversationId))
		sendCardRequest.ImGroupOpenSpaceModel = imGroupOpenSpaceModel
		sendCardRequest.ImGroupOpenDeliverModel = imGroupOpenDeliverModel
	} else {
		// 单聊
		sendCardRequest.OpenSpaceId = tea.String(fmt.Sprintf("dtv1.card//IM_ROBOT.%s", data.SenderStaffId))
		sendCardRequest.ImRobotOpenSpaceModel = imRobotOpenSpaceModel
		sendCardRequest.ImRobotOpenDeliverModel = imRobotOpenDeliverModel
	}

	// 创建并投放卡片
	sendCardResponse, err := dingtalkClient.SendCard(sendCardRequest)
	if err != nil {
		logger.GetLogger().Errorf("reply card failed: %+v", sendCardResponse)
		return nil, err
	}
	logger.GetLogger().Infof("reply card: %v %+v", outTrackId, sendCardRequest.CardData)

	return []byte(""), nil
}

func onCardCallback(ctx context.Context, request *card.CardRequest) (*card.CardResponse, error) {
	/**
	 * 卡片事件回调文档：https://open.dingtalk.com/document/orgapp/event-callback-card
	 */
	logger.GetLogger().Infof("card callback message: %v", request)

	updateCardData := make(map[string]string)
	userPrivateData := make(map[string]string)

	params := request.CardActionData.CardPrivateData.Params
	if variable, ok := params["var"]; ok && variable != nil {
		if variable == "pub_url" {
			updateCardData["pub_url"] = ""
			updateCardData["pub_url_msg"] = ""
			status := rand.Intn(2)
			if status == 0 {
				randomNumber := rand.Intn(101)
				updateCardData["pub_url_status"] = "failed"
				updateCardData["pub_url_msg"] = fmt.Sprintf("更新失败%d", randomNumber)
			} else {
				updateCardData["pub_url_status"] = "success"
				updateCardData["pub_url"] = "dingtalk://dingtalkclient/page/link?web_wnd=workbench&pc_slide=true&hide_bar=true&url=https://github.com/open-dingtalk/dingtalk-card-examples"
			}
		} else if variable == "pri_url" {
			userPrivateData["pri_url"] = ""
			userPrivateData["pri_url_msg"] = ""
			status := rand.Intn(2)
			if status == 0 {
				randomNumber := rand.Intn(101)
				userPrivateData["pri_url_status"] = "failed"
				userPrivateData["pri_url_msg"] = fmt.Sprintf("更新失败%d", randomNumber)
			} else {
				userPrivateData["pri_url_status"] = "success"
				userPrivateData["pri_url"] = "dingtalk://dingtalkclient/page/link?web_wnd=workbench&pc_slide=true&hide_bar=true&url=https://github.com/open-dingtalk/dingtalk-card-examples"
			}
		}
	}

	response := &card.CardResponse{
		CardUpdateOptions: &card.CardUpdateOptions{
			UpdateCardDataByKey:    true,
			UpdatePrivateDataByKey: true,
		},
		CardData: &card.CardDataDto{
			CardParamMap: updateCardData,
		},
		UserPrivateData: &card.CardDataDto{
			CardParamMap: userPrivateData,
		},
	}

	responseJSON, err := json.MarshalIndent(response, "", "    ") // 使用 MarshalIndent 来美化输出
	if err == nil {
		logger.GetLogger().Infof("card callback response: \n%s", string(responseJSON))
	}
	return response, nil
}

func main() {
	var clientId, clientSecret string
	flag.StringVar(&clientId, "client_id", os.Getenv("DINGTALK_APP_CLIENT_ID"), "your-client-id")
	flag.StringVar(&clientSecret, "client_secret", os.Getenv("DINGTALK_APP_CLIENT_SECRET"), "your-client-secret")
	flag.Parse()
	if len(clientId) == 0 || len(clientSecret) == 0 {
		panic("command line options --client_id and --client_secret required")
	}

	logger.SetLogger(logger.NewStdTestLogger())

	dingtalkClient = NewDingTalkClient(clientId, clientSecret)

	cli := client.NewStreamClient(client.WithAppCredential(client.NewAppCredentialConfig(clientId, clientSecret)))
	cli.RegisterChatBotCallbackRouter(OnChatBotMessageReceived)
	cli.RegisterCardCallbackRouter(onCardCallback)

	err := cli.Start(context.Background())
	if err != nil {
		panic(err)
	}

	defer cli.Close()

	select {}
}
