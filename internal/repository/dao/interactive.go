package dao

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type UserLikeBiz struct {
	Id     int64  `gorm:"primaryKey,autoIncrement"`
	Uid    int64  `gorm:"uniqueIndex:uid_biz_type_id"`
	BizId  int64  `gorm:"uniqueIndex:uid_biz_type_id"`
	Biz    string `gorm:"type:varchar(128);uniqueIndex:uid_biz_type_id"`
	Status int    // 新增 Status 字段，0 表示取消，1 表示有效
	Utime  int64
	Ctime  int64
}

type UserCollectionBiz struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 这边还是保留了了唯一索引
	// 收藏夹的ID
	// 收藏夹ID本身有索引
	Uid    int64  `gorm:"uniqueIndex:uid_biz_cid_id"`
	BizId  int64  `gorm:"uniqueIndex:uid_biz_cid_id"`
	Biz    string `gorm:"type:varchar(128);uniqueIndex:uid_biz_cid_id"`
	Cid    int64  `gorm:"uniqueIndex:uid_biz_cid_id,index"`
	Status int    // 新增 Status 字段，0 表示取消，1 表示有效

	Utime int64
	Ctime int64
}

type Interactive struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// <bizid, biz>
	BizId int64 `gorm:"uniqueIndex:biz_type_id"`
	// WHERE biz = ?
	Biz string `gorm:"type:varchar(128);uniqueIndex:biz_type_id"`

	ReadCnt    int64
	LikeCnt    int64
	CollectCnt int64
	Utime      int64
	Ctime      int64
}
type InteractiveDAO interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	BatchIncrReadCnt(ctx context.Context, bizs []string, bizIds []int64) error
	InsertLikeInfo(ctx context.Context, biz string, id int64, uid int64) error
	DeleteLikeInfo(ctx context.Context, biz string, id int64, uid int64) error
	InsertCollectionBiz(ctx context.Context, cb UserCollectionBiz) error
	GetLikeInfo(ctx context.Context, biz string, id int64, uid int64) (UserLikeBiz, error)
	GetCollectInfo(ctx context.Context, biz string, id int64, uid int64) (UserCollectionBiz, error)
	Get(ctx context.Context, biz string, id int64) (Interactive, error)
	GetByIds(ctx context.Context, biz string, ids []int64) ([]Interactive, error)
}
type GORMInteractiveDAO struct {
	db *gorm.DB
}

func NewGORMInteractiveDAO(db *gorm.DB) *GORMInteractiveDAO {
	return &GORMInteractiveDAO{db: db}
}

func (g *GORMInteractiveDAO) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	now := time.Now().UnixMilli()
	return g.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "biz_id"},
			{Name: "biz"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"read_cnt": gorm.Expr("`read_cnt` + 1"),
			"utime":    now,
		}),
	}).Create(&Interactive{
		Biz:     biz,
		BizId:   bizId,
		ReadCnt: 1,
		Ctime:   now,
		Utime:   now,
	}).Error
}
func (g *GORMInteractiveDAO) BatchIncrReadCnt(ctx context.Context, bizs []string, bizIds []int64) error {
	return g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txDAO := NewGORMInteractiveDAO(tx)
		for i := 0; i < len(bizs); i++ {
			err := txDAO.IncrReadCnt(ctx, bizs[i], bizIds[i])
			if err != nil {
				return err
			}
		}
		return nil
	})
}
func (g *GORMInteractiveDAO) InsertLikeInfo(ctx context.Context, biz string, id int64, uid int64) error {
	now := time.Now().UnixMilli()
	return g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "uid"},
				{Name: "biz_id"},
				{Name: "biz"},
			},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"utime":  now,
				"status": 1,
			}),
		}).Create(&UserLikeBiz{
			Uid:    uid,
			Biz:    biz,
			BizId:  id,
			Status: 1,
			Utime:  now,
			Ctime:  now,
		}).Error
		if err != nil {
			return err
		}
		return tx.WithContext(ctx).Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "biz_id"},
				{Name: "biz"},
			},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"like_cnt": gorm.Expr("`like_cnt` + 1"),
				"utime":    now,
			}),
		}).Create(&Interactive{
			Biz:     biz,
			BizId:   id,
			LikeCnt: 1,
			Ctime:   now,
			Utime:   now,
		}).Error
	})
}

