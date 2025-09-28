package ioc

import (
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

func InitRedis() redis.Cmdable {
	// 这个是假设你有一个独立的 Redis 的配置文件
	return redis.NewClient(&redis.Options{
		Addr: viper.GetString("redis.addr"),
	})
}

/*
var client *redis.Client

// Init 初始化连接
func Init(cfg *conf.RedisConf) {
	client = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,           // 0 表示使用默认数据库
		PoolSize:     cfg.PoolSize,     // 连接池大小
		MinIdleConns: cfg.MaxIdleConns, // 最小空闲连接数
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		zap.L().Error("connect _redis failed", zap.Error(err))
	}
	return
}

func Close() {
	_ = client.Close()
}
*/
