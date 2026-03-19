package ai

// ArticleQAInput 针对文章问答场景的输入 DTO
type ArticleQAInput struct {
	ArticleID int64  `json:"article_id"`
	Content   string `json:"content"`  // 文章全文
	Question  string `json:"question"` // 用户提问
}
