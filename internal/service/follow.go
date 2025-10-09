package service

import (
	"archi/internal/domain"
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
	repo repository.FollowRepository
	l    logger.Logger
}

func NewDefaultFollowRelationService(repo repository.FollowRepository, logger logger.Logger) FollowRelationService {
	return &DefaultFollowRelationService{
		repo: repo,
		l:    logger,
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

	return nil
}
func (f *DefaultFollowRelationService) CancelFollow(ctx context.Context, follower, followee int64) error {
	return f.repo.InactiveFollowRelation(ctx, follower, followee)
}

func (f *DefaultFollowRelationService) GetFollowStatic(ctx context.Context, uid int64) (domain.FollowStatics, error) {
	return f.repo.GetFollowStatics(ctx, uid)
}
