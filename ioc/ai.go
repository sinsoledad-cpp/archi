package ioc

import (
	"archi/internal/domain"
	"archi/internal/service/ai"
	"context"
	"fmt"
	"os"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/components/model"
)

func InitAiProvider(factory *ai.AiFactory) *ai.AiProvider {
	provider := ai.NewAiProvider()

	// 在这里统一注册所有 AI 场景
	// 场景：文章课代表总结
	summaryRunnable, err := factory.Create(domain.SceneArticleSummary)
	if err != nil {
		panic(fmt.Sprintf("failed to create ai scene %s: %v", domain.SceneArticleSummary, err))
	}
	provider.Register(domain.SceneArticleSummary, summaryRunnable)

	return provider
}

func InitVolcanoModel() model.ToolCallingChatModel {
	// 直接使用 os.Getenv 读取环境变量
	apiKey := os.Getenv("ARK_API_KEY")
	modelId := os.Getenv("ARK_MODEL_ID")

	if apiKey == "" {
		panic("ARK_API_KEY is not set in environment variables")
	}
	if modelId == "" {
		// 如果没有设置，给一个默认值
		modelId = "doubao-1.5-pro-32k-250115"
	}

	chatModel, err := ark.NewChatModel(context.Background(), &ark.ChatModelConfig{
		APIKey: apiKey,
		Model:  modelId,
	})
	if err != nil {
		panic(err)
	}
	return chatModel
}
