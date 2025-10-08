//go:build wireinject

package main

import (
	"archi/internal/event/article"
	"archi/internal/repository"
	"archi/internal/repository/cache"
	"archi/internal/repository/dao"
	searchDAO "archi/internal/repository/dao/search"
	searchRepo "archi/internal/repository/search"
	"archi/internal/service"
	"archi/internal/web"
	"archi/internal/web/middleware/jwt"
	"archi/ioc"
	"github.com/google/wire"
)

var thirdPartyProviderSet = wire.NewSet(
	ioc.InitLogger,
	ioc.InitMySQL,
	ioc.InitRedis,
	ioc.InitRlockClient,
	ioc.InitSaramaClient,
	//ioc.InitESClient,
)

var userSvcProviderSet = wire.NewSet(
	cache.NewRedisUserCache,
	dao.NewGORMUserDAO,
	repository.NewCachedUserRepository,
	service.NewUserService,
)

var codeSvcProviderSet = wire.NewSet(
	cache.NewRedisCodeCache,
	repository.NewCachedCodeRepository,
	ioc.InitSMSService,
	service.NewDefaultCodeService,
)

// var wechatSvc = wire.NewSet(
//	ioc.InitWechatService,
// )

var articleSvcProviderSet = wire.NewSet(
	cache.NewRedisArticleCache,
	dao.NewGORMArticleDAO,
	repository.NewCachedArticleRepository,
	service.NewDefaultArticleService,
)

var interactiveSvcProviderSet = wire.NewSet(
	cache.NewRedisInteractiveCache,
	dao.NewGORMInteractiveDAO,
	repository.NewCachedInteractiveRepository,
	service.NewDefaultInteractiveService,
)

var rankingSvcProviderSet = wire.NewSet(
	cache.NewRedisRankingCache,
	cache.NewLocalRankingCache,
	repository.NewCachedRankingRepository,
	service.NewBatchRankingService,
)

var commentSvcProviderSet = wire.NewSet(
	dao.NewGORMCommentDAO,
	repository.NewCachedCommentRepository,
	service.NewDefaultCommentService,
)

var followSvcProviderSet = wire.NewSet(
	cache.NewRedisFollowCache,
	dao.NewGORMFollowRelationDAO,
	repository.NewCachedFollowRepository,
	service.NewDefaultFollowRelationService,
)

var searchSvcProviderSet = wire.NewSet(
	searchDAO.NewESUserDAO,
	searchDAO.NewESTagDAO,
	searchDAO.NewESArticleDAO,
	searchRepo.NewDefaultUserRepository,
	searchRepo.NewDefaultArticleRepository,
	service.NewDefaultSearchService,
)
var eventsProviderSet = wire.NewSet(
	ioc.InitSyncProducer,
	ioc.InitConsumers,
	article.NewSaramaSyncProducer,
	article.NewReadEventConsumer,
)

var handlerProviderSet = wire.NewSet(
	jwt.NewRedisJWTHandler,
	web.NewUserHandler,
	//web.NewOAuth2WechatHandler,
	web.NewArticleHandler,
	web.NewCommentHandler,
	web.NewFollowHandler,
)

var jobProviderSet = wire.NewSet(
	ioc.InitRankingJob,
	ioc.InitJobs,
)

func InitApp() *App {
	wire.Build(
		thirdPartyProviderSet,

		userSvcProviderSet,
		codeSvcProviderSet,
		//wechatSvc,
		articleSvcProviderSet,
		interactiveSvcProviderSet,
		rankingSvcProviderSet,
		commentSvcProviderSet,
		followSvcProviderSet,

		handlerProviderSet,
		jobProviderSet,
		eventsProviderSet,

		ioc.InitWebEngine,
		ioc.InitGinMiddlewares,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
