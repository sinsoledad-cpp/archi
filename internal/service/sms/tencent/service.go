package tencent

import (
	"archi/internal/service/sms"
	"context"
	"fmt"
	"github.com/ecodeclub/ekit"
	"github.com/ecodeclub/ekit/slice"
	tencentsms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
)

var _ sms.Service = &Service{}

type Service struct {
	client   *tencentsms.Client
	appID    *string
	signName *string
}

func NewService(client *tencentsms.Client, appID string, signName string) sms.Service {
	return &Service{
		client:   client,
		appID:    &appID,
		signName: &signName,
	}
}

func (s *Service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	request := tencentsms.NewSendSmsRequest()
	request.SetContext(ctx)
	request.SmsSdkAppId = s.appID
	request.SignName = s.signName
	request.TemplateId = ekit.ToPtr[string](tplId)
	request.TemplateParamSet = s.toPtrSlice(args)
	request.PhoneNumberSet = s.toPtrSlice(numbers)
	response, err := s.client.SendSms(request)
	//zap.L().Debug("请求腾讯SendSMS接口",
	//	zap.Any("req", request),
	//	zap.Any("resp", response))
	// 处理异常
	if err != nil {
		return err
	}
	for _, statusPtr := range response.Response.SendStatusSet {
		if statusPtr == nil {
			// 不可能进来这里
			continue
		}
		status := *statusPtr
		if status.Code == nil || *(status.Code) != "Ok" {
			// 发送失败
			return fmt.Errorf("发送短信失败 code: %s, msg: %s", *status.Code, *status.Message)
		}
	}
	return nil
}

func (s *Service) toPtrSlice(data []string) []*string {
	return slice.Map[string, *string](data,
		func(idx int, src string) *string {
			return &src
		})
}
