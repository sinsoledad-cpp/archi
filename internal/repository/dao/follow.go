package dao

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

const (
	FollowRelationStatusUnknown uint8 = iota
	FollowRelationStatusActive
	FollowRelationStatusInactive
)

type FollowRelation struct {
	ID int64 `gorm:"primaryKey,autoIncrement,column:id"`

	Follower int64 `gorm:"type:int(11);not null;uniqueIndex:follower_followee"`
	Followee int64 `gorm:"type:int(11);not null;uniqueIndex:follower_followee"`

	Status uint8

	// 这里你可以根据自己的业务来增加字段，比如说
	// 关系类型，可以搞些什么普通关注，特殊关注
	// Type int64 `gorm:"column:type;type:int(11);comment:关注类型 0-普通关注"`
	// 备注
	// Note string `gorm:"column:remark;type:varchar(255);"`
	// 创建时间
	Ctime int64
	Utime int64
}

type FollowRelationDao interface {
	// 获取莫人的粉丝列表
	GetFollowerList(ctx context.Context, followee int64) ([]FollowRelation, error)
	// FollowRelationList 获取某人的关注列表
	FollowRelationList(ctx context.Context, follower, offset, limit int64) ([]FollowRelation, error)
	FollowRelationDetail(ctx context.Context, follower int64, followee int64) (FollowRelation, error)
	// CreateFollowRelation 创建联系人
	CreateFollowRelation(ctx context.Context, c FollowRelation) error
	// UpdateStatus 更新状态
	UpdateStatus(ctx context.Context, followee int64, follower int64, status uint8) error
	// CntFollower 统计计算关注自己的人有多少
	CntFollower(ctx context.Context, uid int64) (int64, error)
	// CntFollowee 统计自己关注了多少人
	CntFollowee(ctx context.Context, uid int64) (int64, error)
}
type GORMFollowRelationDAO struct {
	db *gorm.DB
}

func NewGORMFollowRelationDAO(db *gorm.DB) FollowRelationDao {
	return &GORMFollowRelationDAO{
		db: db,
	}
}
func (g *GORMFollowRelationDAO) GetFollowerList(ctx context.Context, followee int64) ([]FollowRelation, error) {
	var res []FollowRelation
	err := g.db.WithContext(ctx).
		Where("followee = ? AND status = ?", followee, FollowRelationStatusActive).
		Find(&res).Error
	return res, err
}

func (g *GORMFollowRelationDAO) FollowRelationList(ctx context.Context, follower, offset, limit int64) ([]FollowRelation, error) {
	var res []FollowRelation
	err := g.db.WithContext(ctx).
		Where("follower = ? AND status = ?", follower, FollowRelationStatusActive).
		Offset(int(offset)).Limit(int(limit)).Find(&res).Error
	return res, err
}

func (g *GORMFollowRelationDAO) FollowRelationDetail(ctx context.Context, follower int64, followee int64) (FollowRelation, error) {
	var res FollowRelation
	err := g.db.WithContext(ctx).
		Where("follower = ? AND followee = ? AND status = ?", follower, followee, FollowRelationStatusActive).
		First(&res).Error
	return res, err
}

func (g *GORMFollowRelationDAO) CreateFollowRelation(ctx context.Context, f FollowRelation) error {
	now := time.Now().UnixMilli()
	f.Utime = now
	f.Ctime = now
	f.Status = FollowRelationStatusActive
	return g.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "follower"}, {Name: "followee"}},
		DoUpdates: clause.Assignments(map[string]any{
			"utime":  now,
			"status": FollowRelationStatusActive,
		}),
	}).Create(&f).Error
}

func (g *GORMFollowRelationDAO) UpdateStatus(ctx context.Context, followee int64, follower int64, status uint8) error {
	now := time.Now().UnixMilli()
	return g.db.WithContext(ctx).
		Where("follower = ? AND followee = ?", follower, followee).
		Updates(map[string]any{
			"status": status,
			"utime":  now,
		}).Error
}

func (g *GORMFollowRelationDAO) CntFollower(ctx context.Context, uid int64) (int64, error) {
	var res int64
	//err := g.db.WithContext(ctx).
	//	Select("count(follower)").
	//	Where("followee = ? AND status = ?", uid, FollowRelationStatusActive).
	//	Count(&res).Error

	err := g.db.WithContext(ctx).Model(&FollowRelation{}).
		Where("followee = ? AND status = ?", uid, FollowRelationStatusActive).
		Count(&res).Error
	return res, err
}

func (g *GORMFollowRelationDAO) CntFollowee(ctx context.Context, uid int64) (int64, error) {
	var res int64
	//err := g.db.WithContext(ctx).
	//	Select("count(followee)").
	//	Where("follower = ? AND status = ?", uid, FollowRelationStatusActive).
	//	Count(&res).Error

	err := g.db.WithContext(ctx).Model(&FollowRelation{}).
		Where("follower = ? AND status = ?", uid, FollowRelationStatusActive).
		Count(&res).Error
	return res, err
}
