package ioc

import (
	"github.com/natefinch/lumberjack"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

// lg 是全局的 zap.Logger 实例，用于记录日志
var lg *zap.Logger

// logConf 配置信息
type logConf struct {
	Level      string `mapstructure:"level"`
	Path       string `mapstructure:"path"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"` // 最多保留的备份文件数量
}

func InitZap(mode string) (err error) {
	var cfg = logConf{
		Level:      "debug",
		Path:       "logs/app.log",
		MaxSize:    128,
		MaxAge:     7,
		MaxBackups: 3,
	}
	if err := viper.UnmarshalKey("log", &cfg); err != nil {
		panic(err)
	}
	writeSyncer := getLogWriter(cfg.Path, cfg.MaxSize, cfg.MaxBackups, cfg.MaxAge) // 获取日志写入器，支持日志轮转
	encoder := getEncoder()                                                        // 获取日志编码器，定义日志的输出格式
	var l = new(zapcore.Level)                                                     // 创建一个 zapcore.Level 实例，并根据配置解析日志级别
	if err = l.UnmarshalText([]byte(cfg.Level)); err != nil {
		return
	}
	var core zapcore.Core
	if mode == "debug" {
		core = zapcore.NewTee(
			zapcore.NewCore(getConsoleEncoder(), zapcore.Lock(os.Stdout), zapcore.DebugLevel),
			zapcore.NewCore(encoder, writeSyncer, l),
		)
	} else {
		core = zapcore.NewCore(encoder, writeSyncer, l) // 生产模式下，只将日志写入文件
	}
	lg = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1)) // 创建一个新的 zap.Logger 实例，并添加调用者信息
	zap.ReplaceGlobals(lg)                                    // 替换全局 logger 实例
	zap.L().Info("init logger success!")                      // 记录初始化成功的日志信息
	return
}

// getEncoder 配置并返回一个 zapcore.Encoder 实例，用于定义日志的编码格式。该编码器使用 JSON 格式输出日志，并对时间、级别、持续时间和调用者信息进行自定义编码。
func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()             // 使用 zap 的生产环境编码配置作为基础配置
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder         // 设置时间编码器为 ISO8601 格式，例如 "2006-01-02T15:04:05.000Z0700"
	encoderConfig.TimeKey = "time"                                // 设置时间字段的键名为 "time"
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder       // 设置日志级别的编码器为大写形式，例如 "INFO", "ERROR"
	encoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder // 设置持续时间的编码器为秒格式
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder       // 设置调用者信息的编码器为短格式，仅包含文件名和行号
	return zapcore.NewJSONEncoder(encoderConfig)
}
func getConsoleEncoder() zapcore.Encoder {
	encoderConfig := zap.NewDevelopmentEncoderConfig()            // 使用 zap 提供的开发环境配置作为基础
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder         // 设置时间编码器为 ISO8601 格式
	encoderConfig.TimeKey = "time"                                // 设置时间字段的键名为 "time"
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder  // 【关键】设置日志级别的编码器为大写带颜色
	encoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder // 设置持续时间的编码器为秒格式
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder       // 设置调用者信息的编码器为短格式，仅包含文件名和行号
	return zapcore.NewConsoleEncoder(encoderConfig)               // 返回一个 ConsoleEncoder，它能够正确处理颜色代码并以人类可读的格式输出
}

// getLogWriter 创建并返回一个 zapcore.WriteSyncer 实例，用于日志轮转写入
func getLogWriter(path string, maxSize, maxBackups, maxAge int) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   path,       // 日志文件路径
		MaxSize:    maxSize,    // 单个日志文件最大大小(MB)
		MaxBackups: maxBackups, // 最多保留备份数量
		MaxAge:     maxAge,     // 最大保留天数
		LocalTime:  true,       // 使用本地时间格式
		Compress:   false,      // 不压缩旧日志文件
	}
	// 将 lumberjack.Logger 包装为 zapcore.WriteSyncer 并返回
	return zapcore.AddSync(lumberJackLogger)
}
