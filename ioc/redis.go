package ioc

import (
	"archi/pkg/redisx/metrics"

	rlock "github.com/gotomicro/redis-lock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

func InitRedis() redis.Cmdable {
	type redisConfig struct {
		Addr string `yaml:"addr"`
	}
	var cfg redisConfig = redisConfig{
		Addr: "127.0.0.1:6379",
	}
	if err := viper.UnmarshalKey("redis", &cfg); err != nil {
		panic(err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.Addr,
	})

	opts := prometheus.SummaryOpts{
		Namespace: "sinsoledad",
		Subsystem: "archi",
		Name:      "redis_cmd_duration_ms", // 指标名
		Help:      "统计 Redis 命令操作 (毫秒)",
		ConstLabels: map[string]string{
			"instance_id": "my_redis_instance", // 示例
		},
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.75:  0.01,
			0.9:   0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	}
	hook := metrics.NewPrometheusHook(opts)
	rdb.AddHook(hook)
	return rdb
}
func InitRlockClient(client redis.Cmdable) *rlock.Client {
	return rlock.NewClient(client)
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
