package service

import (
	"archi/internal/domain"
	"archi/internal/event/tag"
	"archi/internal/repository"
	"archi/pkg/logger"
	"context"
	"github.com/ecodeclub/ekit/slice"
	"time"
)

type TagService interface {
	CreateTag(ctx context.Context, uid int64, name string) (int64, error)
	AttachTags(ctx context.Context, uid int64, biz string, bizId int64, tags []int64) error
	GetTags(ctx context.Context, uid int64) ([]domain.Tag, error)
	GetBizTags(ctx context.Context, uid int64, biz string, bizId int64) ([]domain.Tag, error)
}

type tagService struct {
	repo     repository.TagRepository
	logger   logger.Logger
	producer tag.Producer
}

func NewTagService(repo repository.TagRepository, producer tag.Producer, l logger.Logger) TagService {
	return &tagService{
		producer: producer,
		repo:     repo,
		logger:   l,
	}
}

func (svc *tagService) CreateTag(ctx context.Context, uid int64, name string) (int64, error) {
	return svc.repo.CreateTag(ctx, domain.Tag{
		Uid:  uid,
		Name: name,
	})
}
func (svc *tagService) AttachTags(ctx context.Context, uid int64, biz string, bizId int64, tags []int64) error {
	err := svc.repo.BindTagToBiz(ctx, uid, biz, bizId, tags)
	if err != nil {
		return err
	}
	// 异步发送
	go func() {
		ts, err := svc.repo.GetTagsById(ctx, tags)
		if err != nil {
			// 记录日志
		}
		// 这里要根据 tag_index 的结构来定义
		// 同样要注意顺序，即同一个用户对同一个资源打标签的顺序，
		// 是不能乱的
		pctx, cancel := context.WithTimeout(context.Background(), time.Second)
		err = svc.producer.ProduceSyncEvent(pctx, tag.BizTags{
			Uid:   uid,
			Biz:   biz,
			BizId: bizId,
			Tags: slice.Map(ts, func(idx int, src domain.Tag) string {
				return src.Name
			}),
		})
		cancel()
		if err != nil {
			// 记录日志
			svc.logger.Error("发送标签搜索事件失败", logger.Error(err))
		}
	}()
	return err
}
func (svc *tagService) GetTags(ctx context.Context, uid int64) ([]domain.Tag, error) {
	return svc.repo.GetTags(ctx, uid)
}
func (svc *tagService) GetBizTags(ctx context.Context, uid int64, biz string, bizId int64) ([]domain.Tag, error) {
	return svc.repo.GetBizTags(ctx, uid, biz, bizId)
}
