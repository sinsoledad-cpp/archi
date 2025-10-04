package dao

import (
	"context"
	"database/sql"
	"gorm.io/gorm"
)

// ErrDataNotFound 通用的数据没找到
var ErrDataNotFound = gorm.ErrRecordNotFound

type Comment struct {
	Id int64 `gorm:"column:id;primaryKey" json:"id"`
	// 发表评论的用户
	Uid int64 `gorm:"column:uid;index" json:"uid"`
	// 发表评论的业务类型
	Biz string `gorm:"column:biz;index:biz_type_id" json:"biz"`
	// 对应的业务ID
	BizID int64 `gorm:"column:biz_id;index:biz_type_id" json:"bizID"`
	// 根评论为0表示一级评论
	RootID sql.NullInt64 `gorm:"column:root_id;index" json:"rootID"`
	// 父级评论
	PID sql.NullInt64 `gorm:"column:pid;index" json:"pid"`
	// 外键 用于级联删除
	ParentComment *Comment `gorm:"ForeignKey:PID;AssociationForeignKey:ID;constraint:OnDelete:CASCADE"`
	// 评论内容
	Content string `gorm:"type:text;column:content" json:"content"`
	// 创建时间
	Ctime int64 `gorm:"column:ctime;" json:"ctime"`
	// 更新时间
	Utime int64 `gorm:"column:utime;" json:"utime"`
}

func (*Comment) TableName() string {
	return "comments"
}

//go:generate mockgen -source=./comment.go -package=daomocks -destination=mocks/comment.mock.go CommentDAO
type CommentDAO interface {
	Insert(ctx context.Context, u Comment) (int64, error)
	// FindByBiz 只查找一级评论
	FindByBiz(ctx context.Context, biz string, bizId, minID, limit int64) ([]Comment, error)
	// FindCommentList Comment的id为0 获取一级评论，如果不为0获取对应的评论，和其评论的所有回复
	FindCommentList(ctx context.Context, u Comment) ([]Comment, error)
	FindRepliesByPid(ctx context.Context, pid int64, offset, limit int) ([]Comment, error)
	// Delete 删除本节点和其对应的子节点
	Delete(ctx context.Context, u Comment) error
	FindOneByIDs(ctx context.Context, id []int64) ([]Comment, error)
	FindRepliesByRid(ctx context.Context, rid int64, id int64, limit int64) ([]Comment, error)
}
type GORMCommentDAO struct {
	db *gorm.DB
}

func NewCommentDAO(db *gorm.DB) CommentDAO {
	return &GORMCommentDAO{
		db: db,
	}
}

func (c *GORMCommentDAO) Insert(ctx context.Context, u Comment) (int64, error) {
	err := c.db.WithContext(ctx).Create(u).Error
	return u.Id, err
}

// FindByBiz 只查找一级评论
func (c *GORMCommentDAO) FindByBiz(ctx context.Context, biz string, bizId, minID, limit int64) ([]Comment, error) {
	var res []Comment
	err := c.db.WithContext(ctx).
		Where("biz = ? AND biz_id = ? AND id < ? AND pid IS NULL", biz, bizId, minID).
		Limit(int(limit)).Find(&res).Error
	return res, err
}

// FindCommentList Comment的id为0 获取一级评论，如果不为0 获取对应的评论，和其评论的所有回复
func (c *GORMCommentDAO) FindCommentList(ctx context.Context, u Comment) ([]Comment, error) {
	var res []Comment
	builder := c.db.WithContext(ctx)
	if u.Id == 0 {
		builder = builder.
			Where("biz=?", u.Biz).
			Where("biz_id=?", u.BizID).
			Where("root_id is null")
	} else {
		builder = builder.Where("root_id=? or id =?", u.Id, u.Id)
	}
	err := builder.Find(&res).Error
	return res, err
}

// FindRepliesByPid 查找评论的直接评论
func (c *GORMCommentDAO) FindRepliesByPid(ctx context.Context, pid int64, offset, limit int) ([]Comment, error) {
	var res []Comment
	err := c.db.WithContext(ctx).Where("pid = ?", pid).
		Order("id DESC").Offset(offset).Limit(limit).Find(&res).Error
	return res, err
}
func (c *GORMCommentDAO) Delete(ctx context.Context, u Comment) error {
	return c.db.WithContext(ctx).Delete(&Comment{
		Id: u.Id,
	}).Error
}
func (c *GORMCommentDAO) FindOneByIDs(ctx context.Context, ids []int64) ([]Comment, error) {
	var res []Comment
	err := c.db.WithContext(ctx).Where("id in ?", ids).First(&res).Error
	return res, err
}

// FindRepliesByRid 加载更多的回复
func (c *GORMCommentDAO) FindRepliesByRid(ctx context.Context, rid int64, id int64, limit int64) ([]Comment, error) {
	var res []Comment
	err := c.db.WithContext(ctx).
		Where("root_id = ? AND id < ?", rid, id).
		Order("id DESC").
		Limit(int(limit)).Find(&res).Error
	return res, err
}
