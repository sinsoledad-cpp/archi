package auth

import (
	"archi/internal/service/sms"
	"context"
	"github.com/golang-jwt/jwt/v5"
)

var _ sms.Service = &Service{}

type Service struct {
	svc sms.Service
	key []byte //用于验证 JWT 签名的密钥
}

func NewService(svc sms.Service, key []byte) sms.Service {
	return &Service{
		svc: svc,
		key: key,
	}
}

type SMSClaims struct {
	jwt.RegisteredClaims
	Tpl string
	// 额外加字段
}

func (s *Service) Send(ctx context.Context, tplToken string, args []string, numbers ...string) error {
	var claims SMSClaims
	_, err := jwt.ParseWithClaims(tplToken, &claims, func(token *jwt.Token) (interface{}, error) {
		return s.key, nil
	})
	if err != nil {
		// 如果 Token 无效、过期或格式错误，则直接返回错误，终止流程。
		return err
	}
	return s.svc.Send(ctx, claims.Tpl, args, numbers...)
}
