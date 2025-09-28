package limiter

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

//go:embed lua/slide_window.lua
var luaSlideWindow string

// 将字符串脚本编译成 redis.Script 对象，全局唯一
var slideWindowScript = redis.NewScript(luaSlideWindow)

// RedisSlideWindowLimiter 是一个基于Redis滑动窗口算法的限流器。
type RedisSlideWindowLimiter struct {
	cmd      redis.Cmdable
	interval time.Duration // 窗口大小
	rate     int           // 阈值
	// interval 内允许 rate 个请求
	// 1s 内允许 3000 个请求
}

// NewRedisSlideWindowLimiter 创建一个新的 RedisSlideWindowLimiter 实例
func NewRedisSlideWindowLimiter(cmd redis.Cmdable, interval time.Duration, rate int) Limiter {
	return &RedisSlideWindowLimiter{
		cmd:      cmd,
		interval: interval,
		rate:     rate,
	}
}

// Limit 实现了 Limiter 接口。它使用Redis的Lua脚本来判断给定的key是否应被限流。
func (r *RedisSlideWindowLimiter) Limit(ctx context.Context, key string) (bool, error) {
	// 使用 slideWindowScript.Run 代替 r.cmd.Eval
	res, err := slideWindowScript.Run(ctx, r.cmd, []string{key},
		r.interval.Milliseconds(), r.rate, time.Now().UnixMilli()).Result()

	if err != nil {
		// 当脚本返回的不是整数时，go-redis的Bool()会报错，但这里Lua脚本返回的是0或1
		// 为了健壮性，我们最好直接处理Result()的返回值
		// 如果是redis.Nil，说明key不存在，未被限流
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}

	// Lua脚本通常返回数字 1 表示被限流，0 表示通过
	isLimited, ok := res.(int64)
	if !ok {
		// 如果返回的不是预期的整数，最好记录一个错误
		return false, fmt.Errorf("unexpected result type from redis script: %T", res)
	}

	return isLimited == 1, nil
}

//
//	func (r *RedisSlideWindowLimiter) Limit(ctx context.Context, key string) (bool, error) {
//		return r.cmd.Eval(ctx, luaSlideWindow, []string{key},
//			r.interval.Milliseconds(), r.rate, time.Now().UnixMilli()).Bool()
//	}
