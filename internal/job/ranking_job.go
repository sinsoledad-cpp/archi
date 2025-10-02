package job

import (
	"archi/internal/service"
	"archi/pkg/logger"
	"context"
	rlock "github.com/gotomicro/redis-lock"
	"sync"
	"time"
)

type RankingJob struct {
	rankingSvc service.RankingService
	logger     logger.Logger
	timeout    time.Duration
	client     *rlock.Client
	lockKey    string

	localLock *sync.Mutex
	redisLock *rlock.Lock

	// 随机生成一个，就代表当前负载。你可以每隔一分钟生成一个
	load int32
}

func NewRankingJob(svc service.RankingService, l logger.Logger, client *rlock.Client, timeout time.Duration) *RankingJob {
	return &RankingJob{
		rankingSvc: svc,
		lockKey:    "job:ranking",
		logger:     l,
		client:     client,
		localLock:  &sync.Mutex{},
		timeout:    timeout,
	}
}
func (r *RankingJob) Name() string {
	return "ranking"
}
func (r *RankingJob) Run() error {
	r.localLock.Lock()
	defer r.localLock.Unlock()
	lock := r.redisLock
	if lock == nil {
		// 抢分布式锁
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*4)
		defer cancel()
		lock, err := r.client.Lock(ctx, r.lockKey, r.timeout,
			&rlock.FixIntervalRetry{
				Interval: time.Millisecond * 100,
				Max:      3,
			}, time.Second) // 重试的超时
		if err != nil {
			//r.localLock.Unlock()
			r.logger.Warn("获取分布式锁失败", logger.Error(err))
			return nil
		}
		r.redisLock = lock
		//r.localLock.Unlock()
		go func() {
			// 并不是非得一半就续约
			er := lock.AutoRefresh(r.timeout/2, r.timeout)
			if er != nil {
				// 续约失败了
				// 你也没办法中断当下正在调度的热榜计算（如果有）
				r.localLock.Lock()
				r.redisLock = nil
				//redisLock.Unlock()
				r.localLock.Unlock()
			}
		}()
	}
	// 这边就是你拿到了锁
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	return r.rankingSvc.TopN(ctx)
}

func (r *RankingJob) Close() error {
	r.localLock.Lock()
	lock := r.redisLock
	r.localLock.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return lock.Unlock(ctx)
}
