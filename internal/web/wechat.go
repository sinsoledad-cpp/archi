package web

import (
	"archi/internal/service"
	"archi/internal/service/oauth2/wechat"
	"archi/internal/web/errs"
	"archi/pkg/ginx"
	"archi/pkg/logger"
	"fmt"

	jwtware "archi/internal/web/middleware/jwt"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	uuid "github.com/lithammer/shortuuid/v4"
	"net/http"
)

type StateClaims struct {
	jwt.RegisteredClaims
	State string
}

type OAuth2WechatHandler struct {
	wechatSvc       wechat.Service
	userSvc         service.UserService
	jwtHdl          jwtware.Handler
	key             []byte
	stateCookieName string
	l               logger.Logger
}

func NewOAuth2WechatHandler(svc wechat.Service, hdl jwtware.Handler, userSvc service.UserService) *OAuth2WechatHandler {
	return &OAuth2WechatHandler{
		wechatSvc:       svc,
		userSvc:         userSvc,
		key:             []byte("k6CswdUm77WKcbM68UQUuxVsHSpTCwgB"),
		stateCookieName: "jwt-state",
		jwtHdl:          hdl,
	}
}

func (o *OAuth2WechatHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/oauth2/wechat")
	g.GET("/authurl", ginx.Wrap(o.Auth2URL))
	g.Any("/callback", ginx.Wrap(o.Callback))
}

func (o *OAuth2WechatHandler) Auth2URL(ctx *gin.Context) (ginx.Result, error) {
	state := uuid.New()
	val, err := o.wechatSvc.AuthURL(ctx, state)
	if err != nil {
		o.l.Error("获取微信授权码失败", logger.Error(err))
		return ginx.Result{
			Code: errs.WechatCodeGetDefeated,
			Msg:  "获取微信授权码失败",
		}, err
	}
	err = o.setStateCookie(ctx, state)
	if err != nil {
		o.l.Error("设置 state cookie 失败", logger.Error(err))
		return ginx.Result{
			Code: errs.WechatInternalServerError,
			Msg:  "服务器异常",
		}, err
	}
	return ginx.Result{
		Code: http.StatusOK,
		Msg:  "OK",
		Data: val,
	}, nil

}

func (o *OAuth2WechatHandler) Callback(ctx *gin.Context) (ginx.Result, error) {
	err := o.verifyState(ctx)
	if err != nil {
		return ginx.Result{
			Code: errs.WechatInvalidRequest,
			Msg:  "非法请求",
		}, err
	}
	// 你校验不校验都可以
	code := ctx.Query("code")
	// state := ctx.Query("state")
	wechatInfo, err := o.wechatSvc.VerifyCode(ctx, code)
	if err != nil {
		return ginx.Result{
			Code: errs.WechatInvalidCode,
			Msg:  "授权码有误",
		}, err
	}
	u, err := o.userSvc.FindOrCreateByWechat(ctx, wechatInfo)
	if err != nil {
		return ginx.Result{
			Code: errs.WechatInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	err = o.jwtHdl.SetLoginToken(ctx, u.ID)
	if err != nil {
		return ginx.Result{
			Code: errs.WechatInternalServerError,
			Msg:  "系统错误",
		}, err
	}

	return ginx.Result{
		Code: http.StatusOK,
		Msg:  "微信授权成功",
	}, nil
}

func (o *OAuth2WechatHandler) setStateCookie(ctx *gin.Context,
	state string) error {
	claims := StateClaims{
		State: state,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenStr, err := token.SignedString(o.key)
	if err != nil {

		return err
	}
	ctx.SetCookie(o.stateCookieName, tokenStr,
		600, "/oauth2/wechat/callback",
		"", false, true)
	return nil
}

func (o *OAuth2WechatHandler) verifyState(ctx *gin.Context) error {
	state := ctx.Query("state")
	ck, err := ctx.Cookie(o.stateCookieName)
	if err != nil {
		return fmt.Errorf("无法获得 cookie %w", err)
	}
	var sc StateClaims
	_, err = jwt.ParseWithClaims(ck, &sc, func(token *jwt.Token) (interface{}, error) {
		return o.key, nil
	})
	if err != nil {
		return fmt.Errorf("解析 token 失败 %w", err)
	}
	if state != sc.State {
		// state 不匹配，有人搞你
		return fmt.Errorf("state 不匹配")
	}
	return nil
}
