package user

import (
	"archi/internal/domain"
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
)

const topicSyncUser = "sync_user_event"

type UserEvent struct {
	Id       int64  `json:"id"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Nickname string `json:"nickname"`
}

type Producer interface {
	ProduceSyncEvent(ctx context.Context, user domain.User) error
}

type SaramaSyncProducer struct {
	producer sarama.SyncProducer
}

// NewSaramaSyncProducer 创建一个新的 SaramaSyncProducer 实例
func NewSaramaSyncProducer(producer sarama.SyncProducer) Producer {
	return &SaramaSyncProducer{producer: producer}
}

// ProduceSyncEvent 发送一个标签同步事件
func (p *SaramaSyncProducer) ProduceSyncEvent(ctx context.Context, user domain.User) error {
	val, err := json.Marshal(UserEvent{
		Id:       user.ID,
		Email:    user.Email,
		Phone:    user.Phone,
		Nickname: user.Nickname,
	})
	if err != nil {
		return err
	}
	_, _, err = p.producer.SendMessage(&sarama.ProducerMessage{
		Topic: topicSyncUser,
		Value: sarama.ByteEncoder(val),
	})
	return err
}
