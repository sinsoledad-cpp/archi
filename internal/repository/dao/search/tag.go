package search

import (
	"context"
	"encoding/json"
	"github.com/olivere/elastic/v7"
)

const TagIndexName = "tags_index"

type BizTags struct {
	Uid   int64    `json:"uid"`
	Biz   string   `json:"biz"`
	BizId int64    `json:"biz_id"`
	Tags  []string `json:"tags"`
}

type TagDAO interface {
	Search(ctx context.Context, uid int64, biz string, keywords []string) ([]int64, error)
}
type ESTagDAO struct {
	client *elastic.Client
}

func NewESTagDAO(client *elastic.Client) TagDAO {
	return &ESTagDAO{client: client}
}

func (t *ESTagDAO) Search(ctx context.Context, uid int64, biz string, keywords []string) ([]int64, error) {
	query := elastic.NewBoolQuery().Must(
		elastic.NewTermsQuery("uid", uid),
		elastic.NewTermsQueryFromStrings("tags", keywords...),
		elastic.NewTermQuery("biz", biz))
	resp, err := t.client.Search(TagIndexName).Query(query).Do(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]int64, 0, len(resp.Hits.Hits))
	for _, hit := range resp.Hits.Hits {
		var ele BizTags
		err = json.Unmarshal(hit.Source, &ele)
		if err != nil {
			return nil, err
		}
		res = append(res, ele.BizId)
	}
	return res, nil
}
