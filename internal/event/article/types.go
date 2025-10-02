package article

const TopicReadEvent = "article_read"

type ReadEvent struct {
	Aid int64
	Uid int64
}

type BatchReadEvent struct {
	Aids []int64
	Uids []int64
}
