package service

import (
	"archi/internal/domain"
	"archi/internal/repository/search"
	"context"
)

type SyncService interface {
	InputArticle(ctx context.Context, article domain.ArticleES) error
	InputUser(ctx context.Context, user domain.UserES) error
	InputAny(ctx context.Context, index, docID, data string) error
}

type DefaultSyncService struct {
	userRepo    search.UserRepository
	articleRepo search.ArticleRepository
	anyRepo     search.AnyRepository
}

func NewDefaultSyncService(anyRepo search.AnyRepository, userRepo search.UserRepository, articleRepo search.ArticleRepository) SyncService {
	return &DefaultSyncService{
		userRepo:    userRepo,
		articleRepo: articleRepo,
		anyRepo:     anyRepo,
	}
}

func (s *DefaultSyncService) InputArticle(ctx context.Context, article domain.ArticleES) error {
	return s.articleRepo.InputArticle(ctx, article)
}

func (s *DefaultSyncService) InputUser(ctx context.Context, user domain.UserES) error {
	return s.userRepo.InputUser(ctx, user)
}

func (s *DefaultSyncService) InputAny(ctx context.Context, index, docID, data string) error {
	return s.anyRepo.Input(ctx, index, docID, data)
}
