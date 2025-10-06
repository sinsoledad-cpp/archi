package search

import (
	"archi/internal/domain"
	"archi/internal/repository/dao/search"
	"context"
	"github.com/ecodeclub/ekit/slice"
)

type UserRepository interface {
	InputUser(ctx context.Context, msg domain.UserES) error
	SearchUser(ctx context.Context, keywords []string) ([]domain.UserES, error)
}

type DefaultUserRepository struct {
	dao search.UserDAO
}

func NewDefaultUserRepository(d search.UserDAO) UserRepository {
	return &DefaultUserRepository{
		dao: d,
	}
}
func (u *DefaultUserRepository) InputUser(ctx context.Context, msg domain.UserES) error {
	return u.dao.InputUser(ctx, search.User{
		Id:       msg.Id,
		Email:    msg.Email,
		Nickname: msg.Nickname,
		Phone:    msg.Phone,
	})
}

func (u *DefaultUserRepository) SearchUser(ctx context.Context, keywords []string) ([]domain.UserES, error) {
	users, err := u.dao.Search(ctx, keywords)
	if err != nil {
		return nil, err
	}
	return slice.Map(users, func(idx int, src search.User) domain.UserES {
		return domain.UserES{
			Id:       src.Id,
			Email:    src.Email,
			Nickname: src.Nickname,
			Phone:    src.Phone,
		}
	}), nil
}
