package metrics

import (
	"context"
	"errors"
	"net"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
)

type PrometheusHook struct {
	vector *prometheus.SummaryVec
}

// NewPrometheusHook 创建一个 prometheus hook	Namespace、Subsystem、Name、Help
func NewPrometheusHook(opt prometheus.SummaryOpts) *PrometheusHook {
	vec := prometheus.NewSummaryVec(opt, []string{"cmd", "key_exist"})
	prometheus.MustRegister(vec)
	return &PrometheusHook{
		vector: vec,
	}
}

// DialHook 创建连接钩子
func (p *PrometheusHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return next(ctx, network, addr)
	}
}

// ProcessHook 处理命令钩子
func (p *PrometheusHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		start := time.Now()
		var err error
		defer func() {
			duration := time.Since(start).Milliseconds()
			keyExists := !errors.Is(redis.Nil, err)
			p.vector.WithLabelValues(cmd.Name(), strconv.FormatBool(keyExists)).
				Observe(float64(duration))
		}()
		err = next(ctx, cmd)
		return err
	}
}

// ProcessPipelineHook 处理管道命令钩子
func (p *PrometheusHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		return next(ctx, cmds)
	}
}
