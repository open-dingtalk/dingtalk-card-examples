package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

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
	dingtalkClient *DingTalkClient          = nil
	contentById    map[string][]interface{} = make(map[string][]interface{})
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

func (c *DingTalkClient) GetUserInfoByUserId(userid string) (interface{}, error) {
	accessToken, err := c.GetAccessToken()
	if err != nil {
		return nil, err
	}

	data, _ := json.Marshal(map[string]string{"userid": userid})
	url := "https://oapi.dingtalk.com/topapi/v2/user/get"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		logger.GetLogger().Errorf("NewRequest Error: ", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	q := req.URL.Query()
	q.Add("access_token", accessToken)
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.GetLogger().Errorf("Http request Error: ", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.GetLogger().Errorf("Read body Error: ", err)
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		logger.GetLogger().Errorf("Unmarshal Error: ", err)
		return nil, err
	}

	if errcode, ok := result["errcode"].(float64); ok && errcode == 0 {
		return result["result"], nil
	}

	errmsg, ok := result["errmsg"].(string)
	if !ok {
		errmsg = "unknown error"
	}
	logger.GetLogger().Errorf("Get user info by userid failed, errcode: %v, errmsg: %s\n", result["errcode"], errmsg)
	return nil, fmt.Errorf("get user info by userid failed, errcode: %v, errmsg: %s", result["errcode"], errmsg)
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
	CARD_TEMPLATE_ID := "3d667b86-d30b-43ef-be8c-7fca37965210.schema" // 该模板只用于测试使用，如需投入线上使用，请导入卡片模板 json 到自己的应用下
	// 卡片公有数据，非字符串类型的卡片数据参考文档：https://open.dingtalk.com/document/orgapp/instructions-for-filling-in-api-card-data
	cardData := &dingtalkcard_1_0.CreateAndDeliverRequestCardData{
		CardParamMap: make(map[string]*string),
	}
	cardData.CardParamMap["title"] = tea.String(content)
	cardData.CardParamMap["joined"] = tea.String("false")
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
		logger.GetLogger().Infof("reply card failed: %+v", sendCardResponse)
		return nil, err
	}
	logger.GetLogger().Infof("reply card: %v %+v", outTrackId, sendCardRequest.CardData)

	return []byte(""), nil
}

func onCardCallback(ctx context.Context, request *card.CardRequest) (*card.CardResponse, error) {
	userId := request.UserId
	logger.GetLogger().Infof("card callback message: %v", request)

	cardInstanceId := request.OutTrackId
	userPrivateData := make(map[string]*string)
	userPrivateData["uid"] = tea.String(userId)

	currentContent, exists := contentById[cardInstanceId]
	if !exists {
		currentContent = []interface{}{}
	}
	nextContent := make([]interface{}, 0)

	params := request.CardActionData.CardPrivateData.Params
	deleteUid, exists := params["delete_uid"]
	if !exists {
		deleteUid = ""
	}

	if deleteUid != "" {
		// 取消接龙
		userPrivateData["joined"] = tea.String("false")
		for _, item := range currentContent {
			if item.(map[string]interface{})["uid"] == deleteUid {
				continue
			}
			nextContent = append(nextContent, item)
		}
	} else {
		// 参与接龙
		userPrivateData["joined"] = tea.String("true")
		body := map[string]interface{}{
			"timestamp": time.Now().Format("2006-01-02 15:04:05"),
			"uid":       userId,
			"remark":    params["remark"],
			"nick":      "",
			"avatar":    "",
		}
		userInfo, err := dingtalkClient.GetUserInfoByUserId(userId)
		if err == nil {
			logger.GetLogger().Infof("get user info: %+v", userInfo)
			userInfoMap, ok := userInfo.(map[string]interface{})
			if ok {
				if name, ok := userInfoMap["name"].(string); ok {
					body["nick"] = name
				}
				if avatar, ok := userInfoMap["avatar"].(string); ok {
					body["avatar"] = avatar
				}
			}
		}
		nextContent = append(currentContent, body)
	}
	contentById[cardInstanceId] = nextContent
	nextContentJSON, _ := json.Marshal(nextContent)
	nextContentStr := string(nextContentJSON)

	// 更新接龙列表和参与状态
	updateCardData := &dingtalkcard_1_0.UpdateCardRequestCardData{
		CardParamMap: make(map[string]*string),
	}
	updateCardData.CardParamMap["content"] = tea.String(nextContentStr)
	updatePrivateData := map[string]*dingtalkcard_1_0.PrivateDataValue{
		userId: {
			CardParamMap: userPrivateData,
		},
	}
	updateOptions := &dingtalkcard_1_0.UpdateCardRequestCardUpdateOptions{
		UpdateCardDataByKey:    tea.Bool(true),
		UpdatePrivateDataByKey: tea.Bool(true),
	}
	updateCardRequest := &dingtalkcard_1_0.UpdateCardRequest{
		OutTrackId:        tea.String(cardInstanceId),
		CardData:          updateCardData,
		PrivateData:       updatePrivateData,
		CardUpdateOptions: updateOptions,
	}
	updateCardResponse, err := dingtalkClient.UpdateCard(updateCardRequest)
	if err != nil {
		logger.GetLogger().Errorf("update card failed: %+v", updateCardResponse)
	} else {
		logger.GetLogger().Infof("update card: %v %+v %+v", cardInstanceId, updateCardRequest.CardData, updateCardRequest.PrivateData)
	}

	return nil, nil
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
