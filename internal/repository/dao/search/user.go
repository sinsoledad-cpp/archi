package search

import (
	"context"
	"encoding/json"
	"github.com/olivere/elastic/v7"
	"strconv"
	"strings"
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
	// 假定上面传入的 keywords 是经过了处理的
	queryString := strings.Join(keywords, " ")
	//query := elastic.NewBoolQuery().Must(elastic.NewMatchQuery("nickname", queryString))
	query := elastic.NewBoolQuery().Should(
		// 对 nickname 字段使用 match 查询，支持全文搜索
		elastic.NewMatchQuery("nickname", queryString),
		// 对 email 和 phone 字段使用 term 查询，进行精确匹配
		elastic.NewTermQuery("email", queryString),
		elastic.NewTermQuery("phone", queryString),
	)
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
