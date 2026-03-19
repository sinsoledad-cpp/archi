package ioc

import (
	"context"
	"os"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/components/model"
)

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
