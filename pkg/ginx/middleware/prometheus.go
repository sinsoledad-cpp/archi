package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

// PrometheusBuilder 主要是统计响应时间
type PrometheusBuilder struct {
	// 除了一个 Name 是必选的，其它都是可选的
	// 如果暴露 New 方法，那么就需要考虑暴露其他的方法来允许用户配置 Namespace 等
	// 所以我直接做成了公开字段
	Namespace string
	Subsystem string
	Name      string
	Help      string
	// 这一个实例名字，你可以考虑使用 本地 IP，
	// 又或者在启动的时候配置一个 ID
	InstanceID string
}

func NewPrometheusBuilder(namespace, subsystem, name, help string) *PrometheusBuilder {
	return &PrometheusBuilder{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      name,
		Help:      help,
	}
}

func (p *PrometheusBuilder) BuildResponseTime() gin.HandlerFunc {
	// pattern 是命中路由
	labels := []string{"method", "pattern", "status"}
	vector := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: p.Namespace,
		Subsystem: p.Subsystem,
		Name:      p.Name + "_resp_time",
		Help:      p.Help,
		ConstLabels: map[string]string{
			"instance_id": p.InstanceID,
		},
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.75:  0.01,
			0.90:  0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	}, labels)
	prometheus.MustRegister(vector)
	return func(ctx *gin.Context) {
		method := ctx.Request.Method
		start := time.Now()
		defer func() {
			// 最后我们再来统计一下
			vector.WithLabelValues(method, ctx.FullPath(),
				strconv.Itoa(ctx.Writer.Status())).
				Observe(float64(time.Since(start).Milliseconds())) // 执行时间
		}()
		ctx.Next()
	}
}

func (p *PrometheusBuilder) BuildActiveRequest() gin.HandlerFunc {
	// 一般我们只关心总的活跃请求数
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: p.Namespace,
		Subsystem: p.Subsystem,
		Name:      p.Name + "_active_req",
		Help:      p.Help,
		ConstLabels: map[string]string{
			"instance_id": p.InstanceID,
		},
	})
	prometheus.MustRegister(gauge)
	return func(ctx *gin.Context) {
		gauge.Inc()
		defer gauge.Dec()
		ctx.Next()
	}
}

func (p *PrometheusBuilder) BuildPerRouteActiveRequest() gin.HandlerFunc {
	// 1. 定义标签
	labels := []string{"method", "pattern"}

	// 2. 使用 NewGaugeVec 而不是 NewGauge
	gaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: p.Namespace,
		Subsystem: p.Subsystem,
		Name:      p.Name + "_active_req", // 指标名可以保持一致
		Help:      p.Help,
		ConstLabels: map[string]string{
			"instance_id": p.InstanceID,
		},
	}, labels) // 3. 传入标签定义

	prometheus.MustRegister(gaugeVec)

	return func(ctx *gin.Context) {
		// 4. 使用 WithLabelValues 来获取特定路由的 Gauge
		gauge := gaugeVec.WithLabelValues(ctx.Request.Method, ctx.FullPath())

		gauge.Inc()
		defer gauge.Dec()
		ctx.Next()
	}
}
