package web

import (
	"archi/internal/domain"
	"archi/internal/service"
	"archi/internal/web/errs"
	jwtware "archi/internal/web/middleware/jwt"
	"archi/pkg/ginx"
	"archi/pkg/logger"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"time"

	regexp "github.com/dlclark/regexp2"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var _ Handler = (*UserHandler)(nil)

const (
	emailRegexPattern    = "(?i)^[A-Z0-9_!#$%&'*+/=?`{|}~^.-]+@[A-Z0-9.-]+$"
	passwordRegexPattern = `^(?=.*[a-z])(?=.*[A-Z])(?=.*[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~]).{8,}$`
	bizLogin             = "login"
)

type UserHandler struct {
	log              logger.Logger
	userSvc          service.UserService
	codeSvc          service.CodeService
	jwtHdl           jwtware.Handler
	emailRegexExp    *regexp.Regexp
	passwordRegexExp *regexp.Regexp
}

func NewUserHandler(log logger.Logger, userSvc service.UserService, codeSvc service.CodeService, jwtHdl jwtware.Handler) *UserHandler {
	return &UserHandler{
		log:              log,
		userSvc:          userSvc,
		codeSvc:          codeSvc,
		jwtHdl:           jwtHdl,
		emailRegexExp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRegexExp: regexp.MustCompile(passwordRegexPattern, regexp.None),
	}
}

func (u *UserHandler) RegisterRoutes(e *gin.Engine) {
	g := e.Group("/users")

	g.POST("/signup", ginx.WrapBody(u.SignUp))
	g.POST("/login", ginx.WrapBody(u.LoginJWT))
	g.POST("/logout", ginx.Wrap(u.LogoutJWT))
	g.POST("/refresh_token", ginx.Wrap(u.RefreshToken))

	g.POST("/avatar/upload", ginx.WrapClaims(u.UploadAvatar))
	g.POST("/edit", ginx.WrapBodyAndClaims(u.Edit))
	g.GET("/profile", ginx.WrapClaims(u.Profile))

	g.POST("/login_sms/code/send", ginx.WrapBody(u.SendSMSLoginCode))
	g.POST("/login_sms", ginx.WrapBody(u.LoginSMS))
}

type SignUpReq struct {
	Email           string `json:"email" binding:"required,email"`
	Password        string `json:"password" binding:"required,min=8,max=32"`
	ConfirmPassword string `json:"confirmPassword" binding:"required,eqfield=Password"`
}

func (u *UserHandler) SignUp(ctx *gin.Context, req SignUpReq) (ginx.Result, error) {
	// 校验客户端输入
	isEmail, err := u.emailRegexExp.MatchString(req.Email)
	if err != nil {
		return ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	if !isEmail {

		return ginx.Result{
			Code: errs.UserInvalidInput,
			Msg:  "邮箱格式错误",
		}, nil
	}
	if req.Password != req.ConfirmPassword {
		return ginx.Result{
			Code: errs.UserInvalidInput,
			Msg:  "两次输入密码不同",
		}, nil
	}
	isPassword, err := u.passwordRegexExp.MatchString(req.Password)
	if err != nil {
		return ginx.Result{
			Code: errs.UserInvalidInput,
			Msg:  "系统错误",
		}, err
	}
	if !isPassword {
		return ginx.Result{
			Code: errs.UserInvalidInput,
			Msg:  "密码必须包含数字、特殊字符、大小字母，并且长度不能小于 8 位",
		}, nil
	}

	// 业务逻辑
	err = u.userSvc.Signup(ctx.Request.Context(), domain.User{Email: req.Email, Password: req.ConfirmPassword})
	if errors.Is(err, service.ErrDuplicateEmail) {
		u.log.Warn("用户邮箱冲突", logger.Error(err))
		return ginx.Result{
			Code: errs.UserDuplicateEmail,
			Msg:  "邮箱冲突",
		}, err
	}
	if err != nil {
		return ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "系统错误",
		}, err
	}

	return ginx.Result{
		Code: http.StatusCreated,
		Msg:  "注册成功",
	}, nil
}

type LoginJWTReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=32"`
}

func (u *UserHandler) LoginJWT(ctx *gin.Context, req LoginJWTReq) (ginx.Result, error) {
	user, err := u.userSvc.Login(ctx, req.Email, req.Password)
	switch {
	case err == nil:
		err = u.jwtHdl.SetLoginToken(ctx, user.ID)
		if err != nil {
			return ginx.Result{
				Code: errs.UserInternalServerError,
				Msg:  "系统错误",
			}, err
		}
		return ginx.Result{
			Code: http.StatusOK,
			Msg:  "登录成功",
		}, nil
	case errors.Is(err, service.ErrInvalidUserOrPassword):
		return ginx.Result{
			Code: errs.UserInvalidOrPassword,
			Msg:  "用户名或者密码错误",
		}, err
	default:
		return ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "系统错误",
		}, err
	}
}

