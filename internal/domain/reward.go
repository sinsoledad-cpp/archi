package domain

const (
	RewardStatusUnknown RewardStatus = iota
	RewardStatusInit
	RewardStatusPayed
	RewardStatusFailed
)

type RewardStatus uint8

func (r RewardStatus) AsUint8() uint8 {
	return uint8(r)
}

type CodeURL struct {
	Rid int64  //打赏ID (Reward ID)
	URL string //二维码链接
}

type Target struct {
	Biz     string // 因为什么而打赏
	BizID   int64
	BizName string // 作为一个可选的东西// 也就是你要打赏的东西是什么
	Uid     int64  // 打赏的目标用户
}

type Reward struct {
	ID     int64
	Uid    int64
	Target Target
	Amt    int64 // 同样不着急引入货币。
	Status RewardStatus
}

// Completed 是否已经完成
// 目前来说，也就是是否处理了支付回调
func (r Reward) Completed() bool {
	return r.Status == RewardStatusFailed || r.Status == RewardStatusPayed
}
