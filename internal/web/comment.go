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
	"time"
)

type CommentHandler struct {
	svc service.CommentService
	l   logger.Logger
}

func NewCommentHandler(svc service.CommentService, l logger.Logger) *CommentHandler {
	return &CommentHandler{
		svc: svc,
		l:   l,
	}
}

func (h *CommentHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/comments")
	g.POST("/create", ginx.WrapBodyAndClaims(h.CreateComment))
	g.POST("/delete", ginx.WrapBodyAndClaims(h.DeleteComment))
	g.POST("/list", ginx.WrapBody(h.GetCommentList))
	g.POST("/replies", ginx.WrapBody(h.GetMoreReplies))
}

type CreateCommentReq struct {
	Biz     string `json:"biz"`
	BizID   int64  `json:"biz_id"`
	Content string `json:"content"`
	// 回复评论(父评论)的id
	ParentID int64 `json:"parent_id"`
	// 根评论id
	RootID int64 `json:"root_id"`
}

func (h *CommentHandler) CreateComment(ctx *gin.Context, req CreateCommentReq, uc jwt.UserClaims) (ginx.Result, error) {
	comment, err := h.svc.CreateComment(ctx, domain.Comment{
		Commentator: domain.CommentatorInfo{
			ID: uc.Uid,
		},
		Biz:     req.Biz,
		BizID:   req.BizID,
		Content: req.Content,
		RootComment: &domain.Comment{
			Id: req.RootID,
		},
		ParentComment: &domain.Comment{
			Id: req.ParentID,
		},
	})
	if err != nil {
		return ginx.Result{
			Code: errs.CommentInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	vo := h.toVo(comment)
	return ginx.Result{
		Code: http.StatusOK,
		Msg:  "评论发表成功",
		Data: vo,
	}, nil
}

type DeleteCommentReq struct {
	ID int64 `json:"id"`
}

func (h *CommentHandler) DeleteComment(ctx *gin.Context, req DeleteCommentReq, uc jwt.UserClaims) (ginx.Result, error) {
	// 在service层没有校验评论是否是自己的，可以在这层补充
	err := h.svc.DeleteComment(ctx, req.ID)
	if err != nil {
		return ginx.Result{
			Code: errs.CommentInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		Code: http.StatusOK,
		Msg:  "删除成功",
	}, nil
}

type CommentVo struct {
	Id int64 `json:"id"`
	// 评论者
	Commentator domain.CommentatorInfo `json:"commentator"`
	// 评论对象
	Biz   string `json:"biz"`
	BizID int64  `json:"biz_id"`
	// 评论内容
	Content string `json:"content"`
	// 父评论
	ParentComment *CommentVo  `json:"parent_comment"`
	Children      []CommentVo `json:"children"`
	CTime         string      `json:"ctime"`
}

type GetCommentListReq struct {
	Biz   string `json:"biz"`
	BizID int64  `json:"biz_id"`
	MinID int64  `json:"min_id"`
	Limit int64  `json:"limit"`
}

func (h *CommentHandler) GetCommentList(ctx *gin.Context, req GetCommentListReq) (ginx.Result, error) {
	list, err := h.svc.GetCommentList(ctx, req.Biz, req.BizID, req.MinID, req.Limit)
	if err != nil {
		return ginx.Result{
			Code: errs.CommentInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		Code: http.StatusOK,
		Msg:  "加载评论成功",
		Data: slice.Map(list, func(idx int, src domain.Comment) CommentVo {
			return h.toVo(src)
		}),
	}, nil
}

type GetMoreRepliesReq struct {
	RID   int64 `json:"rid"`
	MaxID int64 `json:"max_id"`
	Limit int64 `json:"limit"`
}

func (h *CommentHandler) GetMoreReplies(ctx *gin.Context, req GetMoreRepliesReq) (ginx.Result, error) {
	list, err := h.svc.GetMoreReplies(ctx, req.RID, req.MaxID, req.Limit)
	if err != nil {
		return ginx.Result{
			Code: errs.CommentInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		Data: slice.Map(list, func(idx int, src domain.Comment) CommentVo {
			return h.toVo(src)
		}),
	}, nil
}

func (h *CommentHandler) toVo(c domain.Comment) CommentVo {
	var children []CommentVo
	if len(c.Children) > 0 {
		children = slice.Map(c.Children, func(idx int, src domain.Comment) CommentVo {
			return h.toVo(src)
		})
	}
	res := CommentVo{
		Id:          c.Id,
		Commentator: c.Commentator,
		Biz:         c.Biz,
		BizID:       c.BizID,
		Content:     c.Content,
		Children:    children,
		CTime:       c.CTime.Format(time.DateTime),
	}
	if c.ParentComment != nil {
		res.ParentComment = &CommentVo{
			Id: c.ParentComment.Id,
		}
	}
	return res
}
