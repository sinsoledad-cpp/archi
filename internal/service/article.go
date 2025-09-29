package service

import (
	"archi/internal/domain"
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
type articleService struct {
	repo repository.ArticleRepository
	//producer article.Producer

	userRepo repository.UserRepository

	// V1 写法专用
	//readerRepo repository.ArticleReaderRepository
	//authorRepo repository.ArticleAuthorRepository
	l logger.Logger
}
