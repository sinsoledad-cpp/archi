package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"gorm.io/gorm"
)

type PrometheusCallbacks struct {
	vector *prometheus.SummaryVec
}

func NewPrometheusCallbacks(opts prometheus.SummaryOpts) *PrometheusCallbacks {
	vector := prometheus.NewSummaryVec(opts, []string{"type", "table"})
	prometheus.MustRegister(vector)
	return &PrometheusCallbacks{
		vector: vector,
	}
}
func (c *PrometheusCallbacks) Name() string {
	return "prometheus"
}

func (c *PrometheusCallbacks) Initialize(db *gorm.DB) error {

	// Create
	if err := db.Callback().Create().Before("*").
		Register("prometheus_create_before", c.Before()); err != nil {
		return err
	}
	if err := db.Callback().Create().After("*").
		Register("prometheus_create_after", c.After("CREATE")); err != nil {
		return err
	}

	// Query
	if err := db.Callback().Query().Before("*").
		Register("prometheus_query_before", c.Before()); err != nil {
		return err
	}
	if err := db.Callback().Query().After("*").
		Register("prometheus_query_after", c.After("QUERY")); err != nil {
		return err
	}

	// Update
	if err := db.Callback().Update().Before("*").
		Register("prometheus_update_before", c.Before()); err != nil {
		return err
	}
	if err := db.Callback().Update().After("*").
		Register("prometheus_update_after", c.After("UPDATE")); err != nil {
		return err
	}

	// Delete
	if err := db.Callback().Delete().Before("*").
		Register("prometheus_delete_before", c.Before()); err != nil {
		return err
	}
	if err := db.Callback().Delete().After("*").
		Register("prometheus_delete_after", c.After("DELETE")); err != nil {
		return err
	}

	// Raw
	if err := db.Callback().Raw().Before("*").
		Register("prometheus_raw_before", c.Before()); err != nil {
		return err
	}
	if err := db.Callback().Raw().After("*").
		Register("prometheus_raw_after", c.After("RAW")); err != nil {
		return err
	}

	// Row
	if err := db.Callback().Row().Before("*").
		Register("prometheus_row_before", c.Before()); err != nil {
		return err
	}
	if err := db.Callback().Row().After("*").
		Register("prometheus_row_after", c.After("ROW")); err != nil {
		return err
	}

	return nil
}

func (c *PrometheusCallbacks) Before() func(db *gorm.DB) {
	return func(db *gorm.DB) {
		start := time.Now()
		db.Set("start_time", start)
	}
}

func (c *PrometheusCallbacks) After(typ string) func(db *gorm.DB) {
	return func(db *gorm.DB) {
		val, _ := db.Get("start_time")
		start, ok := val.(time.Time)
		if ok {
			duration := time.Since(start).Milliseconds()
			c.vector.WithLabelValues(typ, db.Statement.Table).
				Observe(float64(duration))
		}
	}
}
