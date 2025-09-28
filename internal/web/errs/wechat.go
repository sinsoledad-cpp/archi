package errs

// Wechat 部分，模块代码使用 01
const (
	// WechatInvalidInput 这是一个非常含糊的错误码，代表用户相关的API参数不对
	WechatInvalidInput = 402001
	// WechatInternalServerError 这是一个非常含糊的错误码。代表系统内部错误
	WechatInternalServerError = 502001
	// WechatCodeGetDefeated 获取微信授权码失败
	WechatCodeGetDefeated = 402002
	// WechatInvalidRequest 非法请求
	WechatInvalidRequest = 40200
	// WechatInvalidCode 微信授权码错误
	WechatInvalidCode = 402003
)
