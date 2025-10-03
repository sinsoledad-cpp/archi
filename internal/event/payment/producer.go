package payment

import (
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
)

type Producer interface {
	ProducePaymentEvent(ctx context.Context, evt PaymentEvent) error
}

type SaramaSyncProducer struct {
	producer sarama.SyncProducer
}

func NewSaramaSyncProducer(producer sarama.SyncProducer) *SaramaSyncProducer {
	return &SaramaSyncProducer{
		producer: producer,
	}
}

func (s *SaramaSyncProducer) ProducePaymentEvent(ctx context.Context, evt PaymentEvent) error {
	data, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	_, _, err = s.producer.SendMessage(&sarama.ProducerMessage{
		Key:   sarama.StringEncoder(evt.BizTradeNO),
		Topic: evt.Topic(),
		Value: sarama.ByteEncoder(data),
	})
	return err
}
