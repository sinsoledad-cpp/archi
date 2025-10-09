package feed

import (
	"archi/internal/domain"
	"archi/internal/service/feed"
	"archi/pkg/logger"
	"archi/pkg/saramax"
	"context"
	"github.com/IBM/sarama"
	"strconv"
)

const topicFollowEvent = "feed_follow_event"

// FollowEvent 定义了从 Kafka 消息中解析出的数据结构
type FollowEvent struct {
	Follower int64 `json:"follower"`
	Followee int64 `json:"followee"`
}
type FollowEventConsumer struct {
	feedSvc feed.Service
	client  sarama.Client
	l       logger.Logger
}

func NewFollowEventConsumer(svc feed.Service, client sarama.Client, log logger.Logger) *FollowEventConsumer {
	return &FollowEventConsumer{
		feedSvc: svc,
		client:  client,
		l:       log,
	}
}

func (f *FollowEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("feed_follow_event", f.client)
	if err != nil {
		return err
	}
	go func() {
		_ = cg.Consume(context.Background(),
			[]string{topicFollowEvent},
			saramax.NewHandler[FollowEvent](f.l, f.Consume))
	}()
	return nil
}

func (f *FollowEventConsumer) Consume(msg *sarama.ConsumerMessage, evt FollowEvent) error {
	return f.feedSvc.CreateFeedEvent(context.Background(), domain.FeedEvent{
		Type: feed.FollowEventName,
		Ext: map[string]string{
			"follower": strconv.FormatInt(evt.Follower, 10),
			"followee": strconv.FormatInt(evt.Followee, 10),
		},
	})
}
