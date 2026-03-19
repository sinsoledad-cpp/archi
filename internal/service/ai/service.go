package ai

import (
	"archi/internal/domain"
	"archi/internal/repository"
	"context"
	"fmt"

	"github.com/cloudwego/eino/schema"
)

// AiService 负责 AI 业务逻辑的调度与缓存编排
type AiService interface {
	GetArticleSummary(ctx context.Context, art domain.Article) (domain.ArticleSummary, error)
	// AnswerQuestionStream 支持流式返回，提供“打字机”效果
	AnswerQuestionStream(ctx context.Context, artId int64, content string, question string) (*schema.StreamReader[any], error)
	// AuthorHelperStream 创作者助手 Agent 流式入口
	AuthorHelperStream(ctx context.Context, input AuthorHelperInput) (*schema.StreamReader[any], error)
}

type aiService struct {
	provider *AiProvider
	repo     repository.AiRepository
}

func NewAiService(p *AiProvider, r repository.AiRepository) AiService {
	return &aiService{
		provider: p,
		repo:     r,
	}
}

// GetArticleSummary 获取文章课代表总结
func (s *aiService) GetArticleSummary(ctx context.Context, art domain.Article) (domain.ArticleSummary, error) {
	// 1. 优先查缓存 (Cache-Aside)
	//summary, err := s.repo.GetArticleSummary(ctx, art.ID)
	//if err == nil {
	//	return summary, nil
	//}

	// 2. 缓存未命中，调用 AI Provider 获取预编译好的 Graph/Chain
	runnable := s.provider.Get(domain.SceneArticleSummary)
	if runnable == nil {
		return domain.ArticleSummary{}, fmt.Errorf("ai scene %s is not registered", domain.SceneArticleSummary)
	}

	// 3. 执行编排逻辑
	res, err := runnable.Invoke(ctx, art)
	if err != nil {
		return domain.ArticleSummary{}, fmt.Errorf("ai summary generation failed: %w", err)
	}

	// 类型断言并处理结果
	summaryRes, ok := res.(domain.ArticleSummary)
	if !ok {
		return domain.ArticleSummary{}, fmt.Errorf("invalid output type from ai summary")
	}

	// 4. 异步回写缓存，不阻塞主流程
	go func() {
		// 使用 Background Context，避免请求结束后协程被取消
		_ = s.repo.SetArticleSummary(context.Background(), art.ID, summaryRes)
	}()

	return summaryRes, nil
}

// AnswerQuestionStream 实现针对单篇文章的“笔记问答”
func (s *aiService) AnswerQuestionStream(ctx context.Context, artId int64, content string, question string) (*schema.StreamReader[any], error) {
	// 1. 获取执行器
	runnable := s.provider.Get(domain.SceneArticleQA)
	if runnable == nil {
		return nil, fmt.Errorf("ai scene %s is not registered", domain.SceneArticleQA)
	}

	// 2. 构造输入 DTO
	input := ArticleQAInput{
		ArticleID: artId,
		Content:   content,
		Question:  question,
	}

	// 3. 调用流式执行接口
	return runnable.Stream(ctx, input)
}

// AuthorHelperStream 创作者助手 Agent 入口
func (s *aiService) AuthorHelperStream(ctx context.Context, input AuthorHelperInput) (*schema.StreamReader[any], error) {
	// 1. 获取 Agent 执行器
	runnable := s.provider.Get(domain.SceneAuthorHelper)
	if runnable == nil {
		return nil, fmt.Errorf("ai scene %s is not registered", domain.SceneAuthorHelper)
	}

	// 2. 调用流式执行接口
	// ADK Agent 的流会包含 Thought 和 ToolCall 消息，建议在 Web 层进行过滤或全量下发
	return runnable.Stream(ctx, input)
}
