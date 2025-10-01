package repository

import (
	"archi/internal/domain"
	"archi/internal/repository/cache"
	"context"
)

type RankingRepository interface {
	GetTopN(ctx context.Context) ([]domain.Article, error)
	ReplaceTopN(ctx context.Context, arts []domain.Article) error
}
type CachedRankingRepository struct {
	cache cache.RankingCache

	// 下面是给 v1 用的
	redisCache *cache.RedisRankingCache
	localCache *cache.LocalRankingCache
}

func (repo *CachedRankingRepository) GetTopN(ctx context.Context) ([]domain.Article, error) {
	res, err := repo.localCache.Get(ctx)
	if err == nil {
		return res, nil
	}
	res, err = repo.redisCache.Get(ctx)
	if err != nil {
		return repo.localCache.ForceGet(ctx)
	}
	_ = repo.localCache.Set(ctx, res)
	return res, nil
}
func (repo *CachedRankingRepository) ReplaceTopN(ctx context.Context, arts []domain.Article) error {
	_ = repo.localCache.Set(ctx, arts)
	return repo.redisCache.Set(ctx, arts)
}

/*
func NewCachedRankingRepository(cache cache.RankingCache) RankingRepository {
	return &CachedRankingRepository{cache: cache}
}

func (repo *CachedRankingRepository) GetTopN(ctx context.Context) ([]domain.Article, error) {
	return repo.cache.Get(ctx)
}

func (repo *CachedRankingRepository) ReplaceTopN(ctx context.Context, arts []domain.Article) error {
	return repo.cache.Set(ctx, arts)
}
*/
