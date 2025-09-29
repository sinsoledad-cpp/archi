package domain

import "time"

const (
	// ArticleStatusUnknown 这是一个未知状态
	ArticleStatusUnknown = iota
	// ArticleStatusUnpublished 未发表
	ArticleStatusUnpublished
	// ArticleStatusPublished 已发表
	ArticleStatusPublished
	// ArticleStatusPrivate 仅自己可见
	ArticleStatusPrivate
)

type ArticleStatus uint8

func (s ArticleStatus) ToUint8() uint8 {
	return uint8(s)
}

type Author struct {
	ID   int64
	Name string
}

type Article struct {
	ID      int64
	Title   string
	Content string
	Author  Author
	Status  ArticleStatus
	Ctime   time.Time
	Utime   time.Time
}

func (a Article) Abstract() string {
	str := []rune(a.Content)
	// 只取部分作为摘要
	if len(str) > 128 {
		str = str[:128]
	}
	return string(str)
}
