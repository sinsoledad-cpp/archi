package article

import (
	"archi/internal/domain"
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
)

const topicSyncArticle = "sync_article_event"

type ArticleEvent struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Status  int32  `json:"status"`
	Content string `json:"content"`
}

type Producer interface {
	ProduceReadEvent(evt ReadEvent) error
	ProduceSyncEvent(ctx context.Context, art domain.Article) error
}

type SaramaSyncProducer struct {
	producer sarama.SyncProducer
}

func NewSaramaSyncProducer(producer sarama.SyncProducer) Producer {
	return &SaramaSyncProducer{producer: producer}
}

func (s *SaramaSyncProducer) ProduceReadEvent(evt ReadEvent) error {
	val, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	_, _, err = s.producer.SendMessage(&sarama.ProducerMessage{
		Topic: TopicReadEvent,
		Value: sarama.StringEncoder(val),
	})
	return err
}

// ProduceSyncEvent 发送一个文章同步事件
func (p *SaramaSyncProducer) ProduceSyncEvent(ctx context.Context, art domain.Article) error {
	val, err := json.Marshal(ArticleEvent{
		Id:      art.ID,
		Title:   art.Title,
		Status:  int32(art.Status),
		Content: art.Content,
	})
	if err != nil {
		return err
	}
	_, _, err = p.producer.SendMessage(&sarama.ProducerMessage{
		Topic: topicSyncArticle,
		Value: sarama.ByteEncoder(val),
	})
	return err
}
