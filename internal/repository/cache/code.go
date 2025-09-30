package cache

import (
	"context"
	_ "embed"
	"errors"
)

var (
	//go:embed lua/set_code.lua
	luaSetCode string
	//go:embed lua/verify_code.lua
	luaVerifyCode string

	ErrCodeSendTooMany   = errors.New("发送太频繁")
	ErrCodeVerifyTooMany = errors.New("验证太频繁")
	ErrCodeExpired       = errors.New("验证码已失效或不存在")
)

//go:generate mockgen -source=./code.go -package=cachemocks -destination=./mocks/code.mock.go CodeCache
type CodeCache interface {
	Set(ctx context.Context, biz, phone, code string) error
	Verify(ctx context.Context, biz, phone, code string) (bool, error)
}
