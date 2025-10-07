package domain

import (
	"errors"
	"fmt"
	"github.com/ecodeclub/ekit"
	"time"
)

type ExtendFields map[string]string

var errKeyNotFound = errors.New("没有找到对应的 key")

func (f ExtendFields) Get(key string) ekit.AnyValue {
	val, ok := f[key]
	if !ok {
		return ekit.AnyValue{
			Err: fmt.Errorf("%w, key %s", errKeyNotFound),
		}
	}
	return ekit.AnyValue{Val: val}
}

type FeedEvent struct {
	ID int64
	// 以 A 发表了一篇文章为例
	// 如果是 Pull Event，也就是拉模型，那么 Uid 是 A 的id
	// 如果是 Push Event，也就是推模型，那么 Uid 是 A 的某个粉丝的 id
	Uid   int64
	Type  string
	Ctime time.Time
	Ext   ExtendFields
}
