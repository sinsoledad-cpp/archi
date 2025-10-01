package setting

import (
	"archi/pkg/validate"
	"encoding/json"
	"fmt"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func InitViper() {
	if err := validate.InitTrans("zh"); err != nil {
		panic(err)
	}
	file := pflag.String("config", "config/dev.yaml", "配置文件路径")
	// 这一步之后，file 里面才有值
	pflag.Parse()
	//viper.Set("db.dsn", "localhost:3306")
	// 所有的默认值放好s
	viper.SetConfigType("yaml")
	viper.SetConfigFile(*file)
	// 读取配置
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	fmt.Println("配置文件读取成功", viper.ConfigFileUsed())

	// --- 开始打印配置 ---

	// 1. 获取所有配置
	allSettings := viper.AllSettings()

	// 2. 以格式化的 JSON 格式打印
	jsonData, err := json.MarshalIndent(allSettings, "", "  ") // 使用两个空格进行缩进
	if err != nil {
		panic(err)
	}
	fmt.Println("--- All Settings as JSON ---")
	fmt.Println(string(jsonData))
}
