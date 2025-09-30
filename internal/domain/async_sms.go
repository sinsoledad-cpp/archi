package domain

type AsyncSMS struct {
	ID       int64
	TplId    string
	Args     []string
	Numbers  []string
	RetryMax int // 重试的配置
}
