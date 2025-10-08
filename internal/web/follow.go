package web

import (
	"archi/internal/service"
	"archi/internal/web/errs"
	jwtware "archi/internal/web/middleware/jwt"
	"archi/pkg/ginx"
	"archi/pkg/logger"
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

// FollowHandler 负责处理关注关系的路由
type FollowHandler struct {
	svc service.FollowRelationService
	log logger.Logger
}

func NewFollowHandler(svc service.FollowRelationService, l logger.Logger) *FollowHandler {
	return &FollowHandler{
		svc: svc,
		log: l,
	}
}

// RegisterRoutes 注册与关注功能相关的所有路由
func (h *FollowHandler) RegisterRoutes(server *gin.Engine) {
	// 创建一个 /follow 的路由组，便于管理
	g := server.Group("/follows")

	// 定义具体的路由和处理函数
	g.POST("/relation", ginx.WrapBodyAndClaims(h.ToggleFollow))
	g.POST("/list/followees", ginx.WrapBodyAndClaims(h.FolloweeList))
	g.GET("/relation/:followee_id", ginx.WrapClaims(h.GetFollowRelation))
}

type FollowRelationReq struct {
	FolloweeID int64 `json:"followeeId"` // 被关注者的ID
	// 'Follow' 控制操作：true 为关注, false 为取消关注
	Follow bool `json:"follow"`
}

func (h *FollowHandler) ToggleFollow(ctx *gin.Context, req FollowRelationReq, uc jwtware.UserClaims) (ginx.Result, error) {
	followerID := uc.Uid
	if followerID == req.FolloweeID {
		return ginx.Result{
			Code: errs.FollowInvalidInput,
			Msg:  "不能关注/取消关注自己",
		}, errors.New("不能关注/取消关注自己")
	}

	if req.Follow {
		// 执行关注逻辑
		err := h.svc.Follow(ctx, followerID, req.FolloweeID)
		if err != nil {
			h.log.Error("关注用户失败", logger.Error(err),
				logger.Int64("follower", followerID),
				logger.Int64("followee", req.FolloweeID))
			return ginx.Result{Code: errs.FollowInternalServerError, Msg: "系统异常，请重试"}, err
		}
		return ginx.Result{Msg: "关注成功"}, nil
	} else {
		// 执行取消关注逻辑
		err := h.svc.CancelFollow(ctx, followerID, req.FolloweeID)
		if err != nil {
			h.log.Error("取消关注失败", logger.Error(err),
				logger.Int64("follower", followerID),
				logger.Int64("followee", req.FolloweeID))
			return ginx.Result{Code: errs.FollowInternalServerError, Msg: "系统异常，请重试"}, err
		}
		return ginx.Result{Msg: "取消关注成功"}, nil
	}
}

type FolloweeListPage struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// FolloweeList 获取某人关注的列表
func (h *FollowHandler) FolloweeList(ctx *gin.Context, page FolloweeListPage, uc jwtware.UserClaims) (ginx.Result, error) {
	list, err := h.svc.GetFollowee(ctx, uc.Uid, int64(page.Offset), int64(page.Limit))
	if err != nil {
		h.log.Error("获取关注列表失败", logger.Error(err), logger.Int64("uid", uc.Uid))
		return ginx.Result{
			Code: errs.FollowInternalServerError,
			Msg:  "系统异常，请重试",
		}, err
	}
	return ginx.Result{
		Code: http.StatusOK,
		Msg:  "获取关注列表成功",
		Data: list,
	}, nil
}

type FollowRelationVo struct {
	IsFollowed bool `json:"isFollowed"`
}

// GetFollowRelation 获取当前用户与另一个用户的关注关系
func (h *FollowHandler) GetFollowRelation(ctx *gin.Context, uc jwtware.UserClaims) (ginx.Result, error) {
	// 1. 从 URL 参数中解析出被查询用户的 ID (followee)
	followeeID, err := strconv.ParseInt(ctx.Param("followee_id"), 10, 64)
	if err != nil {
		return ginx.Result{
			Code: errs.FollowInvalidInput,
			Msg:  "参数错误",
		}, err
	}

	// 2. 从 JWT claims 中获取当前登录用户的 ID (follower)
	followerID := uc.Uid

	// 3. 调用 Service 层的方法
	_, err = h.svc.FollowInfo(ctx, followerID, followeeID)

	// 4. 根据 Service 返回的 error 来判断关系是否存在，并返回相应的结果
	if err != nil {
		// 如果错误是 gorm.ErrRecordNotFound，这在业务上是正常情况，
		// 意味着“没有关注关系”，而不是系统出错了。
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ginx.Result{
				Code: http.StatusOK,
				Msg:  "查询关注关系成功",
				Data: FollowRelationVo{IsFollowed: false},
			}, nil
		}
		// 如果是其他错误，则记录日志并返回系统异常
		h.log.Error("查询关注关系失败", logger.Error(err),
			logger.Int64("follower", followerID), logger.Int64("followee", followeeID))
		return ginx.Result{
			Code: errs.FollowInternalServerError,
			Msg:  "系统异常，请重试",
		}, err
	}
	// 5. 如果没有错误，说明关注关系存在
	return ginx.Result{
		Code: http.StatusOK,
		Msg:  "查询关注关系成功",
		Data: FollowRelationVo{IsFollowed: true},
	}, nil
}
