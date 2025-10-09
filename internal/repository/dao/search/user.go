package search

import (
	"context"
	"encoding/json"
	"github.com/olivere/elastic/v7"
	"strconv"
)

const UserIndexName = "user_index"

type User struct {
	Id       int64  `json:"id"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
	Phone    string `json:"phone"`
}

type UserDAO interface {
	InputUser(ctx context.Context, user User) error
	Search(ctx context.Context, keywords []string) ([]User, error)
}

type ESUserDAO struct {
	client *elastic.Client
}

func NewESUserDAO(client *elastic.Client) UserDAO {
	return &ESUserDAO{
		client: client,
	}
}

func (h *ESUserDAO) Search(ctx context.Context, keywords []string) ([]User, error) {
	// 创建一个布尔查询
	query := elastic.NewBoolQuery()

	// 遍历每一个关键词
	for _, keyword := range keywords {
		// 为每个关键词创建一组 OR 条件
		// 只要 nickname, email, 或 phone 中任何一个字段匹配到该关键词即可
		shouldQueries := []elastic.Query{
			elastic.NewMatchQuery("nickname", keyword),
			elastic.NewTermQuery("email", keyword),
			elastic.NewTermQuery("phone", keyword),
		}
		// 将这组 OR 条件作为一个整体，添加到外层的 Should 查询中
		// 这意味着，只要满足任何一个关键词的匹配条件，文档就会被选中
		query.Should(elastic.NewBoolQuery().Should(shouldQueries...))
	}
	// 至少要匹配上一个 Should 子句
	query.MinimumNumberShouldMatch(1)
	resp, err := h.client.Search(UserIndexName).Query(query).Do(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]User, 0, len(resp.Hits.Hits))
	for _, hit := range resp.Hits.Hits {
		var ele User
		err = json.Unmarshal(hit.Source, &ele)
		if err != nil {
			return nil, err
		}
		res = append(res, ele)
	}
	return res, nil
}

func (h *ESUserDAO) InputUser(ctx context.Context, user User) error {
	_, err := h.client.Index().Index(UserIndexName).Id(strconv.FormatInt(user.Id, 10)).BodyJson(user).Do(ctx)
	return err
}
