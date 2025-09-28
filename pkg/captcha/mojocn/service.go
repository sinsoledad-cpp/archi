package mojocn

import (
	"archi/pkg/captcha"
	"context"
	"github.com/mojocn/base64Captcha"
)

type Service struct {
	store base64Captcha.Store
}

// Response 是生成验证码的响应结构

// NewService var store = base64Captcha.DefaultMemStore
func NewService(store base64Captcha.Store) *Service {
	return &Service{
		store: store,
	}
}
func (s *Service) Generate(ctx context.Context) (captcha.Response, error) {
	driver := base64Captcha.NewDriverDigit(80, 240, 5, 0.7, 80)
	c := base64Captcha.NewCaptcha(driver, s.store)
	id, b64s, _, err := c.Generate()
	if err != nil {
		return captcha.Response{}, err
	}
	return captcha.Response{ID: id, B64S: b64s}, nil
}

func (s *Service) Verify(ctx context.Context, id string, value string) bool {
	return s.store.Verify(id, value, true) // true 表示验证成功后立即清除
}
