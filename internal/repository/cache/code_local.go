package cache

import (
	"context"
	"errors"
	"fmt"
	lru "github.com/hashicorp/golang-lru" // LRU 是 "Least Recently Used" 的缩写，意为“最近最少使用”。
	"sync"
	"time"
)

type codeItem struct {
	code string
	// 可验证次数
	cnt int
	// 过期时间
	expire time.Time
}

// 技术选型考虑的点
//  1. 功能性：功能是否能够完全覆盖你的需求。
//  2. 社区和支持度：社区是否活跃，文档是否齐全，
//     以及百度（搜索引擎）能不能搜索到你需要的各种信息，有没有帮你踩过坑
//  3. 非功能性：易用性（用户友好度，学习曲线要平滑），
//     扩展性（如果开源软件的某些功能需要定制，框架是否支持定制，以及定制的难度高不高）
//     性能（追求性能的公司，往往有能力自研）

// LocalCodeCache 本地缓存实现
type LocalCodeCache struct {
	cache *lru.Cache
	// 普通锁，或者说写锁
	lock sync.Mutex
	// 读写锁
	expiration time.Duration
}

func NewLocalCodeCache(c *lru.Cache, expiration time.Duration) *LocalCodeCache {
	return &LocalCodeCache{
		cache:      c,
		expiration: expiration,
	}
}
func (l *LocalCodeCache) Set(ctx context.Context, biz string, phone string, code string) error {

	l.lock.Lock()
	defer l.lock.Unlock()
	// 这里可以考虑用读写锁来优化，但是效果不会很好
	// 因为你可以预期，大部分时候是要走到写锁里面的

	// 我选用的本地缓存，很不幸的是，没有获得过期时间的接口，所以都是自己维持了一个过期时间字段
	key := l.key(biz, phone)

	now := time.Now()
	val, ok := l.cache.Get(key)
	if !ok {
		// 说明没有验证码
		l.cache.Add(key, codeItem{
			code:   code,
			cnt:    3,
			expire: now.Add(l.expiration),
		})
		return nil
	}
	itm, ok := val.(codeItem)
	if !ok {
		// 理论上来说这是不可能的
		return errors.New("系统错误")
	}
	if itm.expire.Sub(now) > time.Minute*9 {
		// 不到一分钟
		return ErrCodeSendTooMany
	}
	// 重发
	l.cache.Add(key, codeItem{
		code:   code,
		cnt:    3,
		expire: now.Add(l.expiration),
	})
	return nil
}

func (l *LocalCodeCache) Verify(ctx context.Context, biz string, phone string, inputCode string) (bool, error) {
	l.lock.Lock()
	defer l.lock.Unlock()
	key := l.key(biz, phone)
	val, ok := l.cache.Get(key)
	if !ok {
		// 都没发验证码
		return false, ErrKeyNotExist
	}
	itm, ok := val.(codeItem)
	if !ok {
		// 理论上来说这是不可能的
		return false, errors.New("系统错误")
	}
	if itm.cnt <= 0 {
		return false, ErrCodeVerifyTooMany
	}
	itm.cnt--
	return itm.code == inputCode, nil
}

func (l *LocalCodeCache) key(biz string, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}
