package job

import (
	"archi/internal/service/payment/wechat"
	"archi/pkg/logger"
	"context"
	"time"
)

type SyncWechatOrderJob struct {
	svc *wechat.NativePaymentService
	l   logger.Logger
}

func (s *SyncWechatOrderJob) Name() string {
	return "sync_wechat_order_job"
}

func (s *SyncWechatOrderJob) Run() error {
	offset := 0
	// 也可以做成参数
	const limit = 100
	// 三十分钟之前的订单我们就认为已经过期了。
	now := time.Now().Add(-time.Minute * 30)
	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		pmts, err := s.svc.FindExpiredPayment(ctx, offset, limit, now)
		cancel()
		if err != nil {
			// 直接中断，你也可以仔细区别不同错误
			return err
		}
		// 因为微信没有批量接口，所以我们这里也只能单个查询
		for _, pmt := range pmts {
			// 单个重新设置超时
			ctx, cancel = context.WithTimeout(context.Background(), time.Second)
			err = s.svc.SyncWechatInfo(ctx, pmt.BizTradeNO)
			if err != nil {
				// 这里你也可以中断，不过我个人倾向于处理完毕
				s.l.Error("同步微信支付信息失败",
					logger.String("trade_no", pmt.BizTradeNO),
					logger.Error(err))
			}
			cancel()
		}
		if len(pmts) < limit {
			// 没数据了
			return nil
		}
		offset = offset + len(pmts)
	}
}
