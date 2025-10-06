package search

import (
	"archi/internal/domain"
	"archi/internal/service"
	"archi/pkg/logger"
	"archi/pkg/saramax"
	"context"
	"github.com/IBM/sarama"
	"time"
)

const topicSyncArticle = "sync_article_event"

type ArticleConsumer struct {
	syncSvc service.SyncService
	client  sarama.Client
	l       logger.Logger
}

func NewArticleConsumer(client sarama.Client, l logger.Logger, svc service.SyncService) *ArticleConsumer {
	return &ArticleConsumer{
		syncSvc: svc,
		client:  client,
		l:       l,
	}
}

type ArticleEvent struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Status  int32  `json:"status"`
	Content string `json:"content"`
}

func (a *ArticleConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("sync_article", a.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(context.Background(),
			[]string{topicSyncArticle},
			saramax.NewHandler[ArticleEvent](a.l, a.Consume))
		if err != nil {
			a.l.Error("退出了消费循环异常", logger.Error(err))
		}
	}()
	return err
}

func (a *ArticleConsumer) Consume(sg *sarama.ConsumerMessage, evt ArticleEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return a.syncSvc.InputArticle(ctx, a.toDomain(evt))
}

func (a *ArticleConsumer) toDomain(article ArticleEvent) domain.ArticleES {
	return domain.ArticleES{
		Id:      article.Id,
		Title:   article.Title,
		Status:  article.Status,
		Content: article.Content,
	}
}
