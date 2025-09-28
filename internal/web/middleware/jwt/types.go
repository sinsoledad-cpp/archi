package jwt

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var ErrSessionNotFound = errors.New("会话不存在或已过期")

//go:generate mockgen -source=./types.go -package=jwtmocks -destination=./mocks/handler.mock.go Handler
type Handler interface {
	ClearToken(ctx *gin.Context) error
	SetLoginToken(ctx *gin.Context, uid int64) error
	SetJWTToken(ctx *gin.Context, uid int64, ssid string) error
	CheckSession(ctx *gin.Context, ssid string) error
	ExtractTokenString(ctx *gin.Context) string
}

var AccessTokenKey = []byte("k6CswdUm77WKcbM68UQUuxVsHSpTCwgK")
var RefreshTokenKey = []byte("k6CswdUm77WKcbM68UQUuxVsHSpTCwgA") //Refresh Claims

type UserClaims struct {
	jwt.RegisteredClaims
	Uid       int64
	Ssid      string
	UserAgent string
}

type RefreshClaims struct {
	jwt.RegisteredClaims
	Uid  int64  //User ID
	Ssid string //Session ID
}
