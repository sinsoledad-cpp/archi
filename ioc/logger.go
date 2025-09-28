package ioc

import (
	"archi/pkg/logger"
)

func InitLogger() logger.Logger {
	//type LogConfig struct {
	//	Level      string `mapstructure:"level"`
	//	Filename   string `mapstructure:"filename"`
	//	MaxSize    int    `mapstructure:"max_size"`
	//	MaxAge     int    `mapstructure:"max_age"`
	//	MaxBackups int    `mapstructure:"max_backups"`
	//}
	//var customCfg LogConfig
	//viper.UnmarshalKey("log", &customCfg)

	//// 这里我们用一个小技巧，
	//// 就是直接使用 zap 本身的配置结构体来处理
	//cfg := zap.NewDevelopmentConfig()
	////err := viper.UnmarshalKey("log", &cfg)
	////if err != nil {
	////	panic(err)
	////}
	//l, err := cfg.Build()
	//if err != nil {
	//	panic(err)
	//}

	if err := InitZap("debug"); err != nil {
		panic(err)
	}

	return logger.NewZapLogger(lg)
}
