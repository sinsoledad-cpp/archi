package service

import (
	"archi/internal/domain"
	"archi/internal/repository"
	"archi/pkg/logger"
	"context"
	"time"
)

type CronJobService interface {
	Preempt(ctx context.Context) (domain.Job, error)
	ResetNextTime(ctx context.Context, j domain.Job) error
	//Release(ctx context.Context, job domain.Job) error
	// 暴露 job 的增删改查方法
}
type cronJobService struct {
	repo            repository.JobRepository
	l               logger.Logger
	refreshInterval time.Duration
}

func NewCronJobService(repo repository.JobRepository, l logger.Logger) CronJobService {
	return &cronJobService{
		repo:            repo,
		l:               l,
		refreshInterval: time.Minute,
	}
}

func (c *cronJobService) Preempt(ctx context.Context) (domain.Job, error) {
	j, err := c.repo.Preempt(ctx)
	if err != nil {
		return domain.Job{}, err
	}
	ticker := time.NewTicker(c.refreshInterval)
	go func() {
		for range ticker.C {
			c.refresh(j.ID)
		}
	}()
	j.CancelFunc = func() {
		ticker.Stop()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		err := c.repo.Release(ctx, j.ID)
		if err != nil {
			c.l.Error("释放 job 失败", logger.Error(err), logger.Int64("jib", j.ID))
		}
	}
	return j, err
}
func (c *cronJobService) ResetNextTime(ctx context.Context, j domain.Job) error {
	nextTime := j.NextTime()
	return c.repo.UpdateNextTime(ctx, j.ID, nextTime)
}

func (c *cronJobService) refresh(id int64) {
	// 本质上就是更新一下更新时间
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := c.repo.UpdateUtime(ctx, id)
	if err != nil {
		c.l.Error("续约失败", logger.Error(err), logger.Int64("jid", id))
	}
}
