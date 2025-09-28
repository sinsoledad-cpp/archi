package repository

import (
	"archi/internal/repository/cache"
	"context"
)

var ErrCodeVerifyTooMany = cache.ErrCodeVerifyTooMany
var ErrCodeSendTooMany = cache.ErrCodeSendTooMany
var ErrCodeExpired = cache.ErrCodeExpired

//go:generate mockgen -source=./code.go -package=repomocks -destination=./mocks/code.mock.go CodeRepository
type CodeRepository interface {
	Set(ctx context.Context, biz, phone, code string) error
	Verify(ctx context.Context, biz, phone, code string) (bool, error)
}

type CachedCodeRepository struct {
	cache cache.CodeCache
}

func NewCachedCodeRepository(c cache.CodeCache) CodeRepository {
	return &CachedCodeRepository{
		cache: c,
	}
}

func (c *CachedCodeRepository) Set(ctx context.Context, biz, phone, code string) error {
	return c.cache.Set(ctx, biz, phone, code)
}

func (c *CachedCodeRepository) Verify(ctx context.Context, biz, phone, code string) (bool, error) {
	return c.cache.Verify(ctx, biz, phone, code)
}
