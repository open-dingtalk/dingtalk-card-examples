package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dingtalkcard_1_0 "github.com/alibabacloud-go/dingtalk/card_1_0"
	dingtalkim_1_0 "github.com/alibabacloud-go/dingtalk/im_1_0"
	dingtalkoauth2_1_0 "github.com/alibabacloud-go/dingtalk/oauth2_1_0"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/google/uuid"
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
	CARD_TEMPLATE_ID := "b23d3b9d-1c9c-4a3b-82a8-744d475c483d.schema" // 该模板只用于测试使用，如需投入线上使用，请导入卡片模板 json 到自己的应用下
	// 卡片公有数据，非字符串类型的卡片数据参考文档：https://open.dingtalk.com/document/orgapp/instructions-for-filling-in-api-card-data
	cardData := &dingtalkcard_1_0.CreateAndDeliverRequestCardData{
		CardParamMap: make(map[string]*string),
	}
	cardData.CardParamMap["evaluate_done"] = tea.String("false")
	table := map[string]interface{}{
		"meta": []map[string]interface{}{
			{
				"aliasName": "",
				"dataType":  "STRING",
				"alias":     "rank",
				"weight":    10,
			},
			{
				"aliasName": "应用名",
				"dataType":  "MICROAPP",
				"alias":     "appItem",
				"weight":    40,
			},
			{
				"aliasName": "点击次数",
				"dataType":  "STRING",
				"alias":     "pv",
				"weight":    25,
			},
			{
				"aliasName": "点击人数",
				"dataType":  "STRING",
				"alias":     "uv",
				"weight":    25,
			},
		},
		"data": []map[string]interface{}{
			{
				"uv":   "324",
				"pv":   "433",
				"rank": 1,
				"appItem": map[string]string{
					"icon": "https://static.dingtalk.com/media/lALPDeC2uGvNwy3NArzNArw_700_700.png",
					"name": "考勤打卡",
				},
			},
			{
				"uv":   "350",
				"pv":   "354",
				"rank": 2,
				"appItem": map[string]string{
					"icon": "https://static.dingtalk.com/media/lALPDeC2uGvNwy3NArzNArw_700_700.png",
					"name": "智能人事",
				},
			},
			{
				"uv":   "189",
				"pv":   "322",
				"rank": 3,
				"appItem": map[string]string{
					"icon": "https://static.dingtalk.com/media/lALPDeC2uGvNwy3NArzNArw_700_700.png",
					"name": "日志",
				},
			},
		},
	}
	tableString, _ := json.Marshal(table)
	cardData.CardParamMap["table"] = tea.String(string(tableString))
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

	// 创建并投放卡片: https://open.dingtalk.com/document/isvapp/create-and-deliver-cards
	sendCardResponse, err := dingtalkClient.SendCard(sendCardRequest)
	if err != nil {
		logger.GetLogger().Errorf("reply card failed: %+v", sendCardResponse)
		return nil, err
	}
	logger.GetLogger().Infof("reply card: %v %+v", outTrackId, sendCardRequest.CardData)

	updateCardData := &dingtalkcard_1_0.UpdateCardRequestCardData{
		CardParamMap: make(map[string]*string),
	}
	updateCardData.CardParamMap["more_detail_url"] = tea.String(fmt.Sprintf("dingtalk://dingtalkclient/page/link?pc_slide=true&url=%s", url.QueryEscape(fmt.Sprintf("http://localhost:3000?page=detail&id=%s", outTrackId))))
	updateCardData.CardParamMap["evaluate_url"] = tea.String(fmt.Sprintf("dingtalk://dingtalkclient/page/link?pc_slide=true&url=%s", url.QueryEscape(fmt.Sprintf("http://localhost:3000?page=evaluate&id=%s", outTrackId))))
	updateOptions := &dingtalkcard_1_0.UpdateCardRequestCardUpdateOptions{
		UpdateCardDataByKey:    tea.Bool(true),
		UpdatePrivateDataByKey: tea.Bool(true),
	}
	updateCardRequest := &dingtalkcard_1_0.UpdateCardRequest{
		OutTrackId:        tea.String(outTrackId),
		CardData:          updateCardData,
		CardUpdateOptions: updateOptions,
	}
	// 更新卡片: https://open.dingtalk.com/document/orgapp/interactive-card-update-interface
	updateCardResponse, err := dingtalkClient.UpdateCard(updateCardRequest)
	if err != nil {
		logger.GetLogger().Errorf("update card failed: %+v", updateCardResponse)
		return nil, err
	}
	logger.GetLogger().Infof("update card: %v %+v", outTrackId, updateCardRequest.CardData)

	return []byte(""), nil
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

	err := cli.Start(context.Background())
	if err != nil {
		panic(err)
	}

	defer cli.Close()

	select {}
}
