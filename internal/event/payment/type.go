package payment

type PaymentEvent struct {
	BizTradeNO string
	Status     uint8
}

func (PaymentEvent) Topic() string {
	return "payment_events"
}