func (u *UserHandler) LogoutJWT(ctx *gin.Context) (ginx.Result, error) {
	err := u.jwtHdl.ClearToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{Code: http.StatusInternalServerError, Msg: "系统错误"})
		return ginx.Result{
			Code: http.StatusInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		Code: http.StatusOK,
		Msg:  "退出登录成功",
	}, nil
}

func (u *UserHandler) RefreshToken(ctx *gin.Context) (ginx.Result, error) {
	// 假定长 token 也放在这里
	tokenStr := ctx.GetHeader("X-Refresh-Token")

	var rc jwtware.RefreshClaims
	token, err := jwt.ParseWithClaims(tokenStr, &rc, func(token *jwt.Token) (interface{}, error) {
		return jwtware.RefreshTokenKey, nil
	})
	// 这边要保持和登录校验一直的逻辑，即返回 401 响应
	if err != nil || token == nil || !token.Valid {
		return ginx.Result{
			Code: http.StatusUnauthorized,
			Msg:  "登录已过期，请重新登录",
		}, err
	}

	// 校验 ssid
	err = u.jwtHdl.CheckSession(ctx, rc.Ssid)
	if err != nil {
		// 如果是会话不存在的业务错误，返回 401
		if errors.Is(err, jwtware.ErrSessionNotFound) {
			return ginx.Result{
				Code: http.StatusUnauthorized,
				Msg:  "会话已过期，请重新登录",
			}, err
		}
		// 系统错误或者用户已经主动退出登录了
		// 这里也可以考虑说，如果在 Redis 已经崩溃的时候，
		// 就不要去校验是不是已经主动退出登录了。
		//ctx.AbortWithStatus(http.StatusUnauthorized)
		return ginx.Result{
			Code: http.StatusInternalServerError,
			Msg:  "系统错误",
		}, err
	}

	err = u.jwtHdl.SetJWTToken(ctx, rc.Uid, rc.Ssid)
	if err != nil {
		return ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "系统内部错误",
		}, err
	}
	return ginx.Result{
		Code: http.StatusOK,
		Msg:  "刷新成功",
	}, nil
}

func (u *UserHandler) UploadAvatar(ctx *gin.Context, uc jwtware.UserClaims) (ginx.Result, error) {
	file, err := ctx.FormFile("avatar") // 前端表单字段名应为 'avatar'
	if err != nil {
		return ginx.Result{
			Code: errs.UserInvalidInput,
			Msg:  "请上传头像文件",
		}, err
	}

	// 1. 生成新文件的唯一路径
	ext := filepath.Ext(file.Filename)
	newPath := filepath.Join("uploads", "avatars", uuid.New().String()+ext)

	// 2. 创建目录
	if err := os.MkdirAll(filepath.Dir(newPath), os.ModePerm); err != nil {
		u.log.Error("创建头像目录失败", logger.Error(err))
		return ginx.Result{Code: http.StatusInternalServerError, Msg: "系统错误"}, err
	}

	// 3. 保存文件
	if err := ctx.SaveUploadedFile(file, newPath); err != nil {
		u.log.Error("保存头像文件失败", logger.Error(err))
		return ginx.Result{
			Code: http.StatusInternalServerError,
			Msg:  "系统错误",
		}, err
	}

	// 4. 调用 service 层处理业务逻辑
	err = u.userSvc.UpdateAvatarPath(ctx.Request.Context(), uc.Uid, newPath)
	if err != nil {
		// 如果业务逻辑处理失败（例如数据库更新失败），则删除刚刚保存的文件，进行“回滚”
		// *** 使用绝对路径进行删除，增强代码的健壮性 ***
		absNewPath, absErr := filepath.Abs(newPath)
		if absErr != nil {
			// 如果获取绝对路径失败，记录一个警告，然后尝试用原路径删除
			u.log.Warn("获取新头像绝对路径失败，将尝试使用相对路径删除",
				logger.Error(absErr),
				logger.String("new_avatar_path", newPath),
			)
			absNewPath = newPath // 回退到使用相对路径
		}

		if removeErr := os.Remove(absNewPath); removeErr != nil {
			u.log.Warn("数据库更新失败进行回滚操作,但是删除新头像失败",
				logger.Error(removeErr),
				logger.String("new_avatar_path", absNewPath),
			)
		}

		u.log.Error("更新用户头像业务逻辑失败", logger.Error(err))
		return ginx.Result{
			Code: http.StatusInternalServerError,
			Msg:  "系统错误",
		}, err
	}

	return ginx.Result{
		Code: http.StatusOK,
		Msg:  "头像上传成功",
		Data: gin.H{
			"avatar_url": newPath, // 返回新的头像路径
		},
	}, nil
}

