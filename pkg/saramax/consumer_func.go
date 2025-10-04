package saramax

import (
	"archi/pkg/logger"
	"encoding/json"
	"github.com/IBM/sarama"
)

type ConsumerFunc[T any] struct {
	l  logger.Logger
	fn func(msg *sarama.ConsumerMessage, event T) error
}

func NewHandler[T any](l logger.Logger, fn func(msg *sarama.ConsumerMessage, event T) error) *ConsumerFunc[T] {
	return &ConsumerFunc[T]{l: l, fn: fn}
}

func (h *ConsumerFunc[T]) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h *ConsumerFunc[T]) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h *ConsumerFunc[T]) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	msgs := claim.Messages()
	for msg := range msgs {
		// 在这里调用业务处理逻辑
		var t T
		err := json.Unmarshal(msg.Value, &t)
		if err != nil {
			// 你也可以在这里引入重试的逻辑
			h.l.Error("反序列消息体失败",
				logger.String("topic", msg.Topic),
				logger.Int32("partition", msg.Partition),
				logger.Int64("offset", msg.Offset),
				logger.Error(err))
		}
		err = h.fn(msg, t)
		if err != nil {
			h.l.Error("处理消息失败",
				logger.String("topic", msg.Topic),
				logger.Int32("partition", msg.Partition),
				logger.Int64("offset", msg.Offset),
				logger.Error(err))
		}
		session.MarkMessage(msg, "")
	}
	return nil
}
