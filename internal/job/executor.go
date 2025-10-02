package job

import (
	"archi/internal/domain"
	"context"
	"fmt"
)

// Executor 定义了如何执行一个任务。
type Executor interface {
	Name() string
	// Execute 执行任务。实现者需要正确处理 context 的取消信号。
	Execute(ctx context.Context, job domain.Job) error
}

// LocalFuncExecutor 是一种在本地执行预注册函数的 Executor。
type LocalFuncExecutor struct {
	funcs map[string]func(ctx context.Context, job domain.Job) error
}

func NewLocalFuncExecutor() *LocalFuncExecutor {
	return &LocalFuncExecutor{funcs: make(map[string]func(ctx context.Context, j domain.Job) error)}
}

func (l *LocalFuncExecutor) Name() string {
	return "local"
}

func (l *LocalFuncExecutor) RegisterFunc(name string, fn func(ctx context.Context, j domain.Job) error) {
	l.funcs[name] = fn
}

// Execute 实现了 Executor 接口。
func (l *LocalFuncExecutor) Execute(ctx context.Context, job domain.Job) error {
	fn, ok := l.funcs[job.Name]
	if !ok {
		return fmt.Errorf("未注册本地方法 %s", job.Name)
	}
	return fn(ctx, job)
}
