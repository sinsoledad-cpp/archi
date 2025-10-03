package service

import (
	"archi/internal/domain"
	"archi/internal/repository"
	"archi/internal/service/payment/wechat"
	"archi/pkg/logger"
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

//go:generate mockgen -source=./types.go -destination=mocks/reward.mock.go -package=svcmocks RewardService
type RewardService interface {
	// PreReward 准备打赏，
	// 你也可以直接理解为对标到创建一个打赏的订单
	// 因为目前我们只支持微信扫码支付，所以实际上直接把接口定义成这个样子就可以了
	PreReward(ctx context.Context, r domain.Reward) (domain.CodeURL, error)
	GetReward(ctx context.Context, rid, uid int64) (domain.Reward, error)
	UpdateReward(ctx context.Context, bizTradeNO string, status domain.RewardStatus) error
}
type WechatNativeRewardService struct {
	rewardRepo repository.RewardRepository
	l          logger.Logger
	paymentSvc wechat.PaymentService
	accountSvc AccountService
}

func (s *WechatNativeRewardService) PreReward(ctx context.Context, r domain.Reward) (domain.CodeURL, error) {
	// 先查询缓存，确认是否已经创建过了打赏的预支付订单
	codeUrl, err := s.rewardRepo.GetCachedCodeURL(ctx, r)
	if err == nil {
		return codeUrl, nil
	}
	r.Status = domain.RewardStatusInit
	rid, err := s.rewardRepo.CreateReward(ctx, r)
	if err != nil {
		return domain.CodeURL{}, err
	}
	resp, err := s.paymentSvc.Prepay(ctx, domain.Payment{
		Amt: domain.Amount{
			Total:    r.Amt,
			Currency: "CNY",
		},
		BizTradeNO:  fmt.Sprintf("reward-%d", rid),
		Description: fmt.Sprintf("打赏-%s", r.Target.BizName),
	})

	if err != nil {
		return domain.CodeURL{}, err
	}
	cu := domain.CodeURL{
		Rid: rid,
		URL: resp,
	}
	err1 := s.rewardRepo.CachedCodeURL(ctx, cu, r)
	if err1 != nil {
		s.l.Error("缓存二维码失败", logger.Error(err1))
	}
	return cu, err
}
func (s *WechatNativeRewardService) GetReward(ctx context.Context, rid, uid int64) (domain.Reward, error) {
	// 快路径
	r, err := s.rewardRepo.GetReward(ctx, rid)
	if err != nil {
		return domain.Reward{}, err
	}
	if r.Uid != uid {
		// 说明是非法查询
		return domain.Reward{}, errors.New("查询的打赏记录和打赏人对不上")
	}
	// 已经是完结状态
	if r.Completed() {
		return r, nil
	}
	// 这个时候，考虑到支付到查询结果，我们搞一个慢路径
	resp, err := s.paymentSvc.GetPayment(ctx, s.bizTradeNO(r.ID))
	if err != nil {
		// 这边我们直接返回从数据库查询的数据
		s.l.Error("慢路径查询支付结果失败",
			logger.Int64("rid", r.ID), logger.Error(err))
		return r, nil
	}
	// 更新状态
	switch resp.Status {
	case domain.PaymentStatusFailed:
		r.Status = domain.RewardStatusFailed
	case domain.PaymentStatusInit:
		r.Status = domain.RewardStatusInit
	case domain.PaymentStatusSuccess:
		r.Status = domain.RewardStatusPayed
	case domain.PaymentStatusRefund:
		// 理论上来说不可能出现这个，直接设置为失败
		r.Status = domain.RewardStatusFailed
	default:
		r.Status = domain.RewardStatusUnknown
	}
	err = s.rewardRepo.UpdateStatus(ctx, rid, r.Status)
	if err != nil {
		s.l.Error("更新本地打赏状态失败",
			logger.Int64("rid", r.ID), logger.Error(err))
		return r, nil
	}
	return r, nil
}
func (s *WechatNativeRewardService) bizTradeNO(rid int64) string {
	return fmt.Sprintf("reward-%d", rid)
}
func (s *WechatNativeRewardService) UpdateReward(ctx context.Context, bizTradeNO string, status domain.RewardStatus) error {
	rid := s.toRid(bizTradeNO)
	err := s.rewardRepo.UpdateStatus(ctx, rid, status)
	if err != nil {
		return err
	}
	// 完成了支付，准备入账
	if status == domain.RewardStatusPayed {
		r, err := s.rewardRepo.GetReward(ctx, rid)
		if err != nil {
			return err
		}
		// webook 抽成
		weAmt := int64(float64(r.Amt) * 0.1)
		err = s.accountSvc.Credit(ctx, domain.Credit{
			Biz:   "reward",
			BizID: rid,
			Items: []domain.CreditItem{
				{
					AccountType: domain.AccountTypeSystem,
					// 虽然可能为 0，但是也要记录出来
					Amt:      weAmt,
					Currency: "CNY",
				},
				{
					AccountID:   r.Uid,
					Uid:         r.Uid,
					AccountType: domain.AccountTypeReward,
					Amt:         r.Amt - weAmt,
					Currency:    "CNY",
				},
			},
		})
		if err != nil {
			s.l.Error("入账失败了，快来修数据啊！！！",
				logger.String("biz_trade_no", bizTradeNO),
				logger.Error(err))
			// 做好监控和告警，这里
			return err
		}
	}
	return nil
}
func (s *WechatNativeRewardService) toRid(tradeNO string) int64 {
	ridStr := strings.Split(tradeNO, "-")
	val, _ := strconv.ParseInt(ridStr[1], 10, 64)
	return val
}
