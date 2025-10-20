package opentelemetry

import (
	"archi/internal/service/sms"
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Service struct {
	svc    sms.Service
	tracer trace.Tracer
}

func NewService(svc sms.Service) sms.Service {
	return &Service{
		svc:    svc,
		tracer: otel.Tracer("internal/service/sms/opentelemetry"),
	}
}

func (s *Service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {

	ctx, span := s.tracer.Start(ctx, "sms")
	defer span.End()
	span.SetAttributes(attribute.String("tpl", tplId))
	span.AddEvent("发短信")
	err := s.svc.Send(ctx, tplId, args, numbers...)
	if err != nil {
		span.RecordError(err)
	}
	return err
}
