package repository

import (
	"archi/internal/domain"
	"archi/internal/repository/dao"
	"context"
	"time"
)

type JobRepository interface {
	Preempt(ctx context.Context) (domain.Job, error)
	Release(ctx context.Context, jid int64) error
	UpdateUtime(ctx context.Context, id int64) error
	UpdateNextTime(ctx context.Context, id int64, time time.Time) error
}

type PreemptJobRepository struct {
	dao dao.JobDAO
}

func NewPreemptJobRepository(dao dao.JobDAO) JobRepository {
	return &PreemptJobRepository{dao: dao}
}

func (p *PreemptJobRepository) Preempt(ctx context.Context) (domain.Job, error) {
	j, err := p.dao.Preempt(ctx)
	if err != nil {
		// 如果有错误，直接返回一个 domain.Job 的零值和错误
		return domain.Job{}, err
	}
	return domain.Job{
		ID:         j.ID,
		Expression: j.Expression,
		Executor:   j.Executor,
		Name:       j.Name,
	}, nil
}

func (p *PreemptJobRepository) Release(ctx context.Context, jid int64) error {
	return p.dao.Release(ctx, jid)
}

func (p *PreemptJobRepository) UpdateUtime(ctx context.Context, id int64) error {
	return p.dao.UpdateUtime(ctx, id)
}

func (p *PreemptJobRepository) UpdateNextTime(ctx context.Context, id int64, time time.Time) error {
	return p.dao.UpdateNextTime(ctx, id, time)
}
