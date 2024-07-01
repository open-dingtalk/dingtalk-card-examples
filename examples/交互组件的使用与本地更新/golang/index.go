package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"reflect"
	"strconv"
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

type UserPrivateData map[string]interface{}

type CheckboxItem struct {
	Value int    `json:"value"`
	Text  string `json:"text"`
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

func convertJsonValuesToString(obj map[string]interface{}) map[string]string {
	result := make(map[string]string)
	for key, value := range obj {
		switch v := value.(type) {
		case string:
			result[key] = v
		default:
			encodedValue, _ := json.Marshal(value)
			result[key] = string(encodedValue)
		}
	}
	return result
}

func convertJsonValuesToTeaString(obj map[string]interface{}) map[string]*string {
	result := make(map[string]*string)
	for key, value := range obj {
		switch v := value.(type) {
		case string:
			result[key] = tea.String(v)
		default:
			encodedValue, _ := json.Marshal(value)
			result[key] = tea.String(string(encodedValue))
		}
	}
	return result
}

func OnChatBotMessageReceived(ctx context.Context, data *chatbot.BotCallbackDataModel) ([]byte, error) {
	content := strings.TrimSpace(data.Text.Content)
	logger.GetLogger().Infof("received message: %v", content)

	// 卡片模板 ID
	CARD_TEMPLATE_ID := "737cda86-7a7f-4d83-ba07-321e6933be12.schema" // 该模板只用于测试使用，如需投入线上使用，请导入卡片模板 json 到自己的应用下
	// 卡片公有数据，非字符串类型的卡片数据参考文档：https://open.dingtalk.com/document/orgapp/instructions-for-filling-in-api-card-data
	cardParamMap := map[string]interface{}{
		"lastMessage":        "交互组件本地更新卡片",
		"submitBtnStatus":    "normal",
		"submitBtnText":      "提交",
		"input":              "",
		"selectIndex":        -1,
		"multiSelectIndexes": []int{},
		"date":               "",
		"datetime":           "",
		"checkbox":           false,
		"singleCheckboxItems": []CheckboxItem{
			{Value: 0, Text: "单选复选框选项 1"},
			{Value: 1, Text: "单选复选框选项 2"},
			{Value: 2, Text: "单选复选框选项 3"},
			{Value: 3, Text: "单选复选框选项 4"},
		},
		"multiCheckboxItems": []CheckboxItem{
			{Value: 0, Text: "多选复选框选项 1"},
			{Value: 1, Text: "多选复选框选项 2"},
			{Value: 2, Text: "多选复选框选项 3"},
			{Value: 3, Text: "多选复选框选项 4"},
		},
	}
	cardData := &dingtalkcard_1_0.CreateAndDeliverRequestCardData{
		CardParamMap: convertJsonValuesToTeaString(cardParamMap),
	}

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

func checkRequiredFields(params map[string]interface{}) (map[string]string, error) {
	userPrivateData := make(UserPrivateData)
	requiredFields := map[string]string{
		"input":          "文本输入",
		"select":         "下拉单选",
		"multiSelect":    "下拉多选",
		"date":           "日期选择",
		"datetime":       "日期时间选择",
		"singleCheckbox": "单选列表",
		"multiCheckbox":  "多选列表",
	}

	logger.GetLogger().Infof("input: %s", reflect.TypeOf((params["input"])))
	if input, ok := params["input"].(string); ok && input != "" {
		userPrivateData["input"] = input
		delete(requiredFields, "input")
	}

	logger.GetLogger().Infof("dateType: %s", reflect.TypeOf((params["date"])))
	if date, ok := params["date"].(string); ok && date != "" {
		userPrivateData["date"] = date
		delete(requiredFields, "date")
	}

	logger.GetLogger().Infof("datetimeType: %s", reflect.TypeOf((params["datetime"])))
	if datetime, ok := params["datetime"].(string); ok && datetime != "" {
		userPrivateData["datetime"] = datetime
		delete(requiredFields, "datetime")
	}

	logger.GetLogger().Infof("selectType: %s", reflect.TypeOf((params["select"])))
	if sel, ok := params["select"].(map[string]interface{}); ok {
		if index, exists := sel["index"]; exists {
			userPrivateData["selectIndex"] = index
			delete(requiredFields, "select")
		}
	}

	logger.GetLogger().Infof("multiSelectType: %s", reflect.TypeOf((params["multiSelect"])))
	if msel, ok := params["multiSelect"].(map[string]interface{}); ok {
		if indexRaw, indexExists := msel["index"].([]interface{}); indexExists && len(indexRaw) > 0 {
			var multiSelectIndexes []int
			for _, v := range indexRaw {
				if iv, ok := v.(float64); ok {
					multiSelectIndexes = append(multiSelectIndexes, int(iv))
				}
			}
			if len(multiSelectIndexes) > 0 {
				userPrivateData["multiSelectIndexes"] = multiSelectIndexes
				delete(requiredFields, "multiSelect")
			}
		}
	}

	logger.GetLogger().Infof("checkboxType: %s", reflect.TypeOf((params["checkbox"])))
	if checkbox, ok := params["checkbox"].(bool); ok {
		userPrivateData["checkbox"] = checkbox
	}

	logger.GetLogger().Infof("singleCheckboxType: %s", reflect.TypeOf((params["singleCheckbox"])))
	logger.GetLogger().Infof("singleCheckboxItemsType: %s", reflect.TypeOf((params["singleCheckboxItems"])))
	singleCheckboxValue := -1
	if singleCheckbox, ok := params["singleCheckbox"].(float64); ok {
		singleCheckboxValue = int(singleCheckbox)
	} else if singleCheckboxStr, ok := params["singleCheckbox"].(string); ok && singleCheckboxStr != "" {
		singleCheckboxValue, _ = strconv.Atoi(singleCheckboxStr)
	}
	if singleCheckboxValue >= 0 {
		if itemsRaw, ok := params["singleCheckboxItems"].([]interface{}); ok {
			var updatedItems []map[string]interface{}
			for _, itemRaw := range itemsRaw {
				if item, ok := itemRaw.(map[string]interface{}); ok {
					item["checked"] = singleCheckboxValue == int(item["value"].(float64))
					updatedItems = append(updatedItems, item)
				}
			}
			userPrivateData["singleCheckboxItems"] = updatedItems
			delete(requiredFields, "singleCheckbox")
		}
	}

	logger.GetLogger().Infof("multiCheckboxType: %s", reflect.TypeOf((params["multiCheckbox"])))
	logger.GetLogger().Infof("multiCheckboxItemsType: %s", reflect.TypeOf((params["multiCheckboxItems"])))
	if multiCheckbox, ok := params["multiCheckbox"].([]interface{}); ok {
		if len(multiCheckbox) > 0 {
			valuesMap := make(map[int]bool)
			for _, v := range multiCheckbox {
				logger.GetLogger().Infof("multiCheckboxValueType: %s", reflect.TypeOf((v)))
				multiCheckboxValue := -1
				if val, ok := v.(float64); ok {
					multiCheckboxValue = int(val)
				} else if multiCheckboxValueStr, ok := v.(string); ok && multiCheckboxValueStr != "" {
					multiCheckboxValue, _ = strconv.Atoi(multiCheckboxValueStr)
				}
				if multiCheckboxValue >= 0 {
					valuesMap[int(multiCheckboxValue)] = true
				}
			}

			if itemsRaw, ok := params["multiCheckboxItems"].([]interface{}); ok {
				var updatedItems []map[string]interface{}
				for _, itemRaw := range itemsRaw {
					if item, ok := itemRaw.(map[string]interface{}); ok {
						value, _ := item["value"].(float64)
						_, checked := valuesMap[int(value)]
						item["checked"] = checked
						updatedItems = append(updatedItems, item)
					}
				}
				userPrivateData["multiCheckboxItems"] = updatedItems
				delete(requiredFields, "multiCheckbox")
			}
		}
	}

	if len(requiredFields) > 0 {
		errMsg := "表单未填写完整，以下是必填项: "
		for _, v := range requiredFields {
			errMsg += fmt.Sprintf("%s ", v)
		}
		return convertJsonValuesToString(UserPrivateData{"errMsg": errMsg}), fmt.Errorf(errMsg)
	}

	userPrivateData["submitBtnText"] = "已提交"
	userPrivateData["submitBtnStatus"] = "disabled"
	return convertJsonValuesToString(userPrivateData), nil
}

func onCardCallback(ctx context.Context, request *card.CardRequest) (*card.CardResponse, error) {
	/**
	 * 卡片事件回调文档：https://open.dingtalk.com/document/orgapp/event-callback-card
	 */
	logger.GetLogger().Infof("card callback message: %v", request)

	params := request.CardActionData.CardPrivateData.Params
	userPrivateData, _ := checkRequiredFields(params)

	response := &card.CardResponse{
		CardUpdateOptions: &card.CardUpdateOptions{
			UpdateCardDataByKey:    true,
			UpdatePrivateDataByKey: true,
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
