package ai

import (
	"archi/internal/domain"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// AiFactory 负责根据场景构建所有的 Eino 编排实例
type AiFactory struct {
	chatModel model.ToolCallingChatModel
}

func NewAiFactory(m model.ToolCallingChatModel) *AiFactory {
	return &AiFactory{chatModel: m}
}

// Create 根据场景创建对应的编排链或图
func (f *AiFactory) Create(scene domain.Scene) (compose.Runnable[any, any], error) {
	switch scene {
	case domain.SceneArticleSummary:
		return f.buildSummaryChain()
	case domain.SceneAuthorHelper:
		// 未来扩展
		return nil, fmt.Errorf("scene %s is not implemented yet", scene)
	default:
		return nil, fmt.Errorf("unknown ai scene: %s", scene)
	}
}

// buildSummaryChain 构建读者侧“文章课代表总结”的线性链
func (f *AiFactory) buildSummaryChain() (compose.Runnable[any, any], error) {
	// 定义 System Prompt，这部分内容是固定的
	const systemPrompt = `你是一位资深的社区内容运营“课代表”。请仔细阅读以下文章，然后完成两个任务：
1. 生成一段 100 字以内、风格类似小红书的精炼种草总结。
2. 提取 2-3 句最能引发读者共鸣或彰显作者观点的“金句”。

请严格按照以下 JSON 格式输出，不要包含任何额外说明：
{"content": "你的总结内容", "golden_sentences": ["金句1", "金句2"]}`

	// 创建一个通用的编排链 (输入为 any，输出为 any)
	chain := compose.NewChain[any, any]()
	chain.
		// 第一步：将输入转换为模型需要的 []*schema.Message
		AppendLambda(compose.InvokableLambda(func(ctx context.Context, input any) ([]*schema.Message, error) {
			var art domain.Article
			switch v := input.(type) {
			case domain.Article:
				art = v
			case *domain.Article:
				if v == nil {
					return nil, fmt.Errorf("input article pointer is nil")
				}
				art = *v
			default:
				return nil, fmt.Errorf("invalid input type for summary chain: %T", input)
			}

			// 手动格式化 Prompt，绕过有问题的 PromptTemplate 组件
			userPrompt := fmt.Sprintf("文章标题: %s\n正文内容: %s", art.Title, art.Content)

			log.Printf("[AI-FACTORY-DEBUG] Manually formatted prompt. Content length: %d", len(userPrompt))

			return []*schema.Message{
				schema.SystemMessage(systemPrompt),
				schema.UserMessage(userPrompt),
			}, nil
		})).
		// 第二步：直接调用模型
		AppendChatModel(f.chatModel).
		// 第三步：解析 JSON 结果并返回
		AppendLambda(compose.InvokableLambda(func(ctx context.Context, msg *schema.Message) (any, error) {
			var res domain.ArticleSummary
			err := json.Unmarshal([]byte(msg.Content), &res)
			return res, err
		}))

	// 3. 编译并返回
	return chain.Compile(context.Background())
}
