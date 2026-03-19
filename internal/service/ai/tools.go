package ai

import (
	"archi/internal/repository"
	"context"
	"fmt"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// initAuthorTools 初始化创作者助手的工具集
func (f *AiFactory) initAuthorTools(repo repository.ArticleRepository) []tool.BaseTool {
	// 工具 1: 获取指定文章的内容 (用于后端自取)
	tool1, _ := utils.InferTool("get_article_content", "获取指定文章的标题和全文内容",
		func(ctx context.Context, input struct {
			ArticleID int64 `json:"article_id" label:"文章ID" jsonschema_description:"需要获取内容的文章ID，如果获取当前正在编辑的文章内容，可传0"`
		}) (string, error) {
			targetID := input.ArticleID
			// 如果没传 ID，则默认从 Session 获取当前正在处理的文章 ID
			if targetID <= 0 {
				val, ok := adk.GetSessionValue(ctx, "article_id")
				if !ok {
					return "", fmt.Errorf("article_id not provided and not found in session")
				}
				targetID = val.(int64)
			}

			art, err := repo.GetById(ctx, targetID)
			if err != nil {
				return "", fmt.Errorf("failed to get article content for ID %d: %w", targetID, err)
			}

			// 从 Session 获取 AuthorID 进行权限校验
			uid, ok := adk.GetSessionValue(ctx, "author_id")
			if !ok {
				return "", fmt.Errorf("author_id not found in session")
			}

			// 校验权限：只能获取自己的文章内容
			if art.Author.ID != uid.(int64) {
				return "", fmt.Errorf("permission denied: you are not the author of this article (target: %d, current_user: %d)", targetID, uid)
			}
			return fmt.Sprintf("标题: %s\n内容: %s", art.Title, art.Content), nil
		})

	// 工具 2: 检索作者的历史文章列表 (用于风格参考)
	tool2, _ := utils.InferTool("search_author_articles", "检索当前作者的历史文章列表，返回文章ID和标题",
		func(ctx context.Context, input struct {
			Limit int `json:"limit" label:"数量" jsonschema_description:"检索的文章数量，默认为5"`
		}) (string, error) {
			if input.Limit <= 0 {
				input.Limit = 5
			}

			// 从 Session 获取 AuthorID
			uid, ok := adk.GetSessionValue(ctx, "author_id")
			if !ok {
				return "", fmt.Errorf("author_id not found in session")
			}

			// 获取作者的文章列表 (假设 offset 0)
			arts, err := repo.GetByAuthor(ctx, uid.(int64), 0, input.Limit)
			if err != nil {
				return "", fmt.Errorf("failed to list author articles: %w", err)
			}

			var res string
			for _, art := range arts {
				res += fmt.Sprintf("- ID: %d, 标题: %s\n", art.ID, art.Title)
			}
			if res == "" {
				return "该作者暂无其他历史文章", nil
			}
			return "作者历史文章列表：\n" + res, nil
		})

	// 工具 3: 获取热榜趋势
	tool3, _ := utils.InferTool("get_hot_trends", "获取当前全站最热门的文章列表，帮助创作者把握趋势",
		func(ctx context.Context, input struct {
			Limit int `json:"limit" label:"数量" jsonschema_description:"获取热门文章的数量，默认10"`
		}) (string, error) {
			if input.Limit <= 0 {
				input.Limit = 10
			}
			arts, err := f.rankSvc.GetTopN(ctx)
			if err != nil {
				return "", fmt.Errorf("failed to get hot trends: %w", err)
			}

			if len(arts) > input.Limit {
				arts = arts[:input.Limit]
			}

			var res string
			for i, art := range arts {
				res += fmt.Sprintf("%d. 标题: %s (ID: %d)\n", i+1, art.Title, art.ID)
			}
			return "当前热门趋势文章：\n" + res, nil
		})

	// 工具 4: 获取文章交互统计
	tool4, _ := utils.InferTool("get_article_stats", "获取指定文章的点赞、阅读、收藏等交互统计数据",
		func(ctx context.Context, input struct {
			ArticleID int64 `json:"article_id" label:"文章ID" jsonschema_description:"需要查询统计数据的文章ID，如果查询当前正在编辑的文章，可传0"`
		}) (string, error) {
			targetID := input.ArticleID
			// 同样支持从 Session 获取当前 ID
			if targetID <= 0 {
				val, ok := adk.GetSessionValue(ctx, "article_id")
				if !ok {
					return "", fmt.Errorf("article_id not provided and not found in session")
				}
				targetID = val.(int64)
			}

			// 获取交互数据
			intr, err := f.intrSvc.Get(ctx, "article", targetID, 0)
			if err != nil {
				return "", fmt.Errorf("failed to get article stats for ID %d: %w", targetID, err)
			}

			return fmt.Sprintf("文章统计数据 (ID: %d):\n- 阅读数: %d\n- 点赞数: %d\n- 收藏数: %d",
				targetID, intr.ReadCnt, intr.LikeCnt, intr.CollectCnt), nil
		})

	return []tool.BaseTool{tool1, tool2, tool3, tool4}
}