func (u *UserHandler) Profile(ctx *gin.Context, uc jwtware.UserClaims) (ginx.Result, error) {
	user, err := u.userSvc.FindById(ctx, uc.Uid)
	if err != nil {
		return ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	type User struct {
		Nickname string `json:"nickname"`
		Email    string `json:"email"`
		AboutMe  string `json:"aboutMe"`
		Birthday string `json:"birthday"`
		Avatar   string `json:"avatar"`
	}
	return ginx.Result{
		Code: http.StatusOK,
		Msg:  "获取用户信息成功",
		Data: User{
			Nickname: user.Nickname,
			Email:    user.Email,
			AboutMe:  user.AboutMe,
			Birthday: user.Birthday.Format(time.DateOnly),
			Avatar:   user.Avatar,
		},
	}, nil
}

type UserEditReq struct {
	// 改邮箱，密码，或者能不能改手机号
	Nickname string `json:"nickname"`
	// YYYY-MM-DD
	Birthday string `json:"birthday"`
	AboutMe  string `json:"aboutMe"`
}

func (u *UserHandler) Edit(ctx *gin.Context, req UserEditReq, uc jwtware.UserClaims) (ginx.Result, error) {
	// 嵌入一段刷新过期时间的代码
	//sess := sessions.Default(ctx)
	//sess.Get("uid")
	// 用户输入不对
	birthday, err := time.Parse(time.DateOnly, req.Birthday)
	if err != nil {
		return ginx.Result{
			Code: errs.UserInvalidInput,
			Msg:  "生日格式不对",
		}, err
	}
	err = u.userSvc.UpdateNonSensitiveInfo(ctx, domain.User{
		ID:       uc.Uid,
		Nickname: req.Nickname,
		Birthday: birthday,
		AboutMe:  req.AboutMe,
	})
	if err != nil {
		return ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		Code: http.StatusOK,
		Msg:  "上传成功",
	}, nil
}

type SendSMSCodeReq struct {
	Phone string `json:"phone" binding:"required,len=11,numeric"`
}

func (u *UserHandler) SendSMSLoginCode(ctx *gin.Context, req SendSMSCodeReq) (ginx.Result, error) {
	// 你这边可以校验 Req
	if req.Phone == "" {
		return ginx.Result{
			Code: errs.UserInvalidInput,
			Msg:  "请输入手机号码",
		}, nil
	}
	err := u.codeSvc.Send(ctx.Request.Context(), bizLogin, req.Phone)
	switch {
	case err == nil:
		return ginx.Result{
			Code: http.StatusOK,
			Msg:  "发送成功",
		}, nil
	case errors.Is(err, service.ErrCodeSendTooMany):
		// 事实上，防不住有人不知道怎么触发了
		// 少数这种错误，是可以接受的
		// 但是频繁出现，就代表有人在搞你的系统
		return ginx.Result{
			Code: errs.UserCodeSendTooMany,
			Msg:  "短信发送太频繁，请稍后再试",
		}, nil
	default:
		return ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "系统错误",
		}, err
	}
}

type LoginSMSReq struct {
	Phone string `json:"phone" binding:"required,len=11,numeric"`
	Code  string `json:"code" binding:"required,len=6,numeric"`
}

func (u *UserHandler) LoginSMS(ctx *gin.Context, req LoginSMSReq) (ginx.Result, error) {
	ok, err := u.codeSvc.Verify(ctx, bizLogin, req.Phone, req.Code)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCodeVerifyTooMany):
			return ginx.Result{
				Code: errs.UserCodeVerifyTooMany,
				Msg:  "验证码验证次数太多，请稍后再试",
			}, nil
		case errors.Is(err, service.ErrCodeExpired):
			return ginx.Result{
				Code: errs.UserCodeExpired,
				Msg:  "验证码已过期",
			}, nil
		default:
			return ginx.Result{
				Code: errs.UserInternalServerError,
				Msg:  "系统异常",
			}, err
		}
	}
	if !ok {
		return ginx.Result{
			Code: errs.UserCodeInvalid,
			Msg:  "验证码不对，请重新输入",
		}, nil
	}
	user, err := u.userSvc.FindOrCreate(ctx, req.Phone)
	if err != nil {
		return ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	err = u.jwtHdl.SetLoginToken(ctx, user.ID)
	if err != nil {
		return ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		Code: http.StatusOK,
		Msg:  "登录成功",
	}, nil
}
