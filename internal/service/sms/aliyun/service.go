package aliyun

import (
	"archi/internal/service/sms"
	"context"
	"encoding/json"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
	"strconv"
	"strings"
)

var _ sms.Service = &Service{}

type Service struct {
	client   *dysmsapi.Client
	signName string
}

func NewService(c *dysmsapi.Client, signName string) sms.Service {
	return &Service{
		client:   c,
		signName: signName,
	}
}

func (s *Service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	req := dysmsapi.CreateSendSmsRequest()
	req.Scheme = "https"                          // 使用 HTTPS 协议
	req.PhoneNumbers = strings.Join(numbers, ",") // 阿里云多个手机号为字符串逗号间隔
	req.SignName = s.signName                     //设置短信签名
	// 传的是 JSON
	argsMap := make(map[string]string, len(args))
	for idx, arg := range args {
		argsMap[strconv.Itoa(idx)] = arg
	}
	// 这意味着，你的模板必须是 你的短信验证码是{0}
	// 你的短信验证码是{code}
	bCode, err := json.Marshal(argsMap)
	if err != nil {
		return err
	}
	req.TemplateParam = string(bCode)
	req.TemplateCode = tplId //设置模板 ID

	var resp *dysmsapi.SendSmsResponse
	resp, err = s.client.SendSms(req)
	if err != nil {
		return err
	}

	if resp.Code != "OK" {
		return fmt.Errorf("发送失败，code: %s, 原因：%s",
			resp.Code, resp.Message)
	}
	return nil
}
