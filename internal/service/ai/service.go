package ai

import (
	"archi/internal/domain"
	"archi/internal/repository"
	"context"
	"fmt"
)

// AiService 负责 AI 业务逻辑的调度与缓存编排
type AiService interface {
	GetArticleSummary(ctx context.Context, art domain.Article) (domain.ArticleSummary, error)
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
	summary, err := s.repo.GetArticleSummary(ctx, art.ID)
	if err == nil {
		return summary, nil
	}

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
