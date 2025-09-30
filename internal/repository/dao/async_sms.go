package dao

import (
	"context"
	"github.com/ecodeclub/ekit/sqlx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

const (
	// 因为本身状态没有暴露出去，所以不需要在 domain 里面定义
	asyncStatusWaiting uint8 = iota
	// 失败了，并且超过了重试次数
	asyncStatusFailed
	asyncStatusSuccess
)

var ErrWaitingSMSNotFound = gorm.ErrRecordNotFound

type SMSConfig struct {
	TplId   string
	Args    []string
	Numbers []string
}

type AsyncSMS struct {
	ID       int64                      `gorm:"primaryKey,autoIncrement"`
	Config   sqlx.JsonColumn[SMSConfig] // 使用我在 ekit 里面支持的 JSON 字段
	RetryCnt int                        // 重试次数
	RetryMax int                        // 重试的最大次数
	Status   uint8
	Ctime    int64
	Utime    int64 `gorm:"index"`
}

//go:generate mockgen -source=./async_sms.go -package=mocks -destination=mocks/async_sms.mock.go AsyncSmsDAO
type AsyncSMSDAO interface {
	Insert(ctx context.Context, s AsyncSMS) error
	GetWaitingSMS(ctx context.Context) (AsyncSMS, error)
	MarkSuccess(ctx context.Context, id int64) error
	MarkFailed(ctx context.Context, id int64) error
}

type GORMAsyncSMSDAO struct {
	db *gorm.DB
}

func NewGORMAsyncSMSDAO(db *gorm.DB) AsyncSMSDAO {
	return &GORMAsyncSMSDAO{
		db: db,
	}
}

func (g *GORMAsyncSMSDAO) Insert(ctx context.Context, s AsyncSMS) error {
	now := time.Now().UnixMilli()
	s.Ctime = now
	s.Utime = now
	return g.db.Create(&s).Error
}

func (g *GORMAsyncSMSDAO) GetWaitingSMS(ctx context.Context) (AsyncSMS, error) {
	// 如果在高并发情况下,SELECT for UPDATE 对数据库的压力很大
	// 但是我们不是高并发，因为你部署N台机器，才有 N 个goroutine 来查询
	// 并发不过百，随便写
	var s AsyncSMS
	err := g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 为了避开一些偶发性的失败，我们只找 1 分钟前的异步短信发送
		now := time.Now().UnixMilli()
		endTime := now - time.Minute.Milliseconds()
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("utime < ? and status = ?", endTime, asyncStatusWaiting).First(&s).Error
		// SELECT xx FROM xxx WHERE xx FOR UPDATE，锁住了
		if err != nil {
			return err
		}
		// 只要更新了更新时间，根据我们前面的规则，就不可能被别的节点抢占了
		err = tx.Model(&AsyncSMS{}).
			Where("id = ?", s.ID).
			Updates(map[string]any{
				// 更新成了当前时间戳，确保我在发送过程中，没人会再次抢到它 也相当于，重试间隔一分钟
				"retry_cnt": gorm.Expr("retry_cnt + 1"),
				"utime":     now,
			}).Error
		return err
	})
	return s, err
}

func (g *GORMAsyncSMSDAO) MarkSuccess(ctx context.Context, id int64) error {
	now := time.Now().UnixMilli()
	return g.db.WithContext(ctx).Model(&AsyncSMS{}).
		Where("id =?", id).
		Updates(map[string]any{
			"utime":  now,
			"status": asyncStatusSuccess,
		}).Error
}

func (g *GORMAsyncSMSDAO) MarkFailed(ctx context.Context, id int64) error {
	now := time.Now().UnixMilli()
	return g.db.WithContext(ctx).Model(&AsyncSMS{}).
		Where("id =? and `retry_cnt`>=`retry_max`", id). // 只有到达了重试次数才会更新
		Updates(map[string]any{
			"utime":  now,
			"status": asyncStatusFailed,
		}).Error
}
