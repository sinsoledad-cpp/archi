package service

import (
	"archi/internal/domain"
	"archi/internal/repository"
	"context"
	"errors"
	"github.com/ecodeclub/ekit/queue"
	"github.com/ecodeclub/ekit/slice"
	"math"
	"time"
)

type RankingService interface {
	// TopN 前 100 的
	TopN(ctx context.Context) error
	GetTopN(ctx context.Context) ([]domain.Article, error)
}

type BatchRankingService struct {
	// 用来取点赞数
	intrSvc InteractiveService

	// 用来查找文章
	artSvc ArticleService

	batchSize int
	scoreFunc func(likeCnt int64, utime time.Time) float64
	n         int

	rankingRepository repository.RankingRepository
}

func NewBatchRankingService(intrSvc InteractiveService, artSvc ArticleService, repo repository.RankingRepository) RankingService {
	return &BatchRankingService{
		intrSvc:           intrSvc,
		artSvc:            artSvc,
		batchSize:         100,
		n:                 100,
		rankingRepository: repo,
		scoreFunc: func(likeCnt int64, utime time.Time) float64 {
			// 时间
			duration := time.Since(utime).Seconds()
			return float64(likeCnt-1) / math.Pow(duration+2, 1.5)
		},
	}
}

func (b *BatchRankingService) TopN(ctx context.Context) error {
	arts, err := b.topN(ctx)
	if err != nil {
		return err
	}
	// 最终是要放到缓存里面的
	// 存到缓存里面
	return b.rankingRepository.ReplaceTopN(ctx, arts)
}

func (b *BatchRankingService) topN(ctx context.Context) ([]domain.Article, error) {
	offset := 0
	start := time.Now()
	ddl := start.Add(-7 * 24 * time.Hour)

	type Score struct {
		score float64
		art   domain.Article
	}

	topN := queue.NewPriorityQueue[Score](b.n,
		func(src Score, dst Score) int {
			if src.score > dst.score {
				return 1
			} else if src.score == dst.score {
				return 0
			} else {
				return -1
			}
		})

	for {
		arts, err := b.artSvc.ListPub(ctx, start, offset, b.batchSize)
		if err != nil {
			return nil, err
		}

		ids := slice.Map(arts, func(idx int, art domain.Article) int64 {
			return art.ID
		})

		intrMap, err := b.intrSvc.GetByIds(ctx, "article", ids)
		if err != nil {
			return nil, err
		}

		for _, art := range arts {
			intr := intrMap[art.ID]
			score := b.scoreFunc(intr.LikeCnt, art.Utime)
			ele := Score{
				score: score,
				art:   art,
			}

			// 尝试直接入队
			err = topN.Enqueue(ele)

			// 如果队列已满，说明需要进行比较和替换
			if errors.Is(err, queue.ErrOutOfCapacity) {
				// 先 "偷看" 一下队头分数最低的元素
				minEle, _ := topN.Peek()
				// 如果新元素的得分更高，才进行替换
				if minEle.score < ele.score {
					// 丢弃老的最小元素
					_, _ = topN.Dequeue()
					// 加入新元素
					_ = topN.Enqueue(ele)
				}
				// 如果新元素得分不够高，则什么也不做，继续下一次循环
			}
		}

		offset = offset + len(arts)
		if len(arts) < b.batchSize || arts[len(arts)-1].Utime.Before(ddl) {
			break
		}
	}

	res := make([]domain.Article, topN.Len())
	for i := topN.Len() - 1; i >= 0; i-- {
		ele, _ := topN.Dequeue()
		res[i] = ele.art
	}
	return res, nil
}
func (b *BatchRankingService) GetTopN(ctx context.Context) ([]domain.Article, error) {
	return b.rankingRepository.GetTopN(ctx)
}
