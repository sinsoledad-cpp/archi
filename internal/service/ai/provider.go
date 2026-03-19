package ai

import (
	"archi/internal/domain"
	"sync"

	"github.com/cloudwego/eino/compose"
)

// AiProvider 统一管理所有已编译的 AI 编排实例
type AiProvider struct {
	mu        sync.RWMutex
	runnables map[domain.Scene]compose.Runnable[any, any]
}

func NewAiProvider() *AiProvider {
	return &AiProvider{
		runnables: make(map[domain.Scene]compose.Runnable[any, any]),
	}
}

func (p *AiProvider) Register(scene domain.Scene, r compose.Runnable[any, any]) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.runnables[scene] = r
}

func (p *AiProvider) Get(scene domain.Scene) compose.Runnable[any, any] {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.runnables[scene]
}
