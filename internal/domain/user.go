package domain

import "time"

type User struct {
	ID         int64
	Email      string
	Password   string
	Nickname   string
	Avatar     string    // 头像
	Birthday   time.Time // YYYY-MM-DD
	AboutMe    string
	Phone      string
	Ctime      time.Time // UTC 0 的时区
	WechatInfo WechatInfo
	//Addr Address
}
type WechatInfo struct {
	UnionID string
	OpenID  string
}
