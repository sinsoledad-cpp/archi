package domain

const (
	AccountTypeUnknown AccountType = iota
	AccountTypeReward
	AccountTypeSystem
)

type AccountType uint8

func (a AccountType) AsUint8() uint8 {
	return uint8(a)
}

// Credit 增加余额
type Credit struct {
	Biz   string
	BizId int64
	Items []CreditItem
}

// CreditItem 记账条目
type CreditItem struct {
	Uid         int64 //给哪个用户的
	AccountID   int64
	AccountType AccountType
	Amt         int64  //金额 (Amount)
	Currency    string //币种
}
