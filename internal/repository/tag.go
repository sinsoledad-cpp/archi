package repository

import (
	"archi/internal/domain"
	"archi/internal/repository/cache"
	"archi/internal/repository/dao"
	"archi/pkg/logger"
	"context"
	"github.com/ecodeclub/ekit/slice"
	"time"
)

type TagRepository interface {
	CreateTag(ctx context.Context, tag domain.Tag) (int64, error)
	BindTagToBiz(ctx context.Context, uid int64, biz string, bizId int64, tags []int64) error
	GetTags(ctx context.Context, uid int64) ([]domain.Tag, error)
	GetTagsById(ctx context.Context, ids []int64) ([]domain.Tag, error)
	GetBizTags(ctx context.Context, uid int64, biz string, bizId int64) ([]domain.Tag, error)
}

type CachedTagRepository struct {
	dao   dao.TagDAO
	cache cache.TagCache
	l     logger.Logger
}

func NewCachedTagRepository(tagDAO dao.TagDAO, c cache.TagCache, l logger.Logger) TagRepository {
	return &CachedTagRepository{
		dao:   tagDAO,
		l:     l,
		cache: c,
	}
}

func (repo *CachedTagRepository) CreateTag(ctx context.Context, tag domain.Tag) (int64, error) {
	entity := repo.toEntity(tag)
	id, err := repo.dao.CreateTag(ctx, &entity)
	if err != nil {
		return 0, err
	}
	d := repo.toDomain(entity)
	// 也可以考虑用 DelTags
	err = repo.cache.Append(ctx, d.Uid, d)
	if err != nil {
		// 记录日志
	}
	return id, nil
}

func (repo *CachedTagRepository) BindTagToBiz(ctx context.Context, uid int64, biz string, bizId int64, tags []int64) error {
	return repo.dao.CreateTagBiz(ctx, slice.Map(tags, func(idx int, src int64) dao.TagBiz {
		return dao.TagBiz{
			Tid:   src,
			Uid:   uid,
			Biz:   biz,
			BizId: bizId,
		}
	}))
}

func (repo *CachedTagRepository) GetTags(ctx context.Context, uid int64) ([]domain.Tag, error) {
	res, err := repo.cache.GetTags(ctx, uid)
	if err == nil && len(res) > 0 {
		return res, nil
	}
	tags, err := repo.dao.GetTagsByUid(ctx, uid)
	if err != nil {
		return nil, err
	}

	res = slice.Map(tags, func(idx int, src dao.Tag) domain.Tag {
		return repo.toDomain(src)
	})
	err = repo.cache.Append(ctx, uid, res...)
	if err != nil {
		// 记录日志
	}
	return res, nil
}
func (repo *CachedTagRepository) GetTagsById(ctx context.Context, ids []int64) ([]domain.Tag, error) {
	tags, err := repo.dao.GetTagsById(ctx, ids)
	if err != nil {
		return nil, err
	}
	return slice.Map(tags, func(idx int, src dao.Tag) domain.Tag {
		return repo.toDomain(src)
	}), nil
}
func (repo *CachedTagRepository) GetBizTags(ctx context.Context, uid int64, biz string, bizId int64) ([]domain.Tag, error) {
	tags, err := repo.dao.GetTagsByBiz(ctx, uid, biz, bizId)
	if err != nil {
		return nil, err
	}
	return slice.Map(tags, func(idx int, src dao.Tag) domain.Tag {
		return repo.toDomain(src)
	}), nil
}

// PreloadUserTags 在 toB 的场景下，你可以提前预加载缓存
func (repo *CachedTagRepository) PreloadUserTags(ctx context.Context) error {
	offset := 0
	const batch = 100
	for {
		dbCtx, cancel := context.WithTimeout(ctx, time.Second)
		tags, err := repo.dao.GetTags(dbCtx, offset, batch)
		cancel()
		if err != nil {
			return err
		}
		// 你可以开 goroutine 来并发
		for _, tag := range tags {
			cCtx, cancel := context.WithTimeout(ctx, time.Second)
			err = repo.cache.Append(cCtx, tag.Uid, repo.toDomain(tag))
			cancel()
			if err != nil {
				return err
			}
		}
		if len(tags) < batch {
			break
		}
		offset = offset + batch
	}
	return nil
}
func (repo *CachedTagRepository) toDomain(tag dao.Tag) domain.Tag {
	return domain.Tag{
		Id:   tag.Id,
		Name: tag.Name,
		Uid:  tag.Uid,
	}
}
func (repo *CachedTagRepository) toEntity(tag domain.Tag) dao.Tag {
	return dao.Tag{
		Id:   tag.Id,
		Name: tag.Name,
		Uid:  tag.Uid,
	}
}
