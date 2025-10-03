package dao

import (
	"context"
	"gorm.io/gorm"
	"time"
)

type Reward struct {
	Id      int64  `gorm:"primaryKey,autoIncrement" bson:"id,omitempty"`
	Biz     string `gorm:"index:biz_biz_id"`
	BizId   int64  `gorm:"index:biz_biz_id"`
	BizName string
	// 被打赏的人
	TargetUid int64 `gorm:"index"`
	// 直接采用 RewardStatus 的取值
	Status uint8
	Uid    int64 // 打赏的人
	Amount int64
	Ctime  int64
	Utime  int64
}
type RewardDAO interface {
	Insert(ctx context.Context, r Reward) (int64, error)
	GetReward(ctx context.Context, rid int64) (Reward, error)
	UpdateStatus(ctx context.Context, rid int64, status uint8) error
}
type RewardGORMDAO struct {
	db *gorm.DB
}

func NewRewardGORMDAO(db *gorm.DB) RewardDAO {
	return &RewardGORMDAO{db: db}
}
func (dao *RewardGORMDAO) Insert(ctx context.Context, r Reward) (int64, error) {
	now := time.Now().UnixMilli()
	r.Ctime = now
	r.Utime = now
	err := dao.db.WithContext(ctx).Create(&r).Error
	return r.Id, err
}
func (dao *RewardGORMDAO) GetReward(ctx context.Context, rid int64) (Reward, error) {
	// 通过 uid 来判定是自己的打赏，防止黑客捞数据
	var r Reward
	err := dao.db.WithContext(ctx).
		Where("id = ? ", rid).
		First(&r).Error
	return r, err
}
func (dao *RewardGORMDAO) UpdateStatus(ctx context.Context, rid int64, status uint8) error {
	return dao.db.WithContext(ctx).
		Where("id = ?", rid).
		Updates(map[string]any{
			"status": status,
			"utime":  time.Now().UnixMilli(),
		}).Error
}
