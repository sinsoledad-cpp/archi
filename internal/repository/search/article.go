package search

import (
	"archi/internal/domain"
	"archi/internal/repository/dao/search"
	"context"
	"github.com/ecodeclub/ekit/slice"
)

type ArticleRepository interface {
	InputArticle(ctx context.Context, msg domain.ArticleES) error
	SearchArticle(ctx context.Context, uid int64, keywords []string) ([]domain.ArticleES, error)
}

type DefaultArticleRepository struct {
	dao  search.ArticleDAO
	tags search.TagDAO
}

func NewDefaultArticleRepository(d search.ArticleDAO, td search.TagDAO) ArticleRepository {
	return &DefaultArticleRepository{
		dao:  d,
		tags: td,
	}
}

func (a *DefaultArticleRepository) InputArticle(ctx context.Context, msg domain.ArticleES) error {
	return a.dao.InputArticle(ctx, search.Article{
		Id:      msg.Id,
		Title:   msg.Title,
		Status:  msg.Status,
		Content: msg.Content,
	})
}

func (a *DefaultArticleRepository) SearchArticle(ctx context.Context, uid int64, keywords []string) ([]domain.ArticleES, error) {
	ids, err := a.tags.Search(ctx, uid, "article", keywords)
	if err != nil {
		return nil, err
	}
	arts, err := a.dao.Search(ctx, ids, keywords)
	if err != nil {
		return nil, err
	}
	return slice.Map(arts, func(idx int, src search.Article) domain.ArticleES {
		return domain.ArticleES{
			Id:      src.Id,
			Title:   src.Title,
			Status:  src.Status,
			Content: src.Content,
			Tags:    src.Tags,
		}
	}), nil
}
