package main

import (
	"archi/internal/domain"
	"archi/internal/service/ai"
	"archi/setting"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/joho/godotenv"
)

//TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>

func main() {
	// 在程序最开始加载 .env 文件
	err := godotenv.Load()
	if err != nil {
		// 如果 .env 文件不存在，可以忽略错误，因为我们可能通过真实环境变量注入
		log.Println("Warning: .env file not found, using system environment variables")
	}

	setting.InitViper()
	setting.InitValidate()
	setting.InitPrometheus()
	tpCancel := setting.InitOTEL()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		tpCancel(ctx)
	}()

	app := InitApp()

	// 注册 AI 场景
	registerAiScenes(app.aiFactory, app.aiProvider)

	for _, c := range app.consumers {
		err := c.Start()
		if err != nil {
			panic(err)
		}
	}

	app.cron.Start()
	defer func() {
		<-app.cron.Stop().Done()
	}()

	server := app.engine
	if err := server.Run(":8080"); err != nil {
		return
	}
}

func registerAiScenes(factory *ai.AiFactory, provider *ai.AiProvider) {
	// 场景：文章课代表总结
	summaryRunnable, err := factory.Create(domain.SceneArticleSummary)
	if err != nil {
		panic(fmt.Sprintf("failed to create ai scene %s: %v", domain.SceneArticleSummary, err))
	}
	provider.Register(domain.SceneArticleSummary, summaryRunnable)
}
