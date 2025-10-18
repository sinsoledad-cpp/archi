package ioc

import (
	"archi/internal/repository/dao"
	"archi/pkg/logger"
	"fmt"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	glogger "gorm.io/gorm/logger"
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
