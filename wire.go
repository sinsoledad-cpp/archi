//go:build wireinject

package main

import (
	"archi/internal/repository"
	"archi/internal/repository/cache"
	"archi/internal/repository/dao"
	"archi/internal/service"
	"archi/internal/web"
	"archi/internal/web/middleware/jwt"
	"archi/ioc"
	"github.com/google/wire"
)

var thirdParty = wire.NewSet(
	ioc.InitLogger,
	ioc.InitMySQL,
	ioc.InitRedis,
)

var userSvc = wire.NewSet(
	cache.NewRedisUserCache,
	dao.NewGORMUserDAO,
	repository.NewCachedUserRepository,
	service.NewUserService,
)

var codeSvc = wire.NewSet(
	cache.NewRedisCodeCache,
	repository.NewCachedCodeRepository,
	ioc.InitSMSService,
	service.NewCodeService,
)

//var wechatSvc = wire.NewSet(
//	ioc.InitWechatService,
//)

func InitApp() *App {
	wire.Build(
		thirdParty,

		userSvc,
		codeSvc,
		//wechatSvc,

		jwt.NewRedisJWTHandler,
		web.NewUserHandler,
		//web.NewOAuth2WechatHandler,

		ioc.InitWebEngine,
		ioc.InitGinMiddlewares,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
