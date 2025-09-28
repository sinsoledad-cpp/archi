package memory

import (
	"archi/internal/service/sms"
	"context"
	"fmt"
)

var _ sms.Service = &Service{}

type Service struct {
}

func NewService() sms.Service {
	return &Service{}
}

func (s *Service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	fmt.Println("验证码是", args)
	return nil
}
