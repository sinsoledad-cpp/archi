package logger

type Logger interface {
	Debug(msg string, args ...Field)
	Info(msg string, args ...Field)
	Warn(msg string, args ...Field)
	Error(msg string, args ...Field)
}

type Field struct {
	Key string
	Val any
}

/*

// lg 是全局的 zap.Logger 实例，用于记录日志
var lg *zap.Logger

func Init(cfg *conf.LogConf, mode string) (err error) {
	// 获取日志写入器，支持日志轮转
	writeSyncer := getLogWriter(cfg.Path, cfg.MaxSize, cfg.MaxBackups, cfg.MaxAge)
	// 获取日志编码器，定义日志的输出格式
	encoder := getEncoder()
	// 创建一个 zapcore.Level 实例，并根据配置解析日志级别
	var l = new(zapcore.Level)
	err = l.UnmarshalText([]byte(cfg.Level))
	if err != nil {
		return
	}
	var core zapcore.Core
	if mode == "debug" {
		// 开发模式下，使用控制台编码器将日志同时输出到标准输出和文件
		consoleEncode := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
		core = zapcore.NewTee(
			zapcore.NewCore(consoleEncode, zapcore.Lock(os.Stdout), zapcore.DebugLevel),
			zapcore.NewCore(encoder, writeSyncer, l),
		)
	} else {
		// 生产模式下，只将日志写入文件
		core = zapcore.NewCore(encoder, writeSyncer, l)
	}

	// 创建一个新的 zap.Logger 实例，并添加调用者信息
	lg = zap.New(core, zap.AddCaller())
	// 替换全局 logger 实例
	zap.ReplaceGlobals(lg)
	// 记录初始化成功的日志信息
	zap.L().Info("init logger success!")
	return
}

// getEncoder 配置并返回一个 zapcore.Encoder 实例，用于定义日志的编码格式。
// 该编码器使用 JSON 格式输出日志，并对时间、级别、持续时间和调用者信息进行自定义编码。
func getEncoder() zapcore.Encoder {
	// 使用 zap 的生产环境编码配置作为基础配置
	encoderConfig := zap.NewProductionEncoderConfig()
	// 设置时间编码器为 ISO8601 格式，例如 "2006-01-02T15:04:05.000Z0700"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	// 设置时间字段的键名为 "time"
	encoderConfig.TimeKey = "time"
	// 设置日志级别的编码器为大写形式，例如 "INFO", "ERROR"
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	// 设置持续时间的编码器为秒格式
	encoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder
	// 设置调用者信息的编码器为短格式，仅包含文件名和行号
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	return zapcore.NewJSONEncoder(encoderConfig)
}

// getLogWriter 创建并返回一个 zapcore.WriteSyncer 实例，用于日志轮转写入
// 参数:
//   - path: 日志文件路径
//   - maxSize: 每个日志文件的最大尺寸(MB)
//   - maxBackups: 保留旧日志文件的最大数量
//   - maxAge: 保留旧日志文件的最大天数
//
// 返回值:
//   - zapcore.WriteSyncer: 包装了 lumberjack.Logger 的写入同步器
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

// GinLogger 接收gin框架默认的日志
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		c.Next()

		cost := time.Since(start)
		lg.Info(path,
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()),
			zap.Duration("cost", cost),
		)
	}
}

// GinRecovery recover掉项目可能出现的panic，并使用zap记录相关日志
func GinRecovery(stack bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				if brokenPipe {
					lg.Error(c.Request.URL.Path,
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
					// If the connection is dead, we can't write a status to it.
					c.Error(err.(error)) // nolint: errcheck
					c.Abort()
					return
				}

				if stack {
					lg.Error("[Recovery from panic]",
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
						zap.String("stack", string(debug.Stack())),
					)
				} else {
					lg.Error("[Recovery from panic]",
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
				}
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}
*/
