package errs

/*
对于查询成功 (GET)，返回 200。

对于创建成功 (POST)，返回 201。

对于接受异步处理 (POST/PUT)，返回 202。

对于删除成功 (DELETE)，可以返回 204（此时响应体为空）或者返回 200 并附带一些信息。
*/

// User 部分，模块代码使用 01
const (
	// UserInvalidInput 这是一个非常含糊的错误码，代表用户相关的API参数不对
	UserInvalidInput = 401001
	// UserInternalServerError 这是一个非常含糊的错误码。代表用户模块系统内部错误
	UserInternalServerError = 501001

	// UserInvalidOrPassword 用户输入的账号或者密码不对
	UserInvalidOrPassword = 401002
	// UserDuplicateEmail 邮箱冲突
	UserDuplicateEmail = 401003
	// UserInputPhone 请输入手机号
	UserInputPhone = 401004
	// UserCodeSendTooMany 短信发送太频繁
	UserCodeSendTooMany = 401005
	// UserCodeInvalid 验证码错误
	UserCodeInvalid = 401006
	// UserCodeExpired 验证码过期
	UserCodeExpired = 401007
	// UserCodeVerifyTooMany 验证码错误次数过多
	UserCodeVerifyTooMany = 401008
	// UserSmsStateInvalid 登录短信状态错误
	UserSmsStateInvalid = 401009
	// UserSmsCodeInvalid 登录短信授权码错误
	UserSmsCodeInvalid = 401010
)
