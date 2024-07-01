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

var (
	dingtalkClient *DingTalkClient = nil
)

var valueKeyMap = map[string]string{
	"TEXT":                "default_string",
	"DATE":                "default_string",
	"DATETIME":            "default_string",
	"SELECT":              "default_number",
	"MULTI_SELECT":        "default_number_array",
	"CHECKBOX":            "default_boolean",
	"CHECKBOX_LIST":       "checkbox_items",
	"CHECKBOX_LIST_MULTI": "checkbox_items",
}

var formFieldsByInstanceId = map[string][]map[string]interface{}{}

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
	CARD_TEMPLATE_ID := "9f86e003-e65e-4680-bf4b-8df5958d9f17.schema" // 该模板只用于测试使用，如需投入线上使用，请导入卡片模板 json 到自己的应用下
	// 卡片公有数据，非字符串类型的卡片数据参考文档：https://open.dingtalk.com/document/orgapp/instructions-for-filling-in-api-card-data

	formFields := []map[string]interface{}{
		{
			"type":           "TEXT",
			"required":       true,
			"name":           "text_required",
			"label":          "必填文本输入",
			"placeholder":    "请输入文本",
			"default_string": "",
		},
		{
			"type":           "TEXT",
			"name":           "text",
			"label":          "文本输入",
			"placeholder":    "请输入文本",
			"default_string": "",
		},
		{
			"type":        "DATE",
			"required":    true,
			"name":        "date_required",
			"label":       "必填日期选择",
			"placeholder": "请选择日期",
		},
		{
			"type":           "DATE",
			"name":           "date",
			"label":          "日期选择",
			"placeholder":    "请选择日期",
			"default_string": "2024-06-06",
		},
		{
			"type":        "DATETIME",
			"required":    true,
			"name":        "datetime_required",
			"label":       "必填日期时间选择",
			"placeholder": "请选择日期时间",
		},
		{
			"type":           "DATETIME",
			"name":           "datetime",
			"label":          "日期时间选择",
			"placeholder":    "请选择日期时间",
			"default_string": "2024-06-06 12:00",
		},
		{
			"type":        "SELECT",
			"required":    true,
			"name":        "select_required",
			"label":       "必填单选下拉框",
			"placeholder": "单选请选择",
			"options": []map[string]interface{}{
				{"value": 1, "text": map[string]string{"zh_CN": "选项 1"}},
				{"value": 2, "text": map[string]string{"zh_CN": "选项 2"}},
				{"value": 3, "text": map[string]string{"zh_CN": "选项 3"}},
				{"value": 4, "text": map[string]string{"zh_CN": "选项 4"}},
			},
		},
		{
			"type":           "SELECT",
			"name":           "select",
			"label":          "单选下拉框",
			"placeholder":    "单选请选择",
			"default_number": 1,
			"options": []map[string]interface{}{
				{"value": 1, "text": map[string]string{"zh_CN": "选项 1"}},
				{"value": 2, "text": map[string]string{"zh_CN": "选项 2"}},
				{"value": 3, "text": map[string]string{"zh_CN": "选项 3"}},
				{"value": 4, "text": map[string]string{"zh_CN": "选项 4"}},
			},
		},
		{
			"type":                 "MULTI_SELECT",
			"required":             true,
			"name":                 "multi_select",
			"label":                "必填多选下拉框",
			"placeholder":          "多选请选择",
			"default_number_array": []int{0, 2},
			"options": []map[string]interface{}{
				{"value": 1, "text": map[string]string{"zh_CN": "选项 1"}},
				{"value": 2, "text": map[string]string{"zh_CN": "选项 2"}},
				{"value": 3, "text": map[string]string{"zh_CN": "选项 3"}},
				{"value": 4, "text": map[string]string{"zh_CN": "选项 4"}},
			},
		},
		{
			"type":     "CHECKBOX_LIST",
			"required": true,
			"name":     "checkbox_list",
			"label":    "必填单选列表",
			"checkbox_items": []map[string]interface{}{
				{
					"value":   0,
					"text":    "选项 0",
					"checked": false,
					"name":    "checkbox_list",
					"type":    "CHECKBOX_LIST",
				},
				{
					"value":   1,
					"text":    "选项 1",
					"checked": false,
					"name":    "checkbox_list",
					"type":    "CHECKBOX_LIST",
				},
				{
					"value":   2,
					"text":    "选项 2",
					"checked": false,
					"name":    "checkbox_list",
					"type":    "CHECKBOX_LIST",
				},
				{
					"value":   3,
					"text":    "选项 3",
					"checked": false,
					"name":    "checkbox_list",
					"type":    "CHECKBOX_LIST",
				},
			},
		},
		{
			"type":     "CHECKBOX_LIST_MULTI",
			"required": true,
			"name":     "checkbox_list_multi",
			"label":    "必填多选列表",
			"checkbox_items": []map[string]interface{}{
				{
					"value":   0,
					"text":    "选项 0",
					"checked": false,
					"name":    "checkbox_list_multi",
					"type":    "CHECKBOX_LIST_MULTI",
				},
				{
					"value":   1,
					"text":    "选项 1",
					"checked": true,
					"name":    "checkbox_list_multi",
					"type":    "CHECKBOX_LIST_MULTI",
				},
				{
					"value":   2,
					"text":    "选项 2",
					"checked": false,
					"name":    "checkbox_list_multi",
					"type":    "CHECKBOX_LIST_MULTI",
				},
				{
					"value":   3,
					"text":    "选项 3",
					"checked": true,
					"name":    "checkbox_list_multi",
					"type":    "CHECKBOX_LIST_MULTI",
				},
			},
		},
		{"type": "CHECKBOX", "name": "checkbox", "label": "复选框"},
		{
			"type":            "CHECKBOX",
			"name":            "checkbox_default_true",
			"label":           "复选框默认勾选",
			"default_boolean": true,
		},
	}
	cardParamMap := map[string]interface{}{
		"form_fields": formFields,
		"form_status": "normal",
		"button_text": "提交",
		"title":       content,
		"err_msg":     "",
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
	formFieldsByInstanceId[outTrackId] = formFields
	return []byte(""), nil
}

