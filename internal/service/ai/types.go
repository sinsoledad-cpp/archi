package ai

import (
	"archi/internal/domain"
)

// aiTask 内部任务包装
type aiTask struct {
	scene domain.Scene
	input any
}

// aiMetadata AI 执行元数据
type aiMetadata struct {
	tokenUsage int
	modelName  string
}