func (g *GORMInteractiveDAO) DeleteLikeInfo(ctx context.Context, biz string, id int64, uid int64) error {
	now := time.Now().UnixMilli()
	return g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&UserLikeBiz{}).
			Where("uid=? AND biz_id = ? AND biz=?", uid, id, biz).
			Updates(map[string]interface{}{
				"utime":  now,
				"status": 0,
			})
		if res.Error != nil {
			return res.Error
		}
		// 只有在确实更新了一条记录（即从未取消状态变为取消状态）时，才减少总数
		if res.RowsAffected == 0 {
			return nil
		}
		return tx.Model(&Interactive{}).
			Where("biz =? AND biz_id=?", biz, id).
			Updates(map[string]interface{}{
				"like_cnt": gorm.Expr("`like_cnt` - 1"),
				"utime":    now,
			}).Error
	})
}
func (g *GORMInteractiveDAO) InsertCollectionBiz(ctx context.Context, cb UserCollectionBiz) error {
	now := time.Now().UnixMilli()
	cb.Ctime = now
	cb.Utime = now
	cb.Status = 1 // 确保是有效状态
	return g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 使用 OnConflict 来处理重复收藏的情况
		err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "uid"},
				{Name: "biz_id"},
				{Name: "biz"},
				{Name: "cid"},
			},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"utime":  now,
				"status": 1, // 如果是重新收藏，则更新状态
			}),
		}).Create(&cb).Error

		if err != nil {
			return err
		}

		// 同样，增加收藏计数
		return tx.WithContext(ctx).Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "biz_id"},
				{Name: "biz"},
			},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"collect_cnt": gorm.Expr("`collect_cnt` + 1"),
				"utime":       now,
			}),
		}).Create(&Interactive{
			Biz:        cb.Biz,
			BizId:      cb.BizId,
			CollectCnt: 1,
			Ctime:      now,
			Utime:      now,
		}).Error
	})
}

// DeleteCollectionBiz 替换为逻辑删除的实现
func (g *GORMInteractiveDAO) DeleteCollectionBiz(ctx context.Context, biz string, bizId int64, cid int64, uid int64) error {
	now := time.Now().UnixMilli()
	return g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 逻辑删除 UserCollectionBiz 表中的记录
		res := tx.Model(&UserCollectionBiz{}).
			Where("uid = ? AND biz_id = ? AND biz = ? AND cid = ? AND status = ?", uid, bizId, biz, cid, 1).
			Updates(map[string]interface{}{
				"utime":  now,
				"status": 0, // 将状态更新为 0 (无效)
			})
		if res.Error != nil {
			return res.Error
		}
		// 只有在确实更新了一条记录（即从未取消状态变为取消状态）时，才减少总数
		if res.RowsAffected == 0 {
			return nil
		}
		// 更新 Interactive 表中的收藏总数
		return tx.Model(&Interactive{}).
			Where("biz = ? AND biz_id = ?", biz, bizId).
			Updates(map[string]interface{}{
				"collect_cnt": gorm.Expr("`collect_cnt` - 1"),
				"utime":       now,
			}).Error
	})
}
func (g *GORMInteractiveDAO) GetLikeInfo(ctx context.Context, biz string, id int64, uid int64) (UserLikeBiz, error) {
	var res UserLikeBiz
	err := g.db.WithContext(ctx).
		Where("uid = ? AND biz_id = ? AND biz = ? AND status = ?", uid, id, biz, 1).
		First(&res).Error
	return res, err
}

func (g *GORMInteractiveDAO) GetCollectInfo(ctx context.Context, biz string, id int64, uid int64) (UserCollectionBiz, error) {
	var res UserCollectionBiz
	err := g.db.WithContext(ctx).
		// 增加 status = 1 的查询条件
		Where("uid = ? AND biz_id = ? AND biz = ? AND status = ?", uid, id, biz, 1).
		First(&res).Error
	return res, err
}
func (g *GORMInteractiveDAO) Get(ctx context.Context, biz string, id int64) (Interactive, error) {
	var res Interactive
	err := g.db.WithContext(ctx).
		Where("biz_id =? AND  biz = ?", id, biz).
		First(&res).Error
	return res, err
}
func (g *GORMInteractiveDAO) GetByIds(ctx context.Context, biz string, ids []int64) ([]Interactive, error) {
	var res []Interactive
	err := g.db.WithContext(ctx).
		Where("biz = ? AND biz_id IN ?", biz, ids).
		First(&res).Error
	return res, err
}
