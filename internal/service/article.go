package service

import (
	"archi/internal/domain"
	"archi/internal/event/article"
	"archi/internal/repository"
	"archi/pkg/logger"
	"context"
	"time"
)

//go:generate mockgen -source=./article.go -package=svcmocks -destination=./mocks/article.mock.go ArticleService
type ArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
	Publish(ctx context.Context, art domain.Article) (int64, error)
	Withdraw(ctx context.Context, uid int64, id int64) error
	GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPubById(ctx context.Context, id, uid int64) (domain.Article, error)
	ListPub(ctx context.Context, start time.Time, offset, limit int) ([]domain.Article, error)
}
type DefaultArticleService struct {
	repo repository.ArticleRepository

	producer article.Producer

	userRepo repository.UserRepository

	// V1 写法专用
	//readerRepo repository.ArticleReaderRepository
	//authorRepo repository.ArticleAuthorRepository
	l logger.Logger
}

func NewDefaultArticleService(repo repository.ArticleRepository, producer article.Producer) ArticleService {
	return &DefaultArticleService{
		repo:     repo,
		producer: producer,
	}
}

func (a *DefaultArticleService) Save(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusUnpublished
	if art.ID > 0 {
		err := a.repo.Update(ctx, art)
		return art.ID, err
	}
	return a.repo.Create(ctx, art)
}

func (a *DefaultArticleService) Publish(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusPublished
	return a.repo.Sync(ctx, art)
}
func (a *DefaultArticleService) Withdraw(ctx context.Context, uid int64, id int64) error {
	return a.repo.SyncStatus(ctx, uid, id, domain.ArticleStatusPrivate)
}
func (a *DefaultArticleService) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	return a.repo.GetByAuthor(ctx, uid, offset, limit)
}
func (a *DefaultArticleService) GetById(ctx context.Context, id int64) (domain.Article, error) {
	return a.repo.GetById(ctx, id)
}
func (a *DefaultArticleService) GetPubById(ctx context.Context, id, uid int64) (domain.Article, error) {
	res, err := a.repo.GetPubById(ctx, id)
	go func() {
		if err == nil {
			// 在这里发一个消息
			er := a.producer.ProduceReadEvent(article.ReadEvent{
				Aid: id,
				Uid: uid,
			})
			if er != nil {
				a.l.Error("发送 ReadEvent 失败",
					logger.Int64("aid", id),
					logger.Int64("uid", uid),
					logger.Error(err))
			}
		}
	}()

	return res, err
}
func (a *DefaultArticleService) ListPub(ctx context.Context, start time.Time, offset, limit int) ([]domain.Article, error) {
	return a.repo.ListPub(ctx, start, offset, limit)
}
