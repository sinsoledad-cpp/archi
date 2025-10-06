package dao

import (
	"context"
	"github.com/ecodeclub/ekit/slice"
	"gorm.io/gorm"
	"time"
)

type Tag struct {
	Id    int64  `gorm:"primaryKey,autoIncrement"`
	Name  string `gorm:"type=varchar(4096)"`
	Uid   int64  `gorm:"index"`
	Ctime int64
	Utime int64
}

type TagBiz struct {
	Id    int64  `gorm:"primaryKey,autoIncrement"`
	BizId int64  `gorm:"index:biz_type_id"`
	Biz   string `gorm:"index:biz_type_id"`
	// 冗余字段，加快查询和删除
	Uid   int64 `gorm:"index"`
	Tid   int64
	Tag   *Tag  `gorm:"ForeignKey:Tid;AssociationForeignKey:Id;constraint:OnDelete:CASCADE"`
	Ctime int64 `bson:"ctime,omitempty"`
	Utime int64 `bson:"utime,omitempty"`
}

type TagDAO interface {
	CreateTag(ctx context.Context, tag Tag) (int64, error)
	CreateTagBiz(ctx context.Context, tagBiz []TagBiz) error
	GetTagsByUid(ctx context.Context, uid int64) ([]Tag, error)
	GetTagsByBiz(ctx context.Context, uid int64, biz string, bizId int64) ([]Tag, error)
	GetTags(ctx context.Context, offset, limit int) ([]Tag, error)
	GetTagsById(ctx context.Context, ids []int64) ([]Tag, error)
}

type GORMTagDAO struct {
	db *gorm.DB
}

func NewGORMTagDAO(db *gorm.DB) TagDAO {
	return &GORMTagDAO{
		db: db,
	}
}

func (dao *GORMTagDAO) CreateTag(ctx context.Context, tag Tag) (int64, error) {
	now := time.Now().UnixMilli()
	tag.Ctime = now
	tag.Utime = now
	err := dao.db.WithContext(ctx).Create(&tag).Error
	return tag.Id, err
}

// CreateTagBiz 全部替换 策略
func (dao *GORMTagDAO) CreateTagBiz(ctx context.Context, tagBiz []TagBiz) error {
	if len(tagBiz) == 0 {
		return nil
	}
	now := time.Now().UnixMilli()
	for i := range tagBiz {
		tagBiz[i].Ctime = now
		tagBiz[i].Utime = now
	}
	first := tagBiz[0]
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&TagBiz{}).Delete(" uid = ? AND biz = ? AND biz_id = ?", first.Uid, first.BizId, first.BizId).Error
		if err != nil {
			return err
		}
		return tx.Create(&tagBiz).Error
	})
}

func (dao *GORMTagDAO) GetTagsByUid(ctx context.Context, uid int64) ([]Tag, error) {
	var res []Tag
	err := dao.db.WithContext(ctx).Where("uid= ?", uid).Find(&res).Error
	return res, err
}

func (dao *GORMTagDAO) GetTagsByBiz(ctx context.Context, uid int64, biz string, bizId int64) ([]Tag, error) {
	// 这边使用 JOIN 查询，如果你不想使用 JOIN 查询，
	// 你就在 repository 里面分成两次查询
	// 直接用 preload 特性
	var res []TagBiz
	err := dao.db.WithContext(ctx).Model(&TagBiz{}).InnerJoins("Tag", dao.db.Model(&Tag{})).
		Where("Tag.uid = ? AND biz = ? AND biz_id = ?", uid, biz, bizId).Find(&res).Error
	return slice.Map(res, func(idx int, src TagBiz) Tag {
		return *src.Tag
	}), err
}

func (dao *GORMTagDAO) GetTags(ctx context.Context, offset, limit int) ([]Tag, error) {
	var res []Tag
	err := dao.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&res).Error
	return res, err
}
func (dao *GORMTagDAO) GetTagsById(ctx context.Context, ids []int64) ([]Tag, error) {
	var res []Tag
	err := dao.db.WithContext(ctx).Where("id IN ?", ids).Find(&res).Error
	return res, err
}
