package service

import (
	"archi/internal/repository"
	"archi/internal/service/sms"
	"context"
	"errors"
	"fmt"
	"math/rand"
)

var ErrCodeSendTooMany = repository.ErrCodeSendTooMany
var ErrCodeVerifyTooMany = repository.ErrCodeVerifyTooMany
var ErrCodeExpired = repository.ErrCodeExpired

//go:generate mockgen -source=./code.go -package=svcmocks -destination=./mocks/code.mock.go CodeService
type CodeService interface {
	Send(ctx context.Context, biz, phone string) error
	Verify(ctx context.Context, biz, phone, inputCode string) (bool, error)
}

type DefaultCodeService struct {
	repo   repository.CodeRepository
	smsSvc sms.Service
}

func NewDefaultCodeService(repo repository.CodeRepository, smsSvc sms.Service) CodeService {
	return &DefaultCodeService{
		repo:   repo,
		smsSvc: smsSvc,
	}
}

func (svc *DefaultCodeService) Send(ctx context.Context, biz, phone string) error {
	code := svc.generate()
	err := svc.repo.Set(ctx, biz, phone, code)
	// 你在这儿，是不是要开始发送验证码了？
	if err != nil {
		return err
	}
	const codeTplId = "1877556"
	return svc.smsSvc.Send(ctx, codeTplId, []string{code}, phone)
}

func (svc *DefaultCodeService) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	ok, err := svc.repo.Verify(ctx, biz, phone, inputCode)
	if errors.Is(err, repository.ErrCodeVerifyTooMany) || errors.Is(err, repository.ErrCodeExpired) {
		// 相当于，我们对外面屏蔽了验证次数过多的错误，我们就是告诉调用者，你这个不对
		return false, err
	}
	return ok, err
}

func (svc *DefaultCodeService) generate() string {
	// 0-999999
	code := rand.Intn(1000000)
	return fmt.Sprintf("%06d", code)
}
