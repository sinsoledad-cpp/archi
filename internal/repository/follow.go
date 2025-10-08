package repository

import (
	"archi/internal/domain"
	"archi/internal/repository/cache"
	"archi/internal/repository/dao"
	"archi/pkg/logger"
	"context"
)

type FollowRepository interface {
	// GetFollowee 获取某人的关注列表
	GetFollower(ctx context.Context, followee int64) ([]domain.FollowRelation, error)
	GetFollowee(ctx context.Context, follower, offset, limit int64) ([]domain.FollowRelation, error)
	// FollowInfo 查看关注人的详情
	FollowInfo(ctx context.Context, follower int64, followee int64) (domain.FollowRelation, error)
	// AddFollowRelation 创建关注关系
	AddFollowRelation(ctx context.Context, f domain.FollowRelation) error
	// InactiveFollowRelation 取消关注
	InactiveFollowRelation(ctx context.Context, follower int64, followee int64) error
	GetFollowStatics(ctx context.Context, uid int64) (domain.FollowStatics, error)
}

type CachedFollowRepository struct {
	dao   dao.FollowRelationDao
	cache cache.FollowCache
	l     logger.Logger
}

func NewCachedFollowRepository(dao dao.FollowRelationDao, cache cache.FollowCache, l logger.Logger) FollowRepository {
	return &CachedFollowRepository{
		dao:   dao,
		cache: cache,
		l:     l,
	}
}

func (d *CachedFollowRepository) GetFollower(ctx context.Context, followee int64) ([]domain.FollowRelation, error) {
	followerList, err := d.dao.GetFollowerList(ctx, followee)
	if err != nil {
		return nil, err
	}
	return d.genFollowRelationList(followerList), nil
}

func (d *CachedFollowRepository) GetFollowee(ctx context.Context, follower, offset, limit int64) ([]domain.FollowRelation, error) {
	followerList, err := d.dao.FollowRelationList(ctx, follower, offset, limit)
	if err != nil {
		return nil, err
	}
	return d.genFollowRelationList(followerList), nil
}

func (d *CachedFollowRepository) genFollowRelationList(followerList []dao.FollowRelation) []domain.FollowRelation {
	res := make([]domain.FollowRelation, 0, len(followerList))
	for _, c := range followerList {
		res = append(res, d.toDomain(c))
	}
	return res
}

func (d *CachedFollowRepository) FollowInfo(ctx context.Context, follower int64, followee int64) (domain.FollowRelation, error) {
	c, err := d.dao.FollowRelationDetail(ctx, follower, followee)
	if err != nil {
		return domain.FollowRelation{}, err
	}
	return d.toDomain(c), nil
}

func (d *CachedFollowRepository) AddFollowRelation(ctx context.Context, c domain.FollowRelation) error {
	err := d.dao.CreateFollowRelation(ctx, d.toEntity(c))
	if err != nil {
		return err
	}
	return d.cache.Follow(ctx, c.Follower, c.Followee)
}

func (d *CachedFollowRepository) InactiveFollowRelation(ctx context.Context, follower int64, followee int64) error {
	err := d.dao.UpdateStatus(ctx, followee, follower, dao.FollowRelationStatusInactive)
	if err != nil {
		return err
	}
	return d.cache.CancelFollow(ctx, follower, followee)
}

func (d *CachedFollowRepository) GetFollowStatics(ctx context.Context, uid int64) (domain.FollowStatics, error) {
	// 快路径
	res, err := d.cache.StaticsInfo(ctx, uid)
	if err == nil {
		return res, err
	}
	// 慢路径
	res.Followers, err = d.dao.CntFollower(ctx, uid)
	if err != nil {
		return res, err
	}
	res.Followees, err = d.dao.CntFollowee(ctx, uid)
	if err != nil {
		return res, err
	}
	err = d.cache.SetStaticsInfo(ctx, uid, res)
	if err != nil {
		// 这里记录日志
		d.l.Error("缓存关注统计信息失败",
			logger.Error(err),
			logger.Int64("uid", uid))
	}
	return res, nil
}
func (d *CachedFollowRepository) toDomain(fr dao.FollowRelation) domain.FollowRelation {
	return domain.FollowRelation{
		Followee: fr.Followee,
		Follower: fr.Follower,
	}
}

func (d *CachedFollowRepository) toEntity(c domain.FollowRelation) dao.FollowRelation {
	return dao.FollowRelation{
		Followee: c.Followee,
		Follower: c.Follower,
	}
}
