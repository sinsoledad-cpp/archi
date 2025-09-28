package limiter

import "context"

//go:generate mockgen -source=./limiter.go -package=limitmocks -destination=./mocks/limiter.mock.go Limiter
type Limiter interface {
	// Limit 有咩有触发限流。key 就是限流对象
	// bool 代表是否限流，true 就是要限流
	// err 限流器本身有咩有错误
	Limit(ctx context.Context, key string) (bool, error)
}
