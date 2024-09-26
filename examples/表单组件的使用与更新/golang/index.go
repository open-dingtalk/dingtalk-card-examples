package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
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

type Field struct {
	Name         string         `json:"name"`
	Label        string         `json:"label,omitempty"`
	Type         string         `json:"type"`
	Required     bool           `json:"required,omitempty"`
	Hidden       bool           `json:"hidden,omitempty"`
	Placeholder  string         `json:"placeholder,omitempty"`
	RequiredMsg  string         `json:"requiredMsg,omitempty"`
	ReadOnly     bool           `json:"readOnly,omitempty"`
	DefaultValue interface{}    `json:"defaultValue,omitempty"`
	Options      []SelectOption `json:"options,omitempty"`
}

type SelectOption struct {
	Value string `json:"value"`
	Text  string `json:"text"`
}

type Form struct {
	Fields []Field `json:"fields"`
}

type CardData struct {
	Form        Form   `json:"form"`
	FormStatus  string `json:"form_status"`
	FormBtnText string `json:"form_btn_text"`
	Title       string `json:"title"`
}

type UserPrivateData map[string]interface{}

type DingtalkStream struct {
	DingtalkClient  string
	IncomingMessage string
}

type CardReplier struct {
	DingtalkClient  string
	IncomingMessage string
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
	CARD_TEMPLATE_ID := "280f6d7a-63bc-4905-bf3f-4c6d95e5166b.schema" // 该模板只用于测试使用，如需投入线上使用，请导入卡片模板 json 到自己的应用下
	// 卡片公有数据，非字符串类型的卡片数据参考文档：https://open.dingtalk.com/document/orgapp/instructions-for-filling-in-api-card-data

	cardData := CardData{
		Form: Form{
			Fields: []Field{
				{
					Name:         "system_params_1",
					Type:         "TEXT",
					Hidden:       true,
					DefaultValue: "asdf",
				},
				{
					Name:        "text",
					Label:       "必填文本输入",
					Type:        "TEXT",
					Required:    true,
					Placeholder: "请输入文本",
					RequiredMsg: "自定义必填错误提示",
				},
				{
					Name:        "text_optional",
					Label:       "非必填文本输入",
					Type:        "TEXT",
					Placeholder: "请输入文本",
				},
				{
					Name:         "text_readonly",
					Label:        "非必填只读文本输入有默认值",
					Type:         "TEXT",
					ReadOnly:     true,
					DefaultValue: "文本默认值",
				},
				{
					Name:        "date",
					Label:       "必填日期选择",
					Type:        "DATE",
					Required:    true,
					Placeholder: "请选择日期",
				},
				{
					Name:        "date_optional",
					Label:       "非必填日期选择",
					Type:        "DATE",
					Placeholder: "请选择日期",
				},
				{
					Name:         "date_readonly",
					Label:        "非必填只读日期选择有默认值",
					Type:         "DATE",
					ReadOnly:     true,
					DefaultValue: "2024-05-27",
				},
				{
					Name:        "datetime",
					Label:       "必填日期时间选择",
					Type:        "DATETIME",
					Required:    true,
					Placeholder: "请选择日期时间",
				},
				{
					Name:        "datetime_optional",
					Label:       "非必填日期时间选择",
					Type:        "DATETIME",
					Placeholder: "请选择日期时间",
				},
				{
					Name:         "datetime_readonly",
					Label:        "非必填只读日期时间选择有默认值",
					Type:         "DATETIME",
					ReadOnly:     true,
					DefaultValue: "2024-05-27 12:00",
				},
				{
					Name:        "select",
					Label:       "必填单选",
					Type:        "SELECT",
					Required:    true,
					Placeholder: "单选请选择",
					Options: []SelectOption{
						{"1", "选项1"},
						{"2", "选项2"},
						{"3", "选项3"},
						{"4", "选项4"},
					},
				},
				{
					Name:        "select_optional",
					Label:       "非必填单选",
					Type:        "SELECT",
					Placeholder: "单选请选择",
					Options: []SelectOption{
						{"1", "选项1"},
						{"2", "选项2"},
						{"3", "选项3"},
						{"4", "选项4"},
					},
				},
				{
					Name:         "select_readonly",
					Label:        "非必填只读单选有默认值",
					Type:         "SELECT",
					ReadOnly:     true,
					DefaultValue: map[string]interface{}{"index": 3, "value": "4"},
					Options: []SelectOption{
						{"1", "选项1"},
						{"2", "选项2"},
						{"3", "选项3"},
						{"4", "选项4"},
					},
				},
				{
					Name:        "multi_select",
					Label:       "必填多选",
					Type:        "MULTI_SELECT",
					Required:    true,
					Placeholder: "多选请选择",
					Options: []SelectOption{
						{"1", "选项1"},
						{"2", "选项2"},
						{"3", "选项3"},
						{"4", "选项4"},
					},
				},
				{
					Name:        "multi_select_optional",
					Label:       "非必填多选",
					Type:        "MULTI_SELECT",
					Placeholder: "多选请选择",
					Options: []SelectOption{
						{"1", "选项1"},
						{"2", "选项2"},
						{"3", "选项3"},
						{"4", "选项4"},
					},
				},
				{
					Name:         "multi_select_readonly",
					Label:        "非必填只读多选有默认值",
					Type:         "MULTI_SELECT",
					ReadOnly:     true,
					DefaultValue: map[string]interface{}{"index": []int{1, 3}, "value": []string{"2", "4"}},
					Options: []SelectOption{
						{"1", "选项1"},
						{"2", "选项2"},
						{"3", "选项3"},
						{"4", "选项4"},
					},
				},
				{
					Name:  "checkbox",
					Label: "独立的复选框",
					Type:  "CHECKBOX",
				},
				{
					Name:         "checkbox_readonly",
					Label:        "只读独立的复选框",
					Type:         "CHECKBOX",
					ReadOnly:     true,
					DefaultValue: true,
				},
			},
		},
		FormStatus:  "normal",
		FormBtnText: "提交",
		Title:       "content",
	}

	cardDataMap := make(map[string]interface{})
	cardDataBytes, _ := json.Marshal(cardData)
	json.Unmarshal(cardDataBytes, &cardDataMap)
	cardParamMap := convertJsonValuesToTeaString(cardDataMap)

	requestCardData := &dingtalkcard_1_0.CreateAndDeliverRequestCardData{
		CardParamMap: cardParamMap,
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
		CardData:       requestCardData,
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

	// 创建并投放卡片: https://open.dingtalk.com/document/orgapp/create-and-deliver-cards
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

	userPrivateData := make(UserPrivateData)

	params := request.CardActionData.CardPrivateData.Params

	form := params["form"].(map[string]interface{})
	var currentForm Form
	currentFormBytes, _ := json.Marshal(params["current_form"])
	json.Unmarshal(currentFormBytes, &currentForm)

	if form != nil && len(currentForm.Fields) > 0 {
		logger.GetLogger().Infof("form: %v", form)
		for i, field := range currentForm.Fields {
			if submitValue, exists := form[field.Name]; exists {
				currentForm.Fields[i].DefaultValue = submitValue
			}
		}

		userPrivateData["form"] = currentForm
		userPrivateData["form_btn_text"] = "已提交"
		userPrivateData["form_status"] = "disabled"
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
