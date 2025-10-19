package ioc

import (
	"archi/internal/event"
	"archi/internal/event/article"
	"archi/internal/event/feed"
	"archi/internal/event/search"

	"github.com/IBM/sarama"
	"github.com/spf13/viper"
)

func InitSaramaClient() sarama.Client {
	type Config struct {
		Addr []string `yaml:"addr"`
	}
	var cfg Config
	err := viper.UnmarshalKey("kafka", &cfg)
	if err != nil {
		panic(err)
	}

	scfg := sarama.NewConfig()

	scfg.Producer.RequiredAcks = sarama.WaitForAll // 等待所有副本确认
	scfg.Producer.Retry.Max = 3                    // 重试次数
	scfg.Producer.Return.Successes = true          // 返回成功的消息

	scfg.Consumer.Offsets.Initial = sarama.OffsetNewest

	client, err := sarama.NewClient(cfg.Addr, scfg)
	if err != nil {
		panic(err)
	}
	return client
}
func InitSyncProducer(c sarama.Client) sarama.SyncProducer {
	p, err := sarama.NewSyncProducerFromClient(c)
	if err != nil {
		panic(err)
	}
	return p
}

// InitConsumers “新建消费者组”，准确描述了它“创建了一个组”
func InitConsumers(
	artReadC *article.ReadEventConsumer,
	userSearchC *search.UserConsumer,
	artSearchC *search.ArticleConsumer,
	syncC *search.SyncDataEventConsumer,
	followC *feed.FollowEventConsumer,
) []event.Consumer {
	return []event.Consumer{
		artReadC,
		userSearchC,
		artSearchC,
		syncC,
		followC,
	}
}
