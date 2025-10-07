package feed

import (
	"archi/internal/domain"
	"archi/internal/repository"
	"archi/internal/service"
	"context"
	"github.com/ecodeclub/ekit/slice"
	"golang.org/x/sync/errgroup"
	"sort"
	"sync"
	"time"
)

const (
	ArticleEventName = "article_event"
	threshold        = 4
	//threshold        = 32
)

type ArticleEventHandler struct {
	//followClient followv1.FollowServiceClient
	repo      repository.FeedEventRepo
	followSvc service.FollowRelationService
}

func NewArticleEventHandler(repo repository.FeedEventRepo, followSvc service.FollowRelationService) Handler {
	return &ArticleEventHandler{
		repo:      repo,
		followSvc: followSvc,
	}
}

func (a *ArticleEventHandler) CreateFeedEvent(ctx context.Context, ext domain.ExtendFields) error {
	uid, err := ext.Get("uid").AsInt64()
	if err != nil {
		return err
	}
	// 根据粉丝数判断使用推模型还是拉模型
	resp, err := a.followSvc.GetFollowStatic(ctx, uid)
	//	&followv1.GetFollowStaticRequest{
	//	Followee: uid,
	//})
	if err != nil {
		return err
	}
	// 粉丝数超出阈值使用拉模型
	if resp.Followers > threshold {
		return a.repo.CreatePullEvent(ctx, domain.FeedEvent{
			Uid:   uid,
			Type:  ArticleEventName,
			Ctime: time.Now(),
			Ext:   ext,
		})
	} else {
		// 使用推模型
		// 获取粉丝
		fresp, err := a.followSvc.GetFollower(ctx, uid)
		if err != nil {
			return err
		}
		events := make([]domain.FeedEvent, 0, len(fresp))
		for _, r := range fresp {
			events = append(events, domain.FeedEvent{
				Uid:   r.Follower,
				Type:  ArticleEventName,
				Ctime: time.Now(),
				Ext:   ext,
			})
		}
		return a.repo.CreatePushEvents(ctx, events)
	}
}
func (a *ArticleEventHandler) FindFeedEvents(ctx context.Context, uid, timestamp, limit int64) ([]domain.FeedEvent, error) {
	// 获取推模型事件
	var (
		eg errgroup.Group
		mu sync.Mutex
	)
	events := make([]domain.FeedEvent, 0, limit*2)
	// Push Event
	eg.Go(func() error {
		pushEvents, err := a.repo.FindPushEventsWithTyp(ctx, ArticleEventName, uid, timestamp, limit)
		if err != nil {
			return err
		}
		mu.Lock()
		events = append(events, pushEvents...)
		mu.Unlock()
		return nil
	})

	// Pull Event
	eg.Go(func() error {
		resp, rerr := a.followSvc.GetFollowee(ctx, uid, 0, 200)
		if rerr != nil {
			return rerr
		}
		followeeIds := slice.Map(resp, func(idx int, src domain.FollowRelation) int64 {
			return src.Followee
		})
		pullEvents, err := a.repo.FindPullEventsWithTyp(ctx, ArticleEventName, followeeIds, timestamp, limit)
		if err != nil {
			return err
		}
		mu.Lock()
		events = append(events, pullEvents...)
		mu.Unlock()
		return nil
	})
	err := eg.Wait()
	if err != nil {
		return nil, err
	}
	// 获取拉模型事件
	// 获取默认的关注列表
	sort.Slice(events, func(i, j int) bool {
		return events[i].Ctime.Unix() > events[j].Ctime.Unix()
	})

	return events[:slice.Min[int]([]int{int(limit), len(events)})], nil
}
