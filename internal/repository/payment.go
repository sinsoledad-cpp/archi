package repository

import (
	"archi/internal/domain"
	"archi/internal/repository/dao"
	"context"
	"time"
)

//go:generate mockgen -source=types.go -destination=mocks/payment.mock.go --package=repomocks PaymentRepository
type PaymentRepository interface {
	AddPayment(ctx context.Context, pmt domain.Payment) error
	// UpdatePayment 这个设计有点差，因为
	UpdatePayment(ctx context.Context, pmt domain.Payment) error
	FindExpiredPayment(ctx context.Context, offset int, limit int, t time.Time) ([]domain.Payment, error)
	GetPayment(ctx context.Context, bizTradeNO string) (domain.Payment, error)
}
type DefaultPaymentRepository struct {
	dao dao.PaymentDAO
}

func NewDefaultPaymentRepository(d dao.PaymentDAO) PaymentRepository {
	return &DefaultPaymentRepository{
		dao: d,
	}
}

func (p *DefaultPaymentRepository) AddPayment(ctx context.Context, pmt domain.Payment) error {
	return p.dao.Insert(ctx, p.toEntity(pmt))
}
func (p *DefaultPaymentRepository) UpdatePayment(ctx context.Context, pmt domain.Payment) error {
	return p.dao.UpdateTxnIDAndStatus(ctx, pmt.BizTradeNO, pmt.TxnID, pmt.Status)
}
func (p *DefaultPaymentRepository) FindExpiredPayment(ctx context.Context, offset int, limit int, t time.Time) ([]domain.Payment, error) {
	pmts, err := p.dao.FindExpiredPayment(ctx, offset, limit, t)
	if err != nil {
		return nil, err
	}
	res := make([]domain.Payment, 0, len(pmts))
	for _, pmt := range pmts {
		res = append(res, p.toDomain(pmt))
	}
	return res, nil
}
func (p *DefaultPaymentRepository) GetPayment(ctx context.Context, bizTradeNO string) (domain.Payment, error) {
	r, err := p.dao.GetPayment(ctx, bizTradeNO)
	return p.toDomain(r), err
}

func (p *DefaultPaymentRepository) toDomain(pmt dao.Payment) domain.Payment {
	return domain.Payment{
		Amt: domain.Amount{
			Currency: pmt.Currency,
			Total:    pmt.Amt,
		},
		BizTradeNO:  pmt.BizTradeNO,
		Description: pmt.Description,
		Status:      domain.PaymentStatus(pmt.Status),
		TxnID:       pmt.TxnID.String,
	}
}

func (p *DefaultPaymentRepository) toEntity(pmt domain.Payment) dao.Payment {
	return dao.Payment{
		Amt:         pmt.Amt.Total,
		Currency:    pmt.Amt.Currency,
		BizTradeNO:  pmt.BizTradeNO,
		Description: pmt.Description,
		Status:      domain.PaymentStatusInit.AsUint8(),
	}
}
