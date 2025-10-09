package follow

import (
	"encoding/json"
	"github.com/IBM/sarama"
)

// topicFollowEvent 定义了事件的名称
const topicFollowEvent = "feed_follow_event"

// Event 定义了关注事件的数据结构
type Event struct {
	Follower int64 `json:"follower"`
	Followee int64 `json:"followee"`
}

// Producer 定义了生产者接口
type Producer interface {
	ProduceFollowEvent(evt Event) error
}

type SaramaFollowEventProducer struct {
	producer sarama.SyncProducer
	topic    string
}

func NewFollowEventProducer(producer sarama.SyncProducer) Producer {
	return &SaramaFollowEventProducer{
		producer: producer,
		topic:    topicFollowEvent,
	}
}

func (p *SaramaFollowEventProducer) ProduceFollowEvent(evt Event) error {
	val, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	_, _, err = p.producer.SendMessage(&sarama.ProducerMessage{
		Topic: p.topic,
		Value: sarama.StringEncoder(val),
	})
	return err
}
