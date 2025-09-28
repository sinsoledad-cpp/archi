package cache

import (
	"archi/internal/domain"
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

// ErrKeyNotExist 因为我们目前还是只有一个实现，所以可以保持用别名
var ErrKeyNotExist = redis.Nil

//go:generate mockgen -source=./user.go -package=mocks -destination=mocks/user.mock.go UserCache
type UserCache interface {
	Delete(ctx context.Context, id int64) error
	Get(ctx context.Context, id int64) (domain.User, error)
	Set(ctx context.Context, u domain.User) error
}

type RedisUserCache struct {
	cmd redis.Cmdable
	// 过期时间
	expiration time.Duration
}

func NewRedisUserCache(cmd redis.Cmdable) UserCache {
	return &RedisUserCache{
		cmd:        cmd,
		expiration: time.Minute * 15,
	}
}
func (r *RedisUserCache) Delete(ctx context.Context, id int64) error {
	return r.cmd.Del(ctx, r.key(id)).Err()
}

func (r *RedisUserCache) Get(ctx context.Context, id int64) (domain.User, error) {
	key := r.key(id)
	data, err := r.cmd.Get(ctx, key).Result()
	if err != nil {
		return domain.User{}, err
	}
	// 反序列化回来
	var u domain.User
	err = json.Unmarshal([]byte(data), &u)
	return u, err
}

func (r *RedisUserCache) Set(ctx context.Context, u domain.User) error {
	data, err := json.Marshal(u)
	if err != nil {
		return err
	}
	key := r.key(u.ID)
	return r.cmd.Set(ctx, key, data, r.expiration).Err()
}

func (r *RedisUserCache) key(id int64) string {
	return fmt.Sprintf("user:info:%d", id)
}
