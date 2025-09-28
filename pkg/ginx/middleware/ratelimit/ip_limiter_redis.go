package ratelimit

import (
	"archi/pkg/logger"
	_ "embed"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"net/http"
	"time"
)

var log = logger.NewNopLogger()

func SetLogger(l logger.Logger) {
	log = l
}

//go:embed slide_window.lua
var luaScript string

type RedisIpLimiter struct {
	prefix   string
	cmd      redis.Cmdable
	interval time.Duration
	rate     int // 阈值
}

func NewRedisIpLimiter(cmd redis.Cmdable, interval time.Duration, rate int) *RedisIpLimiter {
	return &RedisIpLimiter{
		cmd:      cmd,
		prefix:   "ip-limiter",
		interval: interval,
		rate:     rate,
	}
}

func (r *RedisIpLimiter) Prefix(prefix string) *RedisIpLimiter {
	r.prefix = prefix
	return r
}
func (r *RedisIpLimiter) limit(ctx *gin.Context) (bool, error) {
	key := fmt.Sprintf("%s:%s", r.prefix, ctx.ClientIP())
	return r.cmd.Eval(ctx, luaScript, []string{key}, r.interval.Milliseconds(), r.rate, time.Now().UnixMilli()).Bool()
}

func (r *RedisIpLimiter) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		limited, err := r.limit(ctx)
		if err != nil {
			// 这一步很有意思，就是如果这边出错了
			// 要怎么办？
			log.Error("ip限流失败", logger.Error(err))
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		if limited {
			log.Warn("ip限流", logger.String("ip", ctx.ClientIP()))
			ctx.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		ctx.Next()
	}
}
