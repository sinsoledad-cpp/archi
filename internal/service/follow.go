package service

import (
	"archi/internal/domain"
	"archi/internal/event/follow"
	"archi/internal/repository"
	"archi/pkg/logger"
	"context"
)

type FollowRelationService interface {
	GetFollowee(ctx context.Context, follower, offset, limit int64) ([]domain.FollowRelation, error)
	GetFollower(ctx context.Context, followee int64) ([]domain.FollowRelation, error)
	FollowInfo(ctx context.Context, follower, followee int64) (domain.FollowRelation, error)
	Follow(ctx context.Context, follower, followee int64) error
	CancelFollow(ctx context.Context, follower, followee int64) error
	GetFollowStatic(ctx context.Context, uid int64) (domain.FollowStatics, error)
}

type DefaultFollowRelationService struct {
	producer follow.Producer
	repo     repository.FollowRepository
	l        logger.Logger
}

func NewDefaultFollowRelationService(producer follow.Producer, repo repository.FollowRepository, logger logger.Logger) FollowRelationService {
	return &DefaultFollowRelationService{
		producer: producer,
		repo:     repo,
		l:        logger,
	}
}
func (f *DefaultFollowRelationService) GetFollowee(ctx context.Context, follower, offset, limit int64) ([]domain.FollowRelation, error) {
	return f.repo.GetFollowee(ctx, follower, offset, limit)
}

func (f *DefaultFollowRelationService) GetFollower(ctx context.Context, followee int64) ([]domain.FollowRelation, error) {
	return f.repo.GetFollower(ctx, followee)
}

func (f *DefaultFollowRelationService) FollowInfo(ctx context.Context, follower, followee int64) (domain.FollowRelation, error) {
	val, err := f.repo.FollowInfo(ctx, follower, followee)
	return val, err
}

func (f *DefaultFollowRelationService) Follow(ctx context.Context, follower, followee int64) error {
	err := f.repo.AddFollowRelation(ctx, domain.FollowRelation{
		Followee: followee,
		Follower: follower,
	})
	if err != nil {
		return err
	}
	go func() {
		// 4. 检查错误并记录日志
		err := f.producer.ProduceFollowEvent(follow.Event{
			Follower: follower,
			Followee: followee,
		})
		if err != nil {
			// 使用结构化日志记录错误和关键上下文信息
			f.l.Error("发送关注事件失败",
				logger.Int64("follower", follower),
				logger.Int64("followee", followee),
				logger.Error(err))
		}
	}()
	return nil
}
func (f *DefaultFollowRelationService) CancelFollow(ctx context.Context, follower, followee int64) error {
	return f.repo.InactiveFollowRelation(ctx, follower, followee)
}

func (f *DefaultFollowRelationService) GetFollowStatic(ctx context.Context, uid int64) (domain.FollowStatics, error) {
	return f.repo.GetFollowStatics(ctx, uid)
}
