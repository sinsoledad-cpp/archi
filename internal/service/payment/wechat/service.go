package wechat

import (
	"archi/internal/domain"
	"archi/internal/event/payment"
	"archi/internal/repository"
	"archi/pkg/logger"
	"context"
	"errors"
	"fmt"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"
	"time"
)

type PaymentService interface {
	// Prepay 预支付，对应于微信创建订单的步骤
	Prepay(ctx context.Context, pmt domain.Payment) (string, error)
}

var errUnknownTransactionState = errors.New("未知的微信事务状态")

type NativePaymentService struct {
	nativeApiSvc *native.NativeApiService     // 1. 微信官方API的客户端
	appID        string                       // 2. 应用ID
	mchID        string                       // 3. 商户ID
	notifyURL    string                       // 4. 支付回调地址
	repo         repository.PaymentRepository // 5. 数据库操作的抽象
	l            logger.Logger                // 6. 日志记录器
	producer     payment.Producer             // 7. 事件生产者（消息队列）

	// 在微信 native 里面，分别是
	// SUCCESS：支付成功
	// REFUND：转入退款
	// NOTPAY：未支付
	// CLOSED：已关闭
	// REVOKED：已撤销（付款码支付）
	// USERPAYING：用户支付中（付款码支付）
	// PAYERROR：支付失败(其他原因，如银行返回失败)
	nativeCBTypeToStatus map[string]domain.PaymentStatus // 状态映射表
}

func NewNativePaymentService(svc *native.NativeApiService, repo repository.PaymentRepository, producer payment.Producer, l logger.Logger, appid, mchid string) *NativePaymentService {
	return &NativePaymentService{
		l:            l,
		repo:         repo,
		nativeApiSvc: svc,
		appID:        appid,
		mchID:        mchid,
		notifyURL:    "http://wechat.meoying.com/pay/callback", // 一般来说，这个都是固定的，基本不会变的
		producer:     producer,
		nativeCBTypeToStatus: map[string]domain.PaymentStatus{
			"SUCCESS":  domain.PaymentStatusSuccess,
			"PAYERROR": domain.PaymentStatusFailed,
			"NOTPAY":   domain.PaymentStatusInit,
			"CLOSED":   domain.PaymentStatusFailed,
			"REVOKED":  domain.PaymentStatusFailed,
			"REFUND":   domain.PaymentStatusRefund,
			// 其它状态你都可以加
		},
	}
}

func (n *NativePaymentService) Prepay(ctx context.Context, pmt domain.Payment) (string, error) {
	err := n.repo.AddPayment(ctx, pmt)
	if err != nil {
		return "", err
	}
	resp, _, err := n.nativeApiSvc.Prepay(ctx,
		native.PrepayRequest{
			Appid:       core.String(n.appID),
			Mchid:       core.String(n.mchID),
			Description: core.String(pmt.Description),
			OutTradeNo:  core.String(pmt.BizTradeNO),
			TimeExpire:  core.Time(time.Now().Add(time.Minute * 30)),
			NotifyUrl:   core.String(n.notifyURL),
			Amount: &native.Amount{
				Currency: core.String(pmt.Amt.Currency),
				Total:    core.Int64(pmt.Amt.Total),
			},
		},
	)
	if err != nil {
		return "", err
	}
	// 这里你可以考虑引入另外一个状态，也就是代表你已经调用了第三方支付，正在等回调的状态
	// 但是这个状态意义不是很大。
	// 因为你在考虑兜底（定时比较数据）的时候，不管有没有调用第三方支付，
	// 你都要问一下第三方支付这个
	return *resp.CodeUrl, nil // 返回支付二维码
}

func (n *NativePaymentService) SyncWechatInfo(ctx context.Context, bizTradeNO string) error {
	txn, _, err := n.nativeApiSvc.QueryOrderByOutTradeNo(ctx,
		native.QueryOrderByOutTradeNoRequest{
			OutTradeNo: core.String(bizTradeNO),
			Mchid:      core.String(n.mchID),
		})
	if err != nil {
		return err
	}
	return n.updateByTxn(ctx, txn)
}

func (n *NativePaymentService) FindExpiredPayment(ctx context.Context, offset, limit int, t time.Time) ([]domain.Payment, error) {
	return n.repo.FindExpiredPayment(ctx, offset, limit, t)
}

func (n *NativePaymentService) GetPayment(ctx context.Context, bizTradeId string) (domain.Payment, error) {
	return n.repo.GetPayment(ctx, bizTradeId)
}

func (n *NativePaymentService) HandleCallback(ctx context.Context, txn *payments.Transaction) error {
	return n.updateByTxn(ctx, txn)
}

func (n *NativePaymentService) updateByTxn(ctx context.Context, txn *payments.Transaction) error {
	status, ok := n.nativeCBTypeToStatus[*txn.TradeState]
	if !ok {
		return fmt.Errorf("%w, %s", errUnknownTransactionState, *txn.TradeState)
	}
	pmt := domain.Payment{
		BizTradeNO: *txn.OutTradeNo,
		TxnID:      *txn.TransactionId,
		Status:     status,
	}
	err := n.repo.UpdatePayment(ctx, pmt)
	if err != nil {
		// 这里有一个小问题，就是如果超时了的话，你都不知道更新成功了没
		return err
	}
	// 就是处于结束状态
	err1 := n.producer.ProducePaymentEvent(ctx, payment.PaymentEvent{
		BizTradeNO: pmt.BizTradeNO,
		Status:     pmt.Status.AsUint8(),
	})
	if err1 != nil {
		// 要做好监控和告警
		n.l.Error("发送支付事件失败", logger.Error(err),
			logger.String("biz_trade_no", pmt.BizTradeNO))
	}
	// 虽然发送事件失败，但是数据库记录了，所以可以返回 Nil
	return nil
}
