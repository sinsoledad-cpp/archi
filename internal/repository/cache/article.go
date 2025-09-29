package cache

import (
	"archi/internal/domain"
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

//go:generate mockgen -source=./article.go -package=mocks -destination=./mocks/article.mock.go ArticleCache
type ArticleCache interface {
	GetFirstPage(ctx context.Context, uid int64) ([]domain.Article, error)
	SetFirstPage(ctx context.Context, uid int64, res []domain.Article) error
	DelFirstPage(ctx context.Context, uid int64) error
	Get(ctx context.Context, id int64) (domain.Article, error)
	Set(ctx context.Context, art domain.Article) error
	GetPub(ctx context.Context, id int64) (domain.Article, error)
	SetPub(ctx context.Context, res domain.Article) error
}

type RedisArticleCache struct {
	client redis.Cmdable
}

func NewRedisArticleCache(client redis.Cmdable) ArticleCache {
	return &RedisArticleCache{
		client: client,
	}
}

func (r *RedisArticleCache) GetFirstPage(ctx context.Context, uid int64) ([]domain.Article, error) {
	key := r.firstKey(uid)
	//val, err := r.client.Get(ctx, firstKey).Result()
	val, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	var res []domain.Article
	err = json.Unmarshal(val, &res)
	return res, err
}

func (r *RedisArticleCache) SetFirstPage(ctx context.Context, uid int64, arts []domain.Article) error {
	for i := 0; i < len(arts); i++ {
		arts[i].Content = arts[i].Abstract()
	}
	key := r.firstKey(uid)
	val, err := json.Marshal(arts)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, val, time.Minute*10).Err()
}

func (r *RedisArticleCache) DelFirstPage(ctx context.Context, uid int64) error {
	return r.client.Del(ctx, r.firstKey(uid)).Err()
}

func (r *RedisArticleCache) firstKey(uid int64) string {
	return fmt.Sprintf("article:first_page:%d", uid)
}

func (r *RedisArticleCache) Get(ctx context.Context, id int64) (domain.Article, error) {
	val, err := r.client.Get(ctx, r.key(id)).Bytes()
	if err != nil {
		return domain.Article{}, err
	}
	var res domain.Article
	err = json.Unmarshal(val, &res)
	return res, err
}

func (r *RedisArticleCache) Set(ctx context.Context, art domain.Article) error {
	val, err := json.Marshal(art)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key(art.ID), val, time.Minute*10).Err()
}
func (r *RedisArticleCache) key(id int64) string {
	return fmt.Sprintf("article:detail:%d", id)
}

func (r *RedisArticleCache) GetPub(ctx context.Context, id int64) (domain.Article, error) {
	val, err := r.client.Get(ctx, r.pubKey(id)).Bytes()
	if err != nil {
		return domain.Article{}, err
	}
	var res domain.Article
	err = json.Unmarshal(val, &res)
	return res, err
}

func (r *RedisArticleCache) SetPub(ctx context.Context, art domain.Article) error {
	val, err := json.Marshal(art)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.pubKey(art.ID), val, time.Minute*10).Err()
}

func (r *RedisArticleCache) pubKey(id int64) string {
	return fmt.Sprintf("article:pub:detail:%d", id)
}
