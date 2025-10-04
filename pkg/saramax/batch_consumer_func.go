package saramax

import (
	"archi/pkg/logger"
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
	"time"
)

// BatchConsumerAtomicFunc 是一种将普通函数适配为 sarama.ConsumerGroupHandler 的类型。
// 它将整个消息批次作为原子单元处理。如果业务函数 fn 返回任何错误，
// 整个批次的消息都不会被确认。
type BatchConsumerAtomicFunc[T any] struct {
	fn func(msgs []*sarama.ConsumerMessage, events []T) error
	l  logger.Logger
}

// NewBatchConsumerAtomicFunc 创建一个新的 BatchConsumerAtomicFunc。
// 业务函数 fn 对整个批次进行原子性处理。
func NewBatchConsumerAtomicFunc[T any](l logger.Logger, fn func(msgs []*sarama.ConsumerMessage, events []T) error) *BatchConsumerAtomicFunc[T] {
	return &BatchConsumerAtomicFunc[T]{fn: fn, l: l}
}

func (b *BatchConsumerAtomicFunc[T]) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (b *BatchConsumerAtomicFunc[T]) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (b *BatchConsumerAtomicFunc[T]) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	msgsCh := claim.Messages()
	const batchSize = 10
	for {
		batch := make([]*sarama.ConsumerMessage, 0, batchSize)
		ts := make([]T, 0, batchSize)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		done := false

		// 批次收集循环
		for i := 0; i < batchSize && !done; i++ {
			select {
			case <-ctx.Done():
				done = true
			case msg, ok := <-msgsCh:
				if !ok {
					cancel()
					// 在退出前，处理最后一批已收集的消息
					b.processBatch(session, batch, ts)
					b.l.Info("消费通道已关闭，ConsumerClaim 退出")
					return nil
				}

				var t T
				err := json.Unmarshal(msg.Value, &t)
				if err != nil {
					b.l.Error("反序列化消息体失败，消息将被跳过",
						logger.String("topic", msg.Topic),
						logger.Int32("partition", msg.Partition),
						logger.Int64("offset", msg.Offset),
						logger.Error(err),
					)
					session.MarkMessage(msg, "")
					continue
				}
				batch = append(batch, msg)
				ts = append(ts, t)
			}
		}
		cancel()

		// 处理收集到的批次
		b.processBatch(session, batch, ts)
	}
}

// processBatch 封装了批次处理和消息确认的核心逻辑
func (b *BatchConsumerAtomicFunc[T]) processBatch(session sarama.ConsumerGroupSession, batch []*sarama.ConsumerMessage, ts []T) {
	if len(batch) == 0 {
		return
	}

	err := b.fn(batch, ts)
	if err != nil {
		// 记录失败的日志，并提供详细信息
		msgOffsets := make([]int64, len(batch))
		for j, m := range batch {
			msgOffsets[j] = m.Offset
		}
		b.l.Error("处理消息批次失败，该批次所有消息将不会被确认",
			logger.Error(err),
			logger.String("topic", batch[0].Topic),
			logger.Field{Key: "offsets", Val: msgOffsets},
		)
		// 直接返回，不执行 MarkMessage，以便 Kafka 重新投递
		return
	}

	// 只有在整个批次处理成功后，才确认所有消息
	for _, msg := range batch {
		session.MarkMessage(msg, "")
	}
}

/*
// BatchConsumerFunc 是一种将普通函数适配为 sarama.ConsumerGroupHandler 的类型，用于批量处理消息。
// 业务函数 fn 应该返回一个 error 切片，其长度与传入的消息数相同。
// 切片中的 nil 值表示对应位置的消息处理成功。
type BatchConsumerFunc[T any] struct {
	fn func(msgs []*sarama.ConsumerMessage, ts []T) []error
	l  logger.Logger
}

// NewBatchConsumerFunc 创建一个新的 BatchConsumerFunc。
// 业务函数 fn 必须返回一个 error 切片，用于精确标记每条消息的处理结果。
func NewBatchConsumerFunc[T any](l logger.Logger, fn func(msgs []*sarama.ConsumerMessage, ts []T) []error) *BatchConsumerFunc[T] {
	return &BatchConsumerFunc[T]{fn: fn, l: l}
}

func (b *BatchConsumerFunc[T]) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (b *BatchConsumerFunc[T]) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (b *BatchConsumerFunc[T]) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	msgsCh := claim.Messages()
	const batchSize = 10
	for {
		batch := make([]*sarama.ConsumerMessage, 0, batchSize)
		ts := make([]T, 0, batchSize)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		done := false

		// 批次收集循环
		for i := 0; i < batchSize && !done; i++ {
			select {
			case <-ctx.Done():
				// 等待超时，当前批次收集结束
				done = true
			case msg, ok := <-msgsCh:
				if !ok {
					// msgsCh 已关闭，消费者需要退出
					cancel() // 确保 context 被清理
					// 在退出前，处理最后一批已收集的消息
					b.processBatch(session, batch, ts)
					b.l.Info("消费通道已关闭，ConsumerClaim 退出")
					return nil
				}

				var t T
				err := json.Unmarshal(msg.Value, &t)
				if err != nil {
					b.l.Error("反序列化消息体失败，消息将被跳过",
						logger.String("topic", msg.Topic),
						logger.Int32("partition", msg.Partition),
						logger.Int64("offset", msg.Offset),
						logger.Error(err),
					)
					// 将有问题的消息直接确认掉，防止它阻塞消费流程。
					// 或者根据业务需求，可以将其发送到死信队列。
					session.MarkMessage(msg, "")
					continue
				}
				// 只有在反序列化成功后，才将消息和事件成对加入批次
				batch = append(batch, msg)
				ts = append(ts, t)
			}
		}
		cancel()

		// 处理收集到的批次
		b.processBatch(session, batch, ts)
	}
}

// processBatch 封装了批次处理和消息确认的核心逻辑
func (b *BatchConsumerFunc[T]) processBatch(session sarama.ConsumerGroupSession, batch []*sarama.ConsumerMessage, ts []T) {
	if len(batch) == 0 {
		return
	}

	errs := b.fn(batch, ts)

	// 确保 errs 的长度与 batch 长度一致，防止业务逻辑实现错误导致 panic
	if len(errs) != len(batch) {
		b.l.Error("业务函数返回的 error 切片长度与消息批次长度不匹配",
			logger.Int("batch_size", len(batch)),
			logger.Int("errors_size", len(errs)))
		// 在这种情况下，我们保守地认为整个批次都处理失败了，不进行任何消息确认。
		// 这样可以防止数据丢失，但可能会导致消息重复消费。
		return
	}

	// 遍历处理结果，只确认成功的消息
	for i, err := range errs {
		if err != nil {
			// 记录每一条处理失败的消息
			b.l.Error("单条消息处理失败",
				logger.Error(err),
				logger.String("topic", batch[i].Topic),
				logger.Int64("offset", batch[i].Offset),
			)
		} else {
			// 只确认处理成功的消息
			session.MarkMessage(batch[i], "")
		}
	}
}
*/
