package sms

import (
	"github.com/denverdino/aliyungo/common"
	"net/http"
)

//阿里云通信
type SendSmsArgs struct {
	PhoneNumbers    string
	SignName        string
	TemplateCode    string
	TemplateParam   string
	SmsUpExtendCode string `ArgName:"smsUpExtendCode"`
	OutId           string
}

type SendSmsResponse struct {
	common.Response
	Code    string
	Message string
	BizId   string
}

func (this *DYSmsClient) SendSms(args *SendSmsArgs) (*SendSmsResponse, error) {
	resp := SendSmsResponse{}
	return &resp, this.InvokeByAnyMethod(http.MethodGet, SendSms, "", args, &resp)
}
