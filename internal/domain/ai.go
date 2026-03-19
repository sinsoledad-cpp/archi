package domain

// Scene 定义 AI 业务场景
type Scene string

const (
	SceneArticleSummary Scene = "article_summary" // 读者侧：文章总结
	SceneAuthorHelper   Scene = "author_helper"   // 创作者侧：创作助手
)
