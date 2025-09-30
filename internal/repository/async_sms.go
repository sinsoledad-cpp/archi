package repository

import (
	"archi/internal/domain"
	"archi/internal/repository/dao"
	"context"
	"github.com/ecodeclub/ekit/sqlx"
)

var ErrWaitingSMSNotFound = dao.ErrWaitingSMSNotFound

//go:generate mockgen -source=./async_sms_repository.go -package=repomocks -destination=mocks/async_sms_repository.mock.go AsyncSmsRepository
type AsyncSMSRepository interface {
	// Add 添加一个异步 SMS 记录。
	// 你叫做 Create 或者 Insert 也可以
	Add(ctx context.Context, s domain.AsyncSMS) error
	PreemptWaitingSMS(ctx context.Context) (domain.AsyncSMS, error)
	ReportScheduleResult(ctx context.Context, id int64, success bool) error
}
type DefaultAsyncSMSRepository struct {
	dao dao.AsyncSMSDAO
}

func NewDefaultAsyncSMSRepository(dao dao.AsyncSMSDAO) AsyncSMSRepository {
	return &DefaultAsyncSMSRepository{
		dao: dao,
	}
}
func (a *DefaultAsyncSMSRepository) Add(ctx context.Context, s domain.AsyncSMS) error {
	return a.dao.Insert(ctx, dao.AsyncSMS{
		Config: sqlx.JsonColumn[dao.SMSConfig]{
			Val: dao.SMSConfig{
				TplId:   s.TplId,
				Args:    s.Args,
				Numbers: s.Numbers,
			},
			Valid: true,
		},
		RetryMax: s.RetryMax,
	})
}

func (a *DefaultAsyncSMSRepository) PreemptWaitingSMS(ctx context.Context) (domain.AsyncSMS, error) {
	as, err := a.dao.GetWaitingSMS(ctx)
	if err != nil {
		return domain.AsyncSMS{}, err
	}
	return domain.AsyncSMS{
		ID:       as.ID,
		TplId:    as.Config.Val.TplId,
		Numbers:  as.Config.Val.Numbers,
		Args:     as.Config.Val.Args,
		RetryMax: as.RetryMax,
	}, nil
}

func (a *DefaultAsyncSMSRepository) ReportScheduleResult(ctx context.Context, id int64, success bool) error {
	if success {
		return a.dao.MarkSuccess(ctx, id)
	}
	return a.dao.MarkFailed(ctx, id)
}
