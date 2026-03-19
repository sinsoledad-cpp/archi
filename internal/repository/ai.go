package repository

import (
	"archi/internal/domain"
	"archi/internal/repository/cache"
	"context"
)

type AiRepository interface {
	GetArticleSummary(ctx context.Context, artId int64) (domain.ArticleSummary, error)
	SetArticleSummary(ctx context.Context, artId int64, summary domain.ArticleSummary) error
}

type CachedAiRepository struct {
	cache cache.AiCache
}

func NewCachedAiRepository(cache cache.AiCache) AiRepository {
	return &CachedAiRepository{
		cache: cache,
	}
}

func (c *CachedAiRepository) GetArticleSummary(ctx context.Context, artId int64) (domain.ArticleSummary, error) {
	return c.cache.GetArticleSummary(ctx, artId)
}

func (c *CachedAiRepository) SetArticleSummary(ctx context.Context, artId int64, summary domain.ArticleSummary) error {
	return c.cache.SetArticleSummary(ctx, artId, summary)
}
