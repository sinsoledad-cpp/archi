package middleware

import (
	jwtware "archi/internal/web/middleware/jwt"
	"github.com/ecodeclub/ekit/set"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"
)

type JWTAuth struct {
	publicPaths set.Set[string]
	hdl         jwtware.Handler
}

func NewJWTAuth(hdl jwtware.Handler) *JWTAuth {
	s := set.NewMapSet[string](6)
	s.Add("/users/signup")
	s.Add("/users/login_sms/code/send")
	s.Add("/users/login_sms")
	s.Add("/users/refresh_token")
	s.Add("/users/login")
	s.Add("/oauth2/wechat/authurl")
	s.Add("/oauth2/wechat/callback")
	s.Add("/test/random")
	return &JWTAuth{
		publicPaths: s,
		hdl:         hdl,
	}
}

func (j *JWTAuth) SetPublicPaths(paths ...string) {
	for _, path := range paths {
		j.publicPaths.Add(path)
	}
}

func (j *JWTAuth) Middleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 不需要校验
		if j.publicPaths.Exist(ctx.Request.URL.Path) {
			return
		}
		// 如果是空字符串，你可以预期后面 Parse 就会报错
		tokenStr := j.hdl.ExtractTokenString(ctx)
		uc := jwtware.UserClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, &uc, func(token *jwt.Token) (interface{}, error) {
			return jwtware.AccessTokenKey, nil
		})
		if err != nil || !token.Valid {
			// 不正确的 token
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		expireTime, err := uc.GetExpirationTime()
		if err != nil {
			// 拿不到过期时间
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if expireTime.Before(time.Now()) {
			// 已经过期
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 注释掉，防止有些同学在浏览器和 postman 混用的时候
		// 没有办法继续访问
		//if ctx.GetHeader("User-Agent") != uc.UserAgent {
		//	// 换了一个 User-Agent，可能是攻击者
		//	ctx.AbortWithStatus(http.StatusUnauthorized)
		//	return
		//}

		err = j.hdl.CheckSession(ctx, uc.Ssid)
		if err != nil {
			// 系统错误或者用户已经主动退出登录了
			// 这里也可以考虑说，如果在 Redis 已经崩溃的时候，
			// 就不要去校验是不是已经主动退出登录了。
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 说明 token 是合法的
		// 我们把这个 token 里面的数据放到 ctx 里面，后面用的时候就不用再次 Parse 了
		ctx.Set("user", uc)
	}
}
