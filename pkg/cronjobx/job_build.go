package cronjobx

import (
	"archi/pkg/logger"
	"github.com/robfig/cron/v3"
)

type JobBuilder struct {
	l logger.Logger
}

func NewCronJobBuilder(l logger.Logger) *JobBuilder {
	return &JobBuilder{
		l: l,
	}
}

func (j *JobBuilder) Build(job Job) cron.Job {
	name := job.Name()
	return JobFunc(func() {
		j.l.Debug("开始运行", logger.String("name", name))
		err := job.Run()
		if err != nil {
			j.l.Error("执行失败", logger.Error(err), logger.String("name", name))
		}
		j.l.Debug("结束运行", logger.String("name", name))
	})
}

// JobFunc 是一个函数类型，实现了 cron.Job 接口。
type JobFunc func()

func (c JobFunc) Run() {
	c()
}
