package web

import (
	"archi/internal/service/feed"
	"archi/internal/web/errs"
	"archi/internal/web/middleware/jwt"
	"archi/pkg/ginx"
	"archi/pkg/logger"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
)

type FeedHandler struct {
	svc feed.Service
	l   logger.Logger
}

func NewFeedHandler(svc feed.Service, l logger.Logger) *FeedHandler {
	return &FeedHandler{
		svc: svc,
		l:   l,
	}
}

func (h *FeedHandler) RegisterRoutes(e *gin.Engine) {
	g := e.Group("/feed")
	g.GET("/events", ginx.WrapClaims(h.GetFeedEventList))
}

func (h *FeedHandler) GetFeedEventList(c *gin.Context, uc jwt.UserClaims) (ginx.Result, error) {
	timestampStr := c.Query("timestamp")
	limitStr := c.Query("limit")

	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil || timestamp == 0 {
		timestamp = time.Now().Unix()
	}

	limit, err := strconv.ParseInt(limitStr, 10, 64)
	if err != nil || limit == 0 {
		limit = 10 // 默认值
	}

	events, err := h.svc.GetFeedEventList(c.Request.Context(), uc.Uid, timestamp, limit)
	if err != nil {
		return ginx.Result{
			Code: errs.FeedInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		Code: http.StatusOK,
		Msg:  "feed流获取成功",
		Data: events,
	}, nil
}
