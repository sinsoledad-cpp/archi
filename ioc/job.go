package ioc

import (
	"archi/internal/job"
	"archi/internal/service"
	"archi/pkg/cronjobx"
	"archi/pkg/logger"
	rlock "github.com/gotomicro/redis-lock"
	"github.com/robfig/cron/v3"
	"time"
)

func InitRankingJob(svc service.RankingService, client *rlock.Client, l logger.Logger) *job.RankingJob {
	return job.NewRankingJob(svc, l, client, time.Second*30)
}

func InitJobs(l logger.Logger, rankingJob *job.RankingJob) *cron.Cron {
	builder := cronjobx.NewCronJobBuilder(l)

	timezone, _ := time.LoadLocation("Asia/Shanghai")
	//crontab := cron.New(cron.WithSeconds(), cron.WithLocation(timezone))

	expr := cron.New(cron.WithSeconds(), cron.WithLocation(timezone))
	_, err := expr.AddJob("@every 10s", builder.Build(rankingJob))
	if err != nil {
		panic(err)
	}
	return expr
}
