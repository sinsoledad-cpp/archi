package setting

import "archi/pkg/validate"

func InitValidate() {
	if err := validate.InitTrans("zh"); err != nil {
		panic(err)
	}
}
