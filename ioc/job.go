package ioc

import (
	"archi/internal/job"
	"archi/internal/service"
	"archi/pkg/cronjobx"
	"archi/pkg/logger"
	"time"

	rlock "github.com/gotomicro/redis-lock"
	"github.com/robfig/cron/v3"
)

func InitRankingJob(svc service.RankingService, client *rlock.Client, l logger.Logger) *job.RankingJob {
	return job.NewRankingJob(svc, l, client, time.Second*30)
}

func InitJobs(l logger.Logger, rankingJob *job.RankingJob) *cron.Cron {
	builder := cronjobx.NewCronJobBuilder(l)

	//timezone, _ := time.LoadLocation("Asia/Shanghai")
	//crontab := cron.New(cron.WithSeconds(), cron.WithLocation(timezone))

	//expr := cron.New(cron.WithSeconds(), cron.WithLocation(timezone))
	expr := cron.New(
		cron.WithSeconds(), //启用 秒 级调度精度
		cron.WithChain( //用来配置「Job 执行链（middleware chain）」
			cron.SkipIfStillRunning(cron.DefaultLogger), //如果上一次任务还没执行完，就跳过当前这次执行
		),
	)
	_, err := expr.AddJob("@every 10m", builder.Build(rankingJob))
	if err != nil {
		panic(err)
	}
	return expr
}
