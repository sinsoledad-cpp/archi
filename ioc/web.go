package ioc

import (
	"archi/internal/web"
	"archi/internal/web/middleware"
	"archi/internal/web/middleware/jwt"
	"archi/pkg/ginx"
	mdl "archi/pkg/ginx/middleware"
	"archi/pkg/logger"
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"time"
)

func InitWebEngine(middlewares []gin.HandlerFunc, l logger.Logger, userHdl *web.UserHandler) *gin.Engine {
	ginx.SetLogger(l)
	gin.ForceConsoleColor()
	engine := gin.Default()
	engine.Static("/uploads", "./uploads")
	engine.Use(middlewares...)
	userHdl.RegisterRoutes(engine)
	//wechatHdl.RegisterRoutes(engine)//, wechatHdl *web.OAuth2WechatHandler
	return engine
}

func InitGinMiddlewares(jwtHdl jwt.Handler, l logger.Logger) []gin.HandlerFunc {
	corsMiddleware := cors.New(cors.Config{
		// 在生产环境中，您应该将 AllowAllOrigins 设置为 false，并具体指定允许的前端域名
		// 例如: AllowOrigins: []string{"http://your-frontend.com"},
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"X-Jwt-Token", "X-Refresh-Token"}, // 允许前端访问后端设置的响应头
		AllowCredentials: true,                                       // 允许携带 Cookie
		MaxAge:           12 * time.Hour,                             // preflight 请求的缓存时间
	})

	logFn := func(ctx context.Context, al mdl.AccessLog) {
		fields := []logger.Field{
			logger.String("path", al.Path),
			logger.String("method", al.Method),
			logger.String("req_body", al.ReqBody),
			logger.Int("status", al.Status),
			logger.String("resp_body", al.RespBody),
			logger.Int64("duration_ms", al.Duration.Milliseconds()),
		}
		l.Info("access log ", fields...)
	}
	accessLogMiddleware := mdl.NewAccessLogBuilder(logFn).AllowReqBody().AllowRespBody().Build()

	return []gin.HandlerFunc{
		middleware.NewJWTAuth(jwtHdl).Middleware(),
		corsMiddleware,
		accessLogMiddleware,
	}
}
