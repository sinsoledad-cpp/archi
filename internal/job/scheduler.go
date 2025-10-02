package job

import (
	"archi/internal/service"
	"archi/pkg/logger"
	"context"
	"golang.org/x/sync/semaphore"
	"time"
)

// Scheduler 负责从服务中获取任务并使用对应的执行器来运行它们。
type Scheduler struct {
	dbTimeout  time.Duration
	jobService service.CronJobService // 原 svc
	executors  map[string]Executor
	logger     logger.Logger // 原 l
	limiter    *semaphore.Weighted
}

func NewScheduler(jobService service.CronJobService, logger logger.Logger) *Scheduler {
	return &Scheduler{
		jobService: jobService,
		logger:     logger,
		dbTimeout:  time.Second,
		limiter:    semaphore.NewWeighted(100),
		executors:  make(map[string]Executor),
	}
}

func (s *Scheduler) RegisterExecutor(exec Executor) {
	s.executors[exec.Name()] = exec
}

func (s *Scheduler) Schedule(ctx context.Context) error {
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		err := s.limiter.Acquire(ctx, 1)
		if err != nil {
			return err
		}

		dbCtx, cancel := context.WithTimeout(ctx, s.dbTimeout)
		job, err := s.jobService.Preempt(dbCtx)
		cancel()
		if err != nil {
			// 抢占任务失败是正常情况，比如没有可用任务，记录日志或直接继续
			// s.logger.Debug("抢占任务失败或无任务", logger.Error(err))
			s.limiter.Release(1) // 别忘了释放信号量
			continue
		}

		executor, ok := s.executors[job.Executor]
		if !ok {
			s.logger.Error("找不到对应的执行器",
				logger.Int64("jid", job.ID),
				logger.String("executor", job.Executor))
			s.limiter.Release(1) // 释放信号量
			// 考虑是否需要释放 job
			job.CancelFunc()
			continue
		}

		go func() {
			defer func() {
				s.limiter.Release(1)
				job.CancelFunc()
			}()

			execErr := executor.Execute(ctx, job)
			if execErr != nil {
				s.logger.Error("执行任务失败",
					logger.Int64("jid", job.ID),
					logger.Error(execErr))
				return
			}

			resetErr := s.jobService.ResetNextTime(ctx, job)
			if resetErr != nil {
				s.logger.Error("重置下次执行时间失败",
					logger.Int64("jid", job.ID),
					logger.Error(resetErr))
			}
		}()
	}
}
