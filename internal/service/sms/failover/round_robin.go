package failover

import (
	"archi/internal/service/sms"
	"context"
	"errors"
	"sync/atomic"
)

var _ sms.Service = &RoundRobinService{}

type RoundRobinService struct {
	providers []sms.Service
	// v1 的字段
	// 当前服务商下标
	idx uint64
}

func NewFailOverSMSService(providers []sms.Service) sms.Service {
	return &RoundRobinService{
		providers: providers,
	}
}

func (f *RoundRobinService) SendV1(ctx context.Context, tplId string, args []string, numbers ...string) error {
	for _, svc := range f.providers {
		err := svc.Send(ctx, tplId, args, numbers...)
		if err == nil {
			return nil
		}
		//log.Println(err)
	}
	return errors.New("轮询了所有的服务商，但是发送都失败了")
}

// 起始下标轮询
// 并且出错也轮询
func (f *RoundRobinService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	// idx 是你的局部变量
	idx := atomic.AddUint64(&f.idx, 1)
	length := uint64(len(f.providers))
	// 我要迭代 length
	for i := idx; i < idx+length; i++ {
		// 取余数来计算下标
		svc := f.providers[i%length]
		err := svc.Send(ctx, tplId, args, numbers...)
		switch {
		case err == nil:
			return nil
		case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
			// 前者是被取消，后者是超时
			return err
		}
		//log.Println(err)
	}
	return errors.New("轮询了所有的服务商，但是发送都失败了")
}
