package main

import (
	"context"
	"fmt"
	"log"

	"archi/ioc"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/joho/godotenv"
)

func main() {
	// .env 文件通常位于项目根目录。此脚本需要从项目根目录执行，
	// 以便能够正确加载 .env 文件和 Go 模块。
	// 例如: 在 e:\Code\GolandProjects\archi 目录下运行 `go run ./cmd/test/ai.go`
	err := godotenv.Load() // 默认从当前工作目录加载 .env
	if err != nil {
		log.Fatalf("加载 .env 文件失败。请确保您在项目根目录下执行此命令，并且 .env 文件存在。错误: %v", err)
	}

	// 1. 初始化 AI 模型
	fmt.Println("Initializing AI model...")
	model := ioc.InitVolcanoModel()
	fmt.Println("AI model initialized successfully.")

	// 2. 将模型封装成一个可执行的 Runnable Chain
	// eino 的组件通常通过编排链(Chain)来执行，而不是直接调用
	chain := compose.NewChain[[]*schema.Message, any]()
	chain.AppendChatModel(model)
	runnable, err := chain.Compile(context.Background())
	if err != nil {
		log.Fatalf("Failed to compile chain: %v", err)
	}

	// 3. 准备输入消息
	messages := []*schema.Message{
		{
			Role:    "user",
			Content: "你好，请你用中文介绍一下自己，并说明你基于什么模型。",
		},
	}

	// 4. 通过 Runnable 的 Invoke 方法调用模型
	fmt.Println("\nSending message to AI model...")
	respAny, err := runnable.Invoke(context.Background(), messages)
	if err != nil {
		log.Fatalf("Failed to call AI model: %v", err)
	}

	// 5. Invoke 的返回类型是 any，需要进行类型断言
	resp, ok := respAny.(*schema.Message)
	if !ok {
		log.Fatalf("AI model returned an unexpected type: %T", respAny)
	}

	// 6. 打印模型返回的结果
	fmt.Println("\n========= AI Response ===========")
	fmt.Println(resp.Content)
	fmt.Println("================================")
}
