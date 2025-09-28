package prometheus

import (
	"archi/internal/service/sms"
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

type Service struct {
	svc    sms.Service
	vector *prometheus.SummaryVec
}

func NewDecorator(svc sms.Service, opt prometheus.SummaryOpts) sms.Service {
	return &Service{
		svc:    svc,
		vector: prometheus.NewSummaryVec(opt, []string{"tpl_id"}),
	}
}

func (s *Service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Milliseconds()
		s.vector.WithLabelValues(tplId).Observe(float64(duration))
	}()
	return s.svc.Send(ctx, tplId, args, numbers...)
}
