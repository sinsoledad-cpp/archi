package repository

import (
	"archi/internal/domain"
	"archi/internal/repository/dao"
	"context"
	"time"
)

type AccountRepository interface {
	AddCredit(ctx context.Context, c domain.Credit) error
}

type DefaultAccountRepository struct {
	dao dao.AccountDAO
}

func NewDefaultAccountRepository(dao dao.AccountDAO) AccountRepository {
	return &DefaultAccountRepository{dao: dao}
}

func (a *DefaultAccountRepository) AddCredit(ctx context.Context, c domain.Credit) error {
	activities := make([]dao.AccountActivity, 0, len(c.Items))
	now := time.Now().UnixMilli()
	for _, itm := range c.Items {
		activities = append(activities, dao.AccountActivity{
			Uid:         itm.Uid,
			Biz:         c.Biz,
			BizID:       c.BizID,
			AccountID:   itm.AccountID,
			AccountType: itm.AccountType.AsUint8(),
			Amount:      itm.Amt,
			Currency:    itm.Currency,
			Ctime:       now,
			Utime:       now,
		})
	}
	return a.dao.AddActivities(ctx, activities...)
}
