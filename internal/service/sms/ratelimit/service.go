package ratelimit

import (
	"archi/internal/service/sms"
	"archi/pkg/limiter"
	"context"
	"errors"
)

var errLimited = errors.New("触发限流")

var _ sms.Service = &Service{}

type Service struct {
	// 被装饰的
	svc     sms.Service
	limiter limiter.Limiter
	key     string
}

func NewService(svc sms.Service, l limiter.Limiter) sms.Service {
	return &Service{
		svc:     svc,
		limiter: l,
		key:     "sms-limiter",
	}
}

func (r *Service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	limited, err := r.limiter.Limit(ctx, r.key)
	if err != nil {
		return err
	}
	if limited {
		return errLimited
	}
	return r.svc.Send(ctx, tplId, args, numbers...)
}
