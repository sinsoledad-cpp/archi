package search

import (
	"context"
	"github.com/olivere/elastic/v7"
)

type AnyDAO interface {
	Input(ctx context.Context, index, docID, data string) error
}
type ESAnyDAO struct {
	client *elastic.Client
}

func NewESAnyDAO(client *elastic.Client) AnyDAO {
	return &ESAnyDAO{client: client}
}

func (a *ESAnyDAO) Input(ctx context.Context, index, docId, data string) error {
	_, err := a.client.Index().Index(index).Id(docId).BodyString(data).Do(ctx)
	return err
}
