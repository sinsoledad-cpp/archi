package repository

import (
	"archi/internal/domain"
	"archi/internal/repository/cache"
	"archi/internal/repository/dao"
	"context"
)

type RewardRepository interface {
	CreateReward(ctx context.Context, reward domain.Reward) (int64, error)
	GetReward(ctx context.Context, rid int64) (domain.Reward, error)
	// GetCachedCodeURL 这两个方法的名字我们明确带上了缓存的字眼  是希望调用者明白这个是我们缓存下来的，属于业务逻辑的一部分
	GetCachedCodeURL(ctx context.Context, r domain.Reward) (domain.CodeURL, error)
	CachedCodeURL(ctx context.Context, cu domain.CodeURL, r domain.Reward) error
	UpdateStatus(ctx context.Context, rid int64, status domain.RewardStatus) error
}
type DefaultRewardRepository struct {
	dao   dao.RewardDAO
	cache cache.RewardCache
}

func NewDefaultRewardRepository(dao dao.RewardDAO, c cache.RewardCache) RewardRepository {
	return &DefaultRewardRepository{dao: dao, cache: c}
}

func (repo *DefaultRewardRepository) CreateReward(ctx context.Context, reward domain.Reward) (int64, error) {
	return repo.dao.Insert(ctx, repo.toEntity(reward))
}
func (repo *DefaultRewardRepository) GetReward(ctx context.Context, rid int64) (domain.Reward, error) {
	r, err := repo.dao.GetReward(ctx, rid)
	if err != nil {
		return domain.Reward{}, err
	}
	return repo.toDomain(r), nil
}
func (repo *DefaultRewardRepository) GetCachedCodeURL(ctx context.Context, r domain.Reward) (domain.CodeURL, error) {
	return repo.cache.GetCachedCodeURL(ctx, r)
}
func (repo *DefaultRewardRepository) CachedCodeURL(ctx context.Context, cu domain.CodeURL, r domain.Reward) error {
	return repo.cache.CachedCodeURL(ctx, cu, r)
}
func (repo *DefaultRewardRepository) UpdateStatus(ctx context.Context, rid int64, status domain.RewardStatus) error {
	return repo.dao.UpdateStatus(ctx, rid, status.AsUint8())
}
func (repo *DefaultRewardRepository) toEntity(r domain.Reward) dao.Reward {
	return dao.Reward{
		Status:    r.Status.AsUint8(),
		Biz:       r.Target.Biz,
		BizName:   r.Target.BizName,
		BizID:     r.Target.BizID,
		TargetUid: r.Target.Uid,
		Uid:       r.Uid,
		Amount:    r.Amt,
	}
}

func (repo *DefaultRewardRepository) toDomain(r dao.Reward) domain.Reward {
	return domain.Reward{
		ID:  r.ID,
		Uid: r.Uid,
		Target: domain.Target{
			Biz:     r.Biz,
			BizID:   r.BizID,
			BizName: r.BizName,
			Uid:     r.Uid,
		},
		Amt:    r.Amount,
		Status: domain.RewardStatus(r.Status),
	}
}
