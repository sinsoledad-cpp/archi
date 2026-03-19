package ai

import (
	"archi/internal/domain"
	"archi/internal/repository"
	"archi/internal/service"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// AiFactory 负责根据场景构建所有的 Eino 编排实例
type AiFactory struct {
	chatModel model.ToolCallingChatModel
	artRepo   repository.ArticleRepository
	rankSvc   service.RankingService
	intrSvc   service.InteractiveService
}

func NewAiFactory(m model.ToolCallingChatModel, artRepo repository.ArticleRepository, rankSvc service.RankingService, intrSvc service.InteractiveService) *AiFactory {
	return &AiFactory{
		chatModel: m,
		artRepo:   artRepo,
		rankSvc:   rankSvc,
		intrSvc:   intrSvc,
	}
}

// Create 根据场景创建对应的编排链或图
func (f *AiFactory) Create(scene domain.Scene) (compose.Runnable[any, any], error) {
	switch scene {
	case domain.SceneArticleSummary:
		return f.buildSummaryChain()
	case domain.SceneArticleQA:
		return f.buildArticleQAChain()
	case domain.SceneAuthorHelper:
		return f.buildAuthorHelperAgent()
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

// buildArticleQAChain 构建“沉浸式笔记问答”编排链 (Long Context 直接注入方案)
func (f *AiFactory) buildArticleQAChain() (compose.Runnable[any, any], error) {
	// 1. 定义编排链
	chain := compose.NewChain[any, any]()

	chain.
		// 第一步：格式化输入，构建带 Long Context 的 Prompt
		AppendLambda(compose.InvokableLambda(func(ctx context.Context, input any) ([]*schema.Message, error) {
			var qaInput ArticleQAInput
			switch v := input.(type) {
			case ArticleQAInput:
				qaInput = v
			case *ArticleQAInput:
				if v == nil {
					return nil, fmt.Errorf("input QA input pointer is nil")
				}
				qaInput = *v
			default:
				return nil, fmt.Errorf("invalid input type for QA chain: %T", input)
			}

			// 构造 System Prompt，直接注入文章全文作为上下文
			systemPrompt := fmt.Sprintf(`你是一位博学且细心的文章助手。以下是用户正在阅读的文章全文：
---
%s
---
请严格基于以上内容回答用户的提问。
规则：
1. 如果文章中没有提到相关信息，请回答：“抱歉，在本文中没有找到相关信息。”
2. 请使用简洁且有亲和力的语气。
3. 必要时使用 Markdown 格式（如加粗或列表）使回答更易读。`, qaInput.Content)

			return []*schema.Message{
				schema.SystemMessage(systemPrompt),
				schema.UserMessage(qaInput.Question),
			}, nil
		})).
		// 第二步：调用模型 (支持流式输出)
		AppendChatModel(f.chatModel).
		// 第三步：后处理，将 *schema.Message 转换为文本内容 (如果是流式，这一步会被 Eino 自动处理或透传)
		AppendLambda(compose.InvokableLambda(func(ctx context.Context, msg *schema.Message) (any, error) {
			return msg.Content, nil
		}))

	// 2. 编译并返回
	return chain.Compile(context.Background())
}

// buildAuthorHelperAgent 使用 Eino ADK 构建创作者助手 Agent
func (f *AiFactory) buildAuthorHelperAgent() (compose.Runnable[any, any], error) {
	const systemPrompt = `你是一位专业的“创作者 AI 助手”。
你的目标是帮助创作者优化文章内容、提供创意建议或参考其历史写作风格。

你可以使用的工具：
1. get_article_content: 当你需要获取当前文章或某篇特定文章的全文时使用。
2. search_author_articles: 当你需要参考作者的历史写作风格或查找其相关旧文时使用。
3. get_hot_trends: 当创作者询问当前热门话题或需要灵感建议时使用。
4. get_article_stats: 当创作者需要了解某篇文章的阅读、点赞等反馈数据时使用。

工作流程：
- 如果用户指令涉及“润色”、“续写”或“改写”当前文章，但你没有看到内容，请先调用 get_article_content。
- 如果用户要求“模仿我的风格”或“参考我之前的文章”，请先调用 search_author_articles 查找相关文章。
- 如果用户询问“最近什么火”或需要“选题建议”，请调用 get_hot_trends。
- 始终以专业、鼓励且具有建设性的语气与创作者交流。`

	// 1. 创建 ChatModelAgent
	authorAgent, err := adk.NewChatModelAgent(context.Background(), &adk.ChatModelAgentConfig{
		Name:        "AuthorHelper",
		Description: "创作者 AI 助手，支持内容润色、风格模仿和自取内容",
		Instruction: systemPrompt,
		Model:       f.chatModel,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: f.initAuthorTools(f.artRepo),
			},
		},
		MaxIterations: 10,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create ADK agent: %w", err)
	}

	// 2. 将 ADK Agent 封装为 compose.Runnable[any, any]
	// 这里通过自定义结构体实现，以同时支持 Invoke 和 Stream
	return &authorHelperRunnable{
		agent: authorAgent,
	}, nil
}

// authorHelperRunnable 封装 ADK Agent 为 Eino Runnable 接口
type authorHelperRunnable struct {
	agent adk.Agent
}

func (r *authorHelperRunnable) Invoke(ctx context.Context, input any, opts ...compose.Option) (any, error) {
	qaInput, err := r.parseInput(input)
	if err != nil {
		return nil, err
	}

	userMsg := r.formatUserMsg(qaInput)
	runner := adk.NewRunner(ctx, adk.RunnerConfig{Agent: r.agent})
	iter := runner.Run(ctx, []adk.Message{{Role: schema.User, Content: userMsg}}, adk.WithSessionValues(map[string]any{
		"article_id": qaInput.ArticleID,
		"author_id":  qaInput.AuthorID,
	}))

	var finalContent string
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		msg, _, err := adk.GetMessage(event)
		if err == nil && msg.Content != "" {
			finalContent = msg.Content
		}
	}
	return finalContent, nil
}

func (r *authorHelperRunnable) Stream(ctx context.Context, input any, opts ...compose.Option) (*schema.StreamReader[any], error) {
	qaInput, err := r.parseInput(input)
	if err != nil {
		return nil, err
	}

	userMsg := r.formatUserMsg(qaInput)
	runner := adk.NewRunner(ctx, adk.RunnerConfig{Agent: r.agent})
	iter := runner.Run(ctx, []adk.Message{{Role: schema.User, Content: userMsg}}, adk.WithSessionValues(map[string]any{
		"article_id": qaInput.ArticleID,
		"author_id":  qaInput.AuthorID,
	}))

	pipeReader, pipeWriter := schema.Pipe[any](10)
	go func() {
		defer pipeWriter.Close()
		for {
			event, ok := iter.Next()
			if !ok {
				break
			}
			msg, _, err := adk.GetMessage(event)
			if err == nil && msg.Content != "" {
				_ = pipeWriter.Send(msg.Content, nil)
			}
		}
	}()

	return pipeReader, nil
}

func (r *authorHelperRunnable) Collect(ctx context.Context, reader *schema.StreamReader[any], opts ...compose.Option) (any, error) {
	// 简单的流聚合实现
	defer reader.Close()
	var finalContent string
	for {
		chunk, err := reader.Recv()
		if err != nil {
			break
		}
		if s, ok := chunk.(string); ok {
			finalContent += s
		}
	}
	return finalContent, nil
}

func (r *authorHelperRunnable) Transform(ctx context.Context, reader *schema.StreamReader[any], opts ...compose.Option) (*schema.StreamReader[any], error) {
	// 简单的流转换实现：将流聚合后再调用 Invoke
	content, err := r.Collect(ctx, reader, opts...)
	if err != nil {
		return nil, err
	}

	// 重新封装为流返回
	pipeReader, pipeWriter := schema.Pipe[any](1)
	go func() {
		defer pipeWriter.Close()
		_ = pipeWriter.Send(content, nil)
	}()
	return pipeReader, nil
}
func (r *authorHelperRunnable) parseInput(input any) (AuthorHelperInput, error) {
	var qaInput AuthorHelperInput
	switch v := input.(type) {
	case AuthorHelperInput:
		qaInput = v
	case *AuthorHelperInput:
		if v == nil {
			return qaInput, fmt.Errorf("input is nil")
		}
		qaInput = *v
	default:
		return qaInput, fmt.Errorf("invalid input type: %T", input)
	}
	return qaInput, nil
}

func (r *authorHelperRunnable) formatUserMsg(input AuthorHelperInput) string {
	userMsg := fmt.Sprintf("指令: %s\n当前文章ID: %d", input.Instruction, input.ArticleID)
	if input.Content != "" {
		userMsg += fmt.Sprintf("\n当前内容: %s", input.Content)
	}
	return userMsg
}
