package search

import (
	"archi/internal/repository/dao/search"
	"context"
)

type AnyRepository interface {
	Input(ctx context.Context, index string, docID string, data string) error
}

type DefaultAnyRepository struct {
	dao search.AnyDAO
}

func NewDefaultAnyRepository(dao search.AnyDAO) AnyRepository {
	return &DefaultAnyRepository{dao: dao}
}

func (repo *DefaultAnyRepository) Input(ctx context.Context, index string, docID string, data string) error {
	return repo.dao.Input(ctx, index, docID, data)
}
