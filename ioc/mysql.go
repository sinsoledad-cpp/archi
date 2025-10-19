package ioc

import (
	"archi/internal/repository/dao"
	"archi/pkg/gormx/metrics"
	"archi/pkg/logger"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
	"gorm.io/plugin/opentelemetry/tracing"
	gormprometheus "gorm.io/plugin/prometheus"
)

func InitMySQL(l logger.Logger) *gorm.DB {
	type mysqlConfig struct {
		DSN string `yaml:"dsn"`
	}
	var cfg mysqlConfig = mysqlConfig{
		DSN: "root:root@tcp(localhost:3306)/archi?charset=utf8mb4&parseTime=True&loc=Local",
	}
	if err := viper.UnmarshalKey("mysql", &cfg); err != nil {
		panic(err)
	}
	// 连接数据库，开启慢日志
	db, err := gorm.Open(mysql.Open(cfg.DSN),
		&gorm.Config{
			Logger: glogger.New(gormLoggerFunc(l.Debug), glogger.Config{
				// 慢查询
				SlowThreshold: 0,
				LogLevel:      glogger.Info,
			}),
		},
	)
	if err != nil {
		panic(err)
	}

	// 对数据库状态监控
	err = db.Use(gormprometheus.New(gormprometheus.Config{
		DBName:          "archi",
		RefreshInterval: 60,
		MetricsCollector: []gormprometheus.MetricsCollector{
			&gormprometheus.MySQL{
				// 指定需要监控的变量,如果没定义,则默认所有变量
				VariableNames: []string{"Threads_running"},
			},
		},
	}))
	if err != nil {
		panic(err)
	}

	// 添加 Prometheus 监控,对数据库操作
	cb := metrics.NewPrometheusCallbacks(prometheus.SummaryOpts{
		Namespace: "sinsoledad",
		Subsystem: "archi",
		Name:      "gorm_db",
		Help:      "统计 GORM 的数据库操作(毫秒)",
		ConstLabels: map[string]string{
			"instance_id": "my_mysql_instance",
		},
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.75:  0.01,
			0.9:   0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	})
	if err = db.Use(cb); err != nil {
		panic(err)
	}

	// OTEL 不要收集指标 (Metrics)
	if err = db.Use(tracing.NewPlugin(tracing.WithoutMetrics())); err != nil {
		panic(err)
	}

	if err = dao.InitTables(db); err != nil {
		panic(err)
	}
	return db
}

type gormLoggerFunc func(msg string, fields ...logger.Field)

func (g gormLoggerFunc) Printf(s string, i ...interface{}) {
	msg := fmt.Sprintf(s, i...) // 手动格式化日志
	g("GORM SQL日志：" + msg)      // 加个前缀
}

//func (g gormLoggerFunc) Printf(s string, i ...interface{}) {
//	g(s, logger.Field{Key: "args", Val: i})
//
//}

/*
var database *gorm.DB
// Init 初始化数据库
func Init(cfg *conf.MySQLConf) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.Timeout)
	var err error
	database, err = gorm.Open(mysql.Open(dsn), &gorm.mysqlConfig{
		SkipDefaultTransaction: true, // 禁用默认事务
		NamingStrategy: schema.NamingStrategy{
			SingularTable: false, // 禁用单数表名
			NoLowerCase:   false, // 禁用小写表名
		},
	})
	if err != nil {
		zap.L().Error("connect mysql failed", zap.Error(err))
		return
	}
	sqlDB, err := database.DB()
	if err != nil {
		zap.L().Error("get _mysql db failed", zap.Error(err))
		return
	}
	err = sqlDB.Ping()
	if err != nil {
		zap.L().Error("ping _mysql failed", zap.Error(err))
		return
	}
	// 设置连接池
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	return
}
*/
