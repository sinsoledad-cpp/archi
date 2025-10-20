package memory

import (
	"archi/internal/service/sms"
	"context"
	"fmt"
	"time"
)

var _ sms.Service = &Service{}

type Service struct {
}

func NewService() sms.Service {
	return &Service{}
}

func (s *Service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	time.Sleep(time.Second * 1)
	fmt.Println("验证码是", args)
	return nil
}
