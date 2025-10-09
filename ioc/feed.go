package ioc

import (
	"archi/internal/repository"
	"archi/internal/service"
	"archi/internal/service/feed"
)

func RegisterFeedHandler(repo repository.FeedEventRepo, followSvc service.FollowRelationService) map[string]feed.Handler {
	articleHandler := feed.NewArticleEventHandler(repo, followSvc)
	followHandler := feed.NewFollowEventHandler(repo)
	likeHandler := feed.NewLikeEventHandler(repo)
	return map[string]feed.Handler{
		feed.ArticleEventName: articleHandler,
		feed.FollowEventName:  followHandler,
		feed.LikeEventName:    likeHandler,
	}
}
