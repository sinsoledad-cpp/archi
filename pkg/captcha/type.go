package captcha

import "context"

type Response struct {
	ID   string
	B64S string // Base64 编码的图片 Base64 String
}

// Service 是验证码服务的接口
type Service interface {
	Generate(ctx context.Context) (Response, error)
	Verify(ctx context.Context, id string, value string) bool
}
