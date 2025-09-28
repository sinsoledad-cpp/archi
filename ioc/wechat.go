package ioc

import (
	"archi/internal/service/oauth2/wechat"
	"archi/pkg/logger"
	"os"
)

func InitWechatService(l logger.Logger) wechat.Service {
	appID, ok := os.LookupEnv("WECHAT_APP_ID")
	if !ok {
		panic("找不到环境变量 WECHAT_APP_ID")
	}
	appSecret, ok := os.LookupEnv("WECHAT_APP_SECRET")
	if !ok {
		panic("找不到环境变量 WECHAT_APP_SECRET")
	}
	return wechat.NewDefaultService(appID, appSecret, l)
}
