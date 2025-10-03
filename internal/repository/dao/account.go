package dao

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

// Account 账号本体
type Account struct {
	ID int64 `gorm:"primaryKey,autoIncrement" bson:"id,omitempty"`
	// 对应的用户的 ID，如果是系统账号
	Uid int64 `gorm:"uniqueIndex:account_uid"`
	// 账号 ID，这个才是对外使用的
	AccountID int64 `gorm:"uniqueIndex:account_uid"`
	// 一个人可能有很多账号，你在这里可以用于区分
	Type uint8 `gorm:"uniqueIndex:account_uid"`

	// 账号本身可以有很多额外的字段
	// 例如跟会计有关的，跟税务有关的，跟洗钱有关的
	// 跟审计有关的，跟安全有关的

	// 可用余额
	// 一般来说，一种货币就一个账号，比较好处理（个人认为）
	// 有些一个账号，但是支持多种货币，那么就需要关联另外一张表。
	// 记录每一个币种的余额
	Balance  int64 //可用余额
	Currency string

	Ctime int64
	Utime int64
}

type AccountActivity struct {
	ID  int64 `gorm:"primaryKey,autoIncrement" bson:"id,omitempty"`
	Uid int64 `gorm:"index:account_uid"`
	// 这边有些设计会只用一个单独的 txn_id 来标记
	// 加上这些 业务 ID，DEBUG 的时候贼好用
	Biz   string
	BizID int64
	// account 账号
	AccountID   int64 `gorm:"index:account_uid"`
	AccountType uint8 `gorm:"index:account_uid"`
	// 调整的金额，有些设计不想引入负数，就会增加一个类型
	// 标记是增加还是减少，暂时我们还不需要
	Amount   int64
	Currency string

	Ctime int64
	Utime int64
}

func (AccountActivity) TableName() string {
	return "account_activities"
}

type AccountDAO interface {
	AddActivities(ctx context.Context, activities ...AccountActivity) error
}
type AccountGORMDAO struct {
	db *gorm.DB
}

func NewCreditGORMDAO(db *gorm.DB) AccountDAO {
	return &AccountGORMDAO{db: db}
}
func (c *AccountGORMDAO) AddActivities(ctx context.Context, activities ...AccountActivity) error {
	return c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UnixMilli()
		for _, act := range activities {
			// 一般在用户注册的时候就会创建好账号，但是我们并咩有，所以要兼容处理一下
			// 注意，系统账号是默认肯定存在的，一般是离线创建好的
			err := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "uid"}, {Name: "account_id"}, {Name: "type"}},
				DoUpdates: clause.Assignments(map[string]interface{}{
					"balance": gorm.Expr("balance + ?", act.Amount),
					"utime":   now,
				}),
			}).Create(&Account{
				Uid:       act.Uid,
				AccountID: act.AccountID,
				Type:      act.AccountType,
				Balance:   act.Amount,
				Currency:  act.Currency,
				Ctime:     now,
				Utime:     now,
			}).Error
			if err != nil {
				return err
			}
		}
		return tx.Create(&activities).Error
	})
}
