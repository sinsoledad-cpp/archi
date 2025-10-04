package payment

import "archi/internal/domain"

type PaymentEvent struct {
	BizTradeNO string
	Status     uint8
}

func (PaymentEvent) Topic() string {
	return "payment_events"
}
func (p PaymentEvent) ToDomainStatus() domain.RewardStatus {
	// 	PaymentStatusInit
	//	PaymentStatusSuccess
	//	PaymentStatusFailed
	//	PaymentStatusRefund
	switch p.Status {
	// 这里不能引用 payment 里面的定义，只能手写
	case 1:
		return domain.RewardStatusInit
	case 2:
		return domain.RewardStatusPayed
	case 3, 4:
		return domain.RewardStatusFailed
	default:
		return domain.RewardStatusUnknown
	}
}
