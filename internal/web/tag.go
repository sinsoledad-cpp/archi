package web

import (
	"archi/internal/domain"
	"archi/internal/service"
	"archi/internal/web/errs"
	"archi/internal/web/middleware/jwt"
	"archi/pkg/ginx"
	"archi/pkg/logger"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type TagHandler struct {
	svc service.TagService
	l   logger.Logger
}

func NewTagHandler(svc service.TagService, l logger.Logger) *TagHandler {
	return &TagHandler{
		svc: svc,
		l:   l,
	}
}

func (h *TagHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/tags")
	g.POST("/create", ginx.WrapBodyAndClaims(h.CreateTag))
	g.POST("/attach", ginx.WrapBodyAndClaims(h.AttachTags))
	g.GET("", ginx.WrapClaims(h.GetTags))
	g.GET("/:biz/:bizId", ginx.WrapClaims(h.GetBizTags))
}

type CreateTagReq struct {
	Name string `json:"name" binding:"required,max=10"`
}

func (h *TagHandler) CreateTag(ctx *gin.Context, req CreateTagReq, uc jwt.UserClaims) (ginx.Result, error) {
	id, err := h.svc.CreateTag(ctx, uc.Uid, req.Name)
	if err != nil {

		return ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		Code: http.StatusOK,
		Msg:  "创建标签成功",
		Data: id,
	}, nil
}

type AttachTagsReq struct {
	Biz   string  `json:"biz"`
	BizId int64   `json:"bizId"`
	Tags  []int64 `json:"tags"`
}

func (h *TagHandler) AttachTags(ctx *gin.Context, req AttachTagsReq, uc jwt.UserClaims) (ginx.Result, error) {
	err := h.svc.AttachTags(ctx, uc.Uid, req.Biz, req.BizId, req.Tags)
	if err != nil {
		return ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		Code: http.StatusOK,
		Msg:  "绑定标签成功",
	}, nil
}

type TagVO struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

func (h *TagHandler) GetTags(ctx *gin.Context, uc jwt.UserClaims) (ginx.Result, error) {
	tags, err := h.svc.GetTags(ctx, uc.Uid)
	if err != nil {
		return ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		Code: http.StatusOK,
		Msg:  "获取标签成功",
		Data: slice.Map(tags, func(idx int, src domain.Tag) TagVO {
			return TagVO{
				Id:   src.Id,
				Name: src.Name,
			}
		}),
	}, nil
}

func (h *TagHandler) GetBizTags(ctx *gin.Context, uc jwt.UserClaims) (ginx.Result, error) {
	biz := ctx.Param("biz")
	bizId, err := strconv.ParseInt(ctx.Param("bizId"), 10, 64)
	if err != nil {
		return ginx.Result{
			Code: http.StatusBadRequest,
			Msg:  "参数错误",
		}, err
	}
	tags, err := h.svc.GetBizTags(ctx, uc.Uid, biz, bizId)
	if err != nil {
		return ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		Code: http.StatusOK,
		Msg:  "获取业务对象标签成功",
		Data: slice.Map(tags, func(idx int, src domain.Tag) TagVO {
			return TagVO{
				Id:   src.Id,
				Name: src.Name,
			}
		}),
	}, nil
}
