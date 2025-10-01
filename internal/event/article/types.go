package article

// TopicReadEvent 是文章阅读事件的 Kafka Topic 名称。
const TopicReadEvent = "article_read"

// ReadEvent 代表一个文章阅读事件。
// 当文章被成功阅读时，由生产者发出。
type ReadEvent struct {
	Aid int64 `json:"aid"` // Article ID
	Uid int64 `json:"uid"` // User ID
}
