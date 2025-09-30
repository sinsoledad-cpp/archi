package web

import (
	"archi/internal/domain"
	"archi/internal/service"
	"archi/internal/web/errs"
	"archi/internal/web/middleware/jwt"
	"archi/pkg/ginx"
	"archi/pkg/logger"
	"fmt"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
)

type ArticleHandler struct {
	svc service.ArticleService //art
	// TODO 互动模块
	//intrSvc service.InteractiveService
	l   logger.Logger
	biz string
}
type ArticleEditReq struct {
	ID      int64
	Title   string `json:"title"`
	Content string `json:"content"`
}

// Edit 接收 Article 输入，返回一个 ID，文章的 ID
func (h *ArticleHandler) Edit(ctx *gin.Context, req ArticleEditReq, uc jwt.UserClaims) (ginx.Result, error) {
	id, err := h.svc.Save(ctx, domain.Article{
		ID:      req.ID,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			ID: uc.Uid,
		},
	})
	if err != nil {
		return ginx.Result{
			Code: errs.ArticleInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		Code: http.StatusOK,
		Msg:  "编辑成功",
		Data: id,
	}, nil
}

type PublishReq struct {
	ID      int64
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (h *ArticleHandler) Publish(ctx *gin.Context,
	req PublishReq,
	uc jwt.UserClaims) (ginx.Result, error) {
	//val, ok := ctx.Get("user")
	//if !ok {
	//	ctx.JSON(http.StatusOK, Result{
	//		Code: 4,
	//		Msg:  "未登录",
	//	})
	//	return
	//}
	id, err := h.svc.Publish(ctx, domain.Article{
		ID:      req.ID,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			ID: uc.Uid,
		},
	})
	if err != nil {
		return ginx.Result{
			Code: errs.ArticleInternalServerError,
			Msg:  "系统错误",
		}, fmt.Errorf("发表文章失败 aid %d, uid %d %w", uc.Uid, req.ID, err)
	}
	return ginx.Result{
		Code: http.StatusOK,
		Msg:  "发表成功",
		Data: id,
	}, nil
}

type ArticleWithdrawReq struct {
	Id int64
}

func (h *ArticleHandler) Withdraw(ctx *gin.Context, req ArticleWithdrawReq, uc jwt.UserClaims) (ginx.Result, error) {
	err := h.svc.Withdraw(ctx, uc.Uid, req.Id)
	if err != nil {
		return ginx.Result{
			Code: errs.ArticleInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		Code: http.StatusOK,
		Msg:  "隐私成功",
	}, nil
}

type ArticleVo struct {
	ID         int64  `json:"id,omitempty"`
	Title      string `json:"title,omitempty"`
	Abstract   string `json:"abstract,omitempty"`
	Content    string `json:"content,omitempty"`
	AuthorId   int64  `json:"authorId,omitempty"`
	AuthorName string `json:"authorName,omitempty"`
	Status     uint8  `json:"status,omitempty"`
	Ctime      string `json:"ctime,omitempty"`
	Utime      string `json:"utime,omitempty"`

	ReadCnt    int64 `json:"readCnt"`
	LikeCnt    int64 `json:"likeCnt"`
	CollectCnt int64 `json:"collectCnt"`

	Liked     bool `json:"liked"`
	Collected bool `json:"collected"`
}

func (h *ArticleHandler) Detail(ctx *gin.Context) {
	idstr := ctx.Param("id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: errs.UserInvalidInput,
			Msg:  "id 参数错误",
		})
		h.l.Warn("查询文章失败，id 格式不对", logger.String("id", idstr), logger.Error(err))
		return
	}
	art, err := h.svc.GetById(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: errs.UserInvalidInput,
			Msg:  "系统错误",
		})
		h.l.Error("查询文章失败", logger.Int64("id", id), logger.Error(err))
		return
	}
	uc := ctx.MustGet("user").(jwt.UserClaims)
	if art.Author.ID != uc.Uid {
		// 有人在搞鬼
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: errs.UserInvalidInput,
			Msg:  "系统错误",
		})
		h.l.Error("非法查询文章", logger.Int64("id", id), logger.Int64("uid", uc.Uid))
		return
	}

	vo := ArticleVo{
		ID:    art.ID,
		Title: art.Title,
		//Abstract: art.Abstract(),

		Content:  art.Content,
		AuthorId: art.Author.ID,
		// 列表，你不需要
		Status: art.Status.ToUint8(),
		Ctime:  art.Ctime.Format(time.DateTime),
		Utime:  art.Utime.Format(time.DateTime),
	}
	ctx.JSON(http.StatusOK, ginx.Result{
		Code: http.StatusOK,
		Msg:  "查询成功",
		Data: vo,
	})
}

type Page struct {
	Limit  int
	Offset int
}

func (h *ArticleHandler) List(ctx *gin.Context) {
	var page Page
	if err := ctx.Bind(&page); err != nil {
		return
	}
	// 我要不要检测一下？
	uc := ctx.MustGet("user").(jwt.UserClaims)
	arts, err := h.svc.GetByAuthor(ctx, uc.Uid, page.Offset, page.Limit)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: errs.ArticleInternalServerError,
			Msg:  "系统错误",
		})
		h.l.Error("查找文章列表失败",
			logger.Error(err),
			logger.Int("offset", page.Offset),
			logger.Int("limit", page.Limit),
			logger.Int64("uid", uc.Uid))
		return
	}
	ctx.JSON(http.StatusOK, ginx.Result{
		Data: slice.Map[domain.Article, ArticleVo](arts, func(idx int, src domain.Article) ArticleVo {
			return ArticleVo{
				ID:       src.ID,
				Title:    src.Title,
				Abstract: src.Abstract(),

				//Content:  src.Content,
				AuthorId: src.Author.ID,
				// 列表，你不需要
				Status: src.Status.ToUint8(),
				Ctime:  src.Ctime.Format(time.DateTime),
				Utime:  src.Utime.Format(time.DateTime),
			}
		}),
	})
}
