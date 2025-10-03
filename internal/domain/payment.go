package domain

const (
	PaymentStatusUnknown PaymentStatus = iota
	PaymentStatusInit
	PaymentStatusSuccess
	PaymentStatusFailed
	PaymentStatusRefund //退款"
)

type PaymentStatus uint8

func (s PaymentStatus) AsUint8() uint8 {
	return uint8(s)
}

type Amount struct {
	Currency string //货币类型
	Total    int64  //金额总数，以最小货币单位存储
}

type Payment struct {
	Amt         Amount
	BizTradeNO  string        // 订单号 代表业务，业务方决定怎么生成，
	Description string        //订单描述信息
	Status      PaymentStatus //支付状态
	TxnID       string        //第三方支付平台返回的交易 ID
}
