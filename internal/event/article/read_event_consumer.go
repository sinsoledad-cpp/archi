package article

import (
	"archi/internal/repository"
	"archi/pkg/logger"
	"archi/pkg/samarax"
	"context"
	"github.com/IBM/sarama"
	"time"
)

type ReadEventConsumer struct {
	repo   repository.InteractiveRepository
	client sarama.Client
	l      logger.Logger
}

func NewReadEventConsumer(repo repository.InteractiveRepository, client sarama.Client, l logger.Logger) *ReadEventConsumer {
	return &ReadEventConsumer{repo: repo, client: client, l: l}
}
func (i *ReadEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("interactive", i.client)
	if err != nil {
		return err
	}
	go func() {
		er := cg.Consume(context.Background(),
			[]string{TopicReadEvent},
			samarax.NewHandler[ReadEvent](i.l, i.Consume))
		if er != nil {
			i.l.Error("退出消费", logger.Error(er))
		}
	}()
	return err
}

func (i *ReadEventConsumer) Consume(msg *sarama.ConsumerMessage,
	event ReadEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return i.repo.IncrReadCnt(ctx, "article", event.Aid)
}

func (i *ReadEventConsumer) StartS() error {
	cg, err := sarama.NewConsumerGroupFromClient("interactive", i.client)
	if err != nil {
		return err
	}
	go func() {
		er := cg.Consume(
			context.Background(),
			[]string{TopicReadEvent},
			samarax.NewBatchConsumerAtomicFunc[ReadEvent](i.l, i.BatchConsume))
		if er != nil {
			i.l.Error("退出消费", logger.Error(er))
		}
	}()
	return err
}
func (i *ReadEventConsumer) BatchConsume(msgs []*sarama.ConsumerMessage, events []ReadEvent) error {
	bizs := make([]string, 0, len(events))
	bizIDs := make([]int64, 0, len(events))
	for _, evt := range events {
		bizs = append(bizs, "article")
		bizIDs = append(bizIDs, evt.Aid)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return i.repo.BatchIncrReadCnt(ctx, bizs, bizIDs)
}
