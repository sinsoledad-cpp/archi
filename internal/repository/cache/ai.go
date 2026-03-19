package cache

import (
	"archi/internal/domain"
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

type AiCache interface {
	GetArticleSummary(ctx context.Context, artId int64) (domain.ArticleSummary, error)
	SetArticleSummary(ctx context.Context, artId int64, summary domain.ArticleSummary) error
}

type RedisAiCache struct {
	client redis.Cmdable
}

func NewRedisAiCache(client redis.Cmdable) AiCache {
	return &RedisAiCache{
		client: client,
	}
}

func (r *RedisAiCache) GetArticleSummary(ctx context.Context, artId int64) (domain.ArticleSummary, error) {
	key := r.summaryKey(artId)
	val, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		return domain.ArticleSummary{}, err
	}
	var res domain.ArticleSummary
	err = json.Unmarshal(val, &res)
	return res, err
}

func (r *RedisAiCache) SetArticleSummary(ctx context.Context, artId int64, summary domain.ArticleSummary) error {
	key := r.summaryKey(artId)
	val, err := json.Marshal(summary)
	if err != nil {
		return err
	}
	// 缓存 7 天
	return r.client.Set(ctx, key, val, time.Hour*24*7).Err()
}

func (r *RedisAiCache) summaryKey(artId int64) string {
	return fmt.Sprintf("ai:article_summary:%d", artId)
}
