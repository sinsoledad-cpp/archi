package prometheus

import (
	"archi/internal/domain"
	"archi/internal/service/oauth2/wechat"
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Service struct {
	wechat.Service
	sum prometheus.Summary
}

func NewDecorator(svc wechat.Service, sum prometheus.Summary) *Service {
	return &Service{
		Service: svc,
		sum:     sum,
	}
}

func (s *Service) VerifyCode(ctx context.Context, code string) (domain.WechatInfo, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Milliseconds()
		s.sum.Observe(float64(duration))
	}()
	return s.Service.VerifyCode(ctx, code)
}
