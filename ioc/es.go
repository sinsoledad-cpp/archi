package ioc

import (
	"archi/internal/repository/dao/search"
	"fmt"
	"github.com/olivere/elastic/v7"
	"github.com/spf13/viper"
	"time"
)

func InitESClient() *elastic.Client {
	type Config struct {
		Url   string `yaml:"url"`
		Sniff bool   `yaml:"sniff"`
	}
	var cfg Config
	err := viper.UnmarshalKey("es", &cfg)
	if err != nil {
		panic(fmt.Errorf("读取 ES 配置失败 %w", err))
	}
	const timeout = 100 * time.Second
	opts := []elastic.ClientOptionFunc{
		elastic.SetURL(cfg.Url),
		elastic.SetHealthcheckTimeoutStartup(timeout),
	}
	client, err := elastic.NewClient(opts...)
	if err != nil {
		panic(err)
	}
	err = search.InitES(client)
	if err != nil {
		panic(err)
	}
	return client
}
