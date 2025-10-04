package reward

import (
	"archi/internal/event/payment"
	"archi/internal/service"
	"archi/pkg/logger"
	"archi/pkg/saramax"
	"context"
	"github.com/IBM/sarama"
	"strings"
	"time"
)

type PaymentEventConsumer struct {
	client sarama.Client
	l      logger.Logger
	svc    service.RewardService
}

// Start 这边就是自己启动 goroutine 了
func (r *PaymentEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("reward", r.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(context.Background(),
			[]string{"payment_events"},
			saramax.NewHandler[payment.PaymentEvent](r.l, r.Consume))
		if err != nil {
			r.l.Error("退出了消费循环异常", logger.Error(err))
		}
	}()
	return err
}

func (r *PaymentEventConsumer) Consume(msg *sarama.ConsumerMessage, evt payment.PaymentEvent) error {
	// 不是我们的
	if !strings.HasPrefix(evt.BizTradeNO, "reward") {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	return r.svc.UpdateReward(ctx, evt.BizTradeNO, evt.ToDomainStatus())
}
