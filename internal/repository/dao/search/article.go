package search

import (
	"context"
	"encoding/json"
	"github.com/ecodeclub/ekit/slice"
	"github.com/olivere/elastic/v7"
	"strconv"
	"strings"
)

const ArticleIndexName = "article_index"

type Article struct {
	Id      int64    `json:"id"`
	Title   string   `json:"title"`
	Status  int32    `json:"status"`
	Content string   `json:"content"`
	Tags    []string `json:"tags"`
}

type ArticleDAO interface {
	InputArticle(ctx context.Context, article Article) error
	Search(ctx context.Context, tagArtIds []int64, keywords []string) ([]Article, error)
}

type ESArticleDAO struct {
	client *elastic.Client
}

func NewESArticleDAO(client *elastic.Client) ArticleDAO {
	return &ESArticleDAO{client: client}
}
func (h *ESArticleDAO) InputArticle(ctx context.Context, art Article) error {
	_, err := h.client.Index().Index(ArticleIndexName).Id(strconv.FormatInt(art.Id, 10)).BodyJson(art).Do(ctx)
	return err
}

func (h *ESArticleDAO) Search(ctx context.Context, tagArtIds []int64, keywords []string) ([]Article, error) {
	// 把关键词数组用空格连接成一个字符串
	queryString := strings.Join(keywords, " ")

	// 把 int64 类型的 ID 数组转换成 any 类型的数组，以符合查询 API 的要求
	ids := slice.Map(tagArtIds, func(idx int, src int64) any {
		return src
	})

	// 构建一个复杂的布尔查询 (Bool Query)
	query := elastic.NewBoolQuery().Must(
		// Must: 相当于 AND, 里面的所有条件都必须满足
		elastic.NewBoolQuery().Should(
			// Should: 相当于 OR, 里面的条件满足一个即可
			elastic.NewTermsQuery("id", ids...).Boost(2),   // 条件1: id 在给定的 id 列表 (tagArtIds) 中。Boost(2) 表示这个条件权重*2，匹配到它会优先展示
			elastic.NewMatchQuery("title", queryString),    // 条件2: title 字段包含关键词
			elastic.NewMatchQuery("content", queryString)), // 条件3: content 字段包含关键词
		elastic.NewTermQuery("status", 2)) // 同时，status 字段必须精确等于 2 (已发布)

	// 执行搜索
	resp, err := h.client.Search(ArticleIndexName).Query(query).Do(ctx)
	if err != nil {
		return nil, err
	}

	// 处理搜索结果
	res := make([]Article, 0, len(resp.Hits.Hits)) // 创建一个切片来存放结果
	for _, hit := range resp.Hits.Hits {
		var ele Article
		// Elasticsearch 返回的是 JSON 字符串 (hit.Source)，需要反序列化回 Go 的 Article 结构体
		err = json.Unmarshal(hit.Source, &ele)
		res = append(res, ele)
	}
	return res, nil
}
