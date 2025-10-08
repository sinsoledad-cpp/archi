package tag

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
)

//	type SyncDataEventConsumer struct {
//		svc    service.SyncService
//		client sarama.Client
//		l      logger.Logger
//	}
type SyncDataEvent struct {
	IndexName string
	DocID     string
	Data      string
}
type BizTags struct {
	Uid   int64    `json:"uid"`
	Biz   string   `json:"biz"`
	BizId int64    `json:"biz_id"`
	Tags  []string `json:"tags"`
}

type Producer interface {
	ProduceSyncEvent(ctx context.Context, data BizTags) error
}

type SaramaSyncProducer struct {
	client sarama.SyncProducer
}

func NewSaramaSyncProducer(client sarama.SyncProducer) Producer {
	return &SaramaSyncProducer{
		client: client,
	}
}

func (p *SaramaSyncProducer) ProduceSyncEvent(ctx context.Context, tags BizTags) error {
	data, _ := json.Marshal(tags)
	evt := SyncDataEvent{
		IndexName: "tags_index",
		DocID:     fmt.Sprintf("%d_%s_%d", tags.Uid, tags.Biz, tags.BizId),
		Data:      string(data),
	}
	data, _ = json.Marshal(evt)
	_, _, err := p.client.SendMessage(&sarama.ProducerMessage{
		Topic: "search_sync_data",
		Value: sarama.ByteEncoder(data),
	})
	return err
}
