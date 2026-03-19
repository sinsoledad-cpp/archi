package ai

// ArticleQAInput 针对文章问答场景的输入 DTO
type ArticleQAInput struct {
	ArticleID int64  `json:"article_id"`
	Content   string `json:"content"`  // 文章全文
	Question  string `json:"question"` // 用户提问
}

// AuthorHelperInput 针对创作者助手的输入 DTO
type AuthorHelperInput struct {
	ArticleID   int64  `json:"article_id"`
	AuthorID    int64  `json:"author_id"`   // 用于工具校验
	Content     string `json:"content"`     // 当前编辑器的实时内容 (可选)
	Instruction string `json:"instruction"` // 用户指令 (如: "参考我之前的风格润色")
}