func getBoolValueByKey(obj map[string]interface{}, key string) bool {
	if value, ok := obj[key]; ok {
		if v, ok := value.(bool); ok {
			return v
		}
	}
	return false
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func isNonNegativeInt(v interface{}) (int, bool) {
	switch value := v.(type) {
	case float64:
		if intValue := int(value); value == float64(intValue) && intValue >= 0 {
			return intValue, true
		}
	case int:
		if value >= 0 {
			return value, true
		}
	case int64:
		if value >= 0 {
			return int(value), true
		}
	case json.Number:
		if intValue, err := value.Int64(); err == nil && intValue >= 0 {
			return int(intValue), true
		}
	}
	return 0, false
}

func isEmpty(value interface{}) bool {
	if value == nil {
		return true
	}

	v := reflect.ValueOf(value)

	switch v.Kind() {
	case reflect.String:
		return v.Len() == 0
	case reflect.Array, reflect.Slice, reflect.Map:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	default:
		return false
	}
}

func interfaceToString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// 处理 []map[string]interface{}
func handleMapSlice(items []map[string]interface{}, params map[string]interface{}, checkboxListMulti bool) []interface{} {
	var updateValue []interface{}

	for _, item := range items {
		isItemOnTap := interfaceToString(item["value"]) == interfaceToString(params["value"])
		if checkboxListMulti {
			if isItemOnTap {
				item["checked"] = !item["checked"].(bool)
			}
		} else {
			item["checked"] = isItemOnTap
		}
		updateValue = append(updateValue, item)
	}

	return updateValue
}

// 处理 []interface{}
func handleInterfaceSlice(items []interface{}, params map[string]interface{}, checkboxListMulti bool) []interface{} {
	var updateValue []interface{}

	for _, item := range items {
		if itemMap, ok := item.(map[string]interface{}); ok {
			isItemOnTap := interfaceToString(itemMap["value"]) == interfaceToString(params["value"])
			if checkboxListMulti {
				if isItemOnTap {
					itemMap["checked"] = !itemMap["checked"].(bool)
				}
			} else {
				itemMap["checked"] = isItemOnTap
			}
			updateValue = append(updateValue, itemMap)
		}
	}

	return updateValue
}

func onCardCallback(ctx context.Context, request *card.CardRequest) (*card.CardResponse, error) {
	/**
	 * 卡片事件回调文档：https://open.dingtalk.com/document/orgapp/event-callback-card
	 */
	logger.GetLogger().Infof("card callback message: %v", request)

	userPrivateData := make(UserPrivateData)

	actionId := request.CardActionData.CardPrivateData.ActionIdList[0]
	params := request.CardActionData.CardPrivateData.Params

	submitFormFields := []interface{}{}
	if value, ok := params["submit_form_fields"]; ok {
		if fields, ok := value.([]interface{}); ok {
			submitFormFields = fields
		}
	}

	if len(submitFormFields) > 0 {
		// 提交表单，做必填校验，响应错误提示或者响应提交成功处理
		requiredErrorLabels := []string{}
		for _, field := range submitFormFields {
			formField := field.(map[string]interface{})
			formFieldType := formField["type"].(string)

			if _, exists := valueKeyMap[formFieldType]; !exists {
				userPrivateData["err_msg"] = fmt.Sprintf("无效的表单类型「%s」", formFieldType)
				break
			}
			formFieldRequired := getBoolValueByKey(formField, "required")
			formFieldValue := formField[valueKeyMap[formFieldType]]
			logger.GetLogger().Infof("formFieldType: %v; formFieldValue: %v", formFieldType, formFieldValue)
			checkboxListTypes := []string{"CHECKBOX_LIST", "CHECKBOX_LIST_MULTI"}
			if contains(checkboxListTypes, formFieldType) {
				checkedCount := 0
				for _, item := range formFieldValue.([]interface{}) {
					if checked, ok := item.(map[string]interface{})["checked"].(bool); ok && checked {
						checkedCount++
					}
				}
				if formFieldRequired && checkedCount == 0 {
					requiredErrorLabels = append(requiredErrorLabels, formField["label"].(string))
				}
			} else if formFieldType == "SELECT" {
				if _, ok := isNonNegativeInt(formFieldValue); !ok {
					if formFieldRequired {
						requiredErrorLabels = append(requiredErrorLabels, formField["label"].(string))
					}
				}
			} else {
				if formFieldRequired && isEmpty(formFieldValue) {
					requiredErrorLabels = append(requiredErrorLabels, formField["label"].(string))
				}
			}
		}

		if userPrivateData["err_msg"] == nil {
			if len(requiredErrorLabels) > 0 {
				userPrivateData["err_msg"] = fmt.Sprintf("请填写必填项: 「%s」", strings.Join(requiredErrorLabels, ", "))
			} else {
				userPrivateData["form_status"] = "disabled"
				userPrivateData["button_text"] = "已提交"
			}
		}
	} else {
		updateName := params["name"].(string)
		if isEmpty(updateName) {
			userPrivateData["err_msg"] = "服务异常"
		} else {
			cardInstanceId := request.OutTrackId
			logger.GetLogger().Infof("cardInstanceId: %v", cardInstanceId)
			formFields := formFieldsByInstanceId[cardInstanceId] // formFields
			for _, formField := range formFields {
				if formField["name"].(string) == updateName {
					remove := false
					if _, exists := params["remove"]; exists {
						remove = true
					}

					updateType := params["type"].(string)
					updateKey := valueKeyMap[updateType]
					var updateValue interface{}

					if remove && updateType == "multiSelect" {
						updateType = "MULTI_SELECT"
						updateName = actionId
						updateKey = valueKeyMap[updateType]
						removeIndex := int(params[actionId].(map[string]interface{})["index"].(float64))
						updateValue = []int{}
						for _, v := range formField[updateKey].([]int) {
							if v != removeIndex {
								updateValue = append(updateValue.([]int), v)
							}
						}
					} else if updateType == "CHECKBOX_LIST" {
						updateValue = []interface{}{}
						items := formField[updateKey]
						switch items := items.(type) {
						case []map[string]interface{}:
							updateValue = handleMapSlice(items, params, false)
						case []interface{}:
							updateValue = handleInterfaceSlice(items, params, false)
						default:
							logger.GetLogger().Errorf("Unsupported checkbox_items type")
						}
					} else if updateType == "CHECKBOX_LIST_MULTI" {
						updateValue = []interface{}{}
						items := formField[updateKey]
						switch items := items.(type) {
						case []map[string]interface{}:
							updateValue = handleMapSlice(items, params, true)
						case []interface{}:
							updateValue = handleInterfaceSlice(items, params, true)
						default:
							logger.GetLogger().Errorf("Unsupported checkbox_items type")
						}
					} else {
						updateValue = params[updateName]
						if updateType == "SELECT" {
							updateValue = int(params[updateName].(map[string]interface{})["index"].(float64))
						} else if updateType == "MULTI_SELECT" {
							tmpUpdateValue := params[updateName].(map[string]interface{})["index"].([]interface{})
							updateValue = []int{}
							for _, index := range tmpUpdateValue {
								if floatIndex, ok := index.(float64); ok {
									updateValue = append(updateValue.([]int), int(floatIndex))
								}
							}
						} else if updateType == "CHECKBOX" {
							updateValue = !getBoolValueByKey(formField, updateKey)
						}
					}
					logger.GetLogger().Infof("update name=%v, type=%v, key=%v, value=%v", updateName, updateType, updateKey, updateValue)
					formField[updateKey] = updateValue
				}
			}
			userPrivateData["form_fields"] = formFields
		}
	}

	response := &card.CardResponse{
		CardUpdateOptions: &card.CardUpdateOptions{
			UpdateCardDataByKey:    true,
			UpdatePrivateDataByKey: true,
		},
		UserPrivateData: &card.CardDataDto{
			CardParamMap: convertJsonValuesToString(userPrivateData),
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
