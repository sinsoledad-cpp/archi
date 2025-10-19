package setting

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func InitPrometheus() {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		// 监听 8081 端口，你也可以做成可配置的
		err := http.ListenAndServe(":8081", nil)
		if err != nil {
			return
		}
	}()
}
