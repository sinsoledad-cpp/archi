package dao

import (
	"archi/internal/domain"
	"context"
	"database/sql"
	"gorm.io/gorm"
	"time"
)

type Payment struct {
	ID  int64 `gorm:"primaryKey,autoIncrement" bson:"id,omitempty"`
	Amt int64
	// 你存储枚举也可以，比如说 0-CNY
	// 目前磁盘内存那么便宜，直接放 string 也可以
	Currency string
	// 可以抽象认为，这是一个简短的描述
	// 也就是说即便是别的支付方式，这边也可以提供一个简单的描述
	// 你可以认为这算是冗余的数据，因为从原则上来说，我们可以完全不保存的。
	// 而是要求调用者直接 BizID 和 Biz 去找业务方要
	// 管得越少，系统越稳
	Description string `gorm:"description"`
	// 后续可以考虑增加字段，来标记是用的是微信支付亦或是支付宝支付
	// 也可以考虑提供一个巨大的 BLOB 字段，
	// 来存储和支付有关的其它字段
	// ExtraData

	// 业务方传过来的
	BizTradeNO string `gorm:"column:biz_trade_no;type:varchar(256);unique"`

	// 第三方支付平台的事务 ID，唯一的
	TxnID sql.NullString `gorm:"column:txn_id;type:varchar(128);unique"`

	Status uint8
	Utime  int64
	Ctime  int64
}
type PaymentDAO interface {
	Insert(ctx context.Context, pmt Payment) error
	UpdateTxnIDAndStatus(ctx context.Context, bizTradeNo string, txnID string, status domain.PaymentStatus) error
	FindExpiredPayment(ctx context.Context, offset int, limit int, t time.Time) ([]Payment, error)
	GetPayment(ctx context.Context, bizTradeNO string) (Payment, error)
}

type GORMPaymentDAO struct {
	db *gorm.DB
}

func NewGORMPaymentDAO(db *gorm.DB) PaymentDAO {
	return &GORMPaymentDAO{db: db}
}
func (p *GORMPaymentDAO) Insert(ctx context.Context, pmt Payment) error {
	now := time.Now().UnixMilli()
	pmt.Utime = now
	pmt.Ctime = now
	return p.db.WithContext(ctx).Create(&pmt).Error
}

func (p *GORMPaymentDAO) UpdateTxnIDAndStatus(ctx context.Context, bizTradeNo string, txnID string, status domain.PaymentStatus) error {
	return p.db.WithContext(ctx).Model(&Payment{}).
		Where("biz_trade_no = ?", bizTradeNo).
		Updates(map[string]any{
			"txn_id": txnID,
			"status": status.AsUint8(),
			"utime":  time.Now().UnixMilli(),
		}).Error
}

func (p *GORMPaymentDAO) FindExpiredPayment(ctx context.Context, offset int, limit int, t time.Time) ([]Payment, error) {
	var res []Payment
	err := p.db.WithContext(ctx).Where("status = ? AND utime < ?",
		domain.PaymentStatusInit.AsUint8(), t.UnixMilli()).
		Offset(offset).Limit(limit).Find(&res).Error
	return res, err
}

func (p *GORMPaymentDAO) GetPayment(ctx context.Context, bizTradeNO string) (Payment, error) {
	var res Payment
	err := p.db.WithContext(ctx).Where("biz_trade_no = ?", bizTradeNO).First(&res).Error
	return res, err
}
