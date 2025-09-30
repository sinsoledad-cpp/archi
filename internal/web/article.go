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
	"golang.org/x/sync/errgroup"
	"net/http"
	"strconv"
	"time"
)

type ArticleHandler struct {
	ArtSvc   service.ArticleService //art
	interSvc service.InteractiveService
	l        logger.Logger
	biz      string
}

func (a *ArticleHandler) RegisterRoutes(e *gin.Engine) {
	g := e.Group("/articles")

	//g.PUT("/", a.Edit)
	g.POST("/edit", ginx.WrapBodyAndClaims(a.Edit))
	g.POST("/publish", ginx.WrapBodyAndClaims(a.Publish))
	g.POST("/withdraw", ginx.WrapBodyAndClaims(a.Withdraw))

	// 创作者接口
	g.GET("/detail/:id", ginx.WrapClaims(a.Detail))
	// 按照道理来说，这边就是 GET 方法
	// /list?offset=?&limit=?
	g.POST("/list", ginx.WrapBodyAndClaims(a.List))

	pub := g.Group("/pub")
	pub.GET("/:id", ginx.WrapClaims(a.PubDetail))
	// 传入一个参数，true 就是点赞, false 就是不点赞
	pub.POST("/like", ginx.WrapBodyAndClaims(a.Like))
	pub.POST("/collect", ginx.WrapBodyAndClaims(a.Collect))
}

type ArticleEditReq struct {
	ID      int64
	Title   string `json:"title"`
	Content string `json:"content"`
}

// Edit 接收 Article 输入，返回一个 ID，文章的 ID
func (a *ArticleHandler) Edit(ctx *gin.Context, req ArticleEditReq, uc jwt.UserClaims) (ginx.Result, error) {
	id, err := a.ArtSvc.Save(ctx, domain.Article{
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

func (a *ArticleHandler) Publish(ctx *gin.Context, req PublishReq, uc jwt.UserClaims) (ginx.Result, error) {
	//val, ok := ctx.Get("user")
	//if !ok {
	//	ctx.JSON(http.StatusOK, Result{
	//		Code: 4,
	//		Msg:  "未登录",
	//	})
	//	return
	//}
	id, err := a.ArtSvc.Publish(ctx, domain.Article{
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

func (a *ArticleHandler) Withdraw(ctx *gin.Context, req ArticleWithdrawReq, uc jwt.UserClaims) (ginx.Result, error) {
	err := a.ArtSvc.Withdraw(ctx, uc.Uid, req.Id)
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

func (a *ArticleHandler) Detail(ctx *gin.Context, uc jwt.UserClaims) (ginx.Result, error) {
	idstr := ctx.Param("id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		a.l.Warn("查询文章失败，id 格式不对", logger.String("id", idstr), logger.Error(err))
		return ginx.Result{
			Code: errs.UserInvalidInput,
			Msg:  "id 参数错误",
		}, err
	}
	art, err := a.ArtSvc.GetById(ctx, id)
	if err != nil {
		a.l.Error("查询文章失败", logger.Int64("id", id), logger.Error(err))
		return ginx.Result{
			Code: errs.UserInvalidInput,
			Msg:  "系统错误",
		}, err
	}

	if art.Author.ID != uc.Uid {
		// 有人在搞鬼
		a.l.Error("非法查询文章", logger.Int64("id", id), logger.Int64("uid", uc.Uid))
		return ginx.Result{
			Code: errs.UserInvalidInput,
			Msg:  "系统错误",
		}, err
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
	return ginx.Result{
		Code: http.StatusOK,
		Msg:  "查询成功",
		Data: vo,
	}, nil
}

type Page struct {
	Limit  int
	Offset int
}

func (a *ArticleHandler) List(ctx *gin.Context, page Page, uc jwt.UserClaims) (ginx.Result, error) {
	arts, err := a.ArtSvc.GetByAuthor(ctx, uc.Uid, page.Offset, page.Limit)
	if err != nil {
		a.l.Error("查找文章列表失败",
			logger.Error(err),
			logger.Int("offset", page.Offset),
			logger.Int("limit", page.Limit),
			logger.Int64("uid", uc.Uid))
		return ginx.Result{
			Code: errs.ArticleInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	data := slice.Map[domain.Article, ArticleVo](arts, func(idx int, src domain.Article) ArticleVo {
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
	})
	return ginx.Result{
		Code: http.StatusOK,
		Msg:  "查询成功",
		Data: data,
	}, nil
}

func (a *ArticleHandler) PubDetail(ctx *gin.Context, uc jwt.UserClaims) (ginx.Result, error) {
	idstr := ctx.Param("id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		a.l.Warn("查询文章失败，id 格式不对",
			logger.String("id", idstr),
			logger.Error(err))
		return ginx.Result{
			Code: errs.ArticleInvalidInput,
			Msg:  "id 参数错误",
		}, err
	}

	var (
		eg   errgroup.Group
		art  domain.Article
		intr domain.Interactive
	)

	eg.Go(func() error {
		var er error
		art, er = a.ArtSvc.GetPubById(ctx, id, uc.Uid)
		return er
	})
	eg.Go(func() error {
		var er error
		intr, er = a.interSvc.Get(ctx, a.biz, id, uc.Uid)
		return er
	})

	// 等待结果
	err = eg.Wait()
	if err != nil {
		a.l.Error("查询文章失败，系统错误",
			logger.Int64("aid", id),
			logger.Int64("uid", uc.Uid),
			logger.Error(err))
		return ginx.Result{
			Code: errs.ArticleInternalServerError,
			Msg:  "系统错误",
		}, err
	}

	//go func() {
	// 1. 如果你想摆脱原本主链路的超时控制，你就创建一个新的
	// 2. 如果你不想，你就用 ctx
	//newCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	//defer cancel()
	//er := a.interSvc.IncrReadCnt(newCtx, a.biz, art.Id)
	//if er != nil {
	//	a.l.Error("更新阅读数失败",
	//		logger.Int64("aid", art.Id),
	//		logger.Error(err))
	//}
	//}()

	data := ArticleVo{
		ID:    art.ID,
		Title: art.Title,

		Content:    art.Content,
		AuthorId:   art.Author.ID,
		AuthorName: art.Author.Name,
		ReadCnt:    intr.ReadCnt,
		CollectCnt: intr.CollectCnt,
		LikeCnt:    intr.LikeCnt,
		Liked:      intr.Liked,
		Collected:  intr.Collected,

		Status: art.Status.ToUint8(),
		Ctime:  art.Ctime.Format(time.DateTime),
		Utime:  art.Utime.Format(time.DateTime),
	}

	return ginx.Result{
		Code: http.StatusOK,
		Msg:  "查询成功",
		Data: data,
	}, nil
}

type ArticleLikeReq struct {
	Id   int64 `json:"id"`
	Like bool  `json:"like"` // true 是点赞，false 是不点赞
}

func (a *ArticleHandler) Like(c *gin.Context, req ArticleLikeReq, uc jwt.UserClaims) (ginx.Result, error) {
	var err error
	if req.Like {
		// 点赞
		err = a.interSvc.Like(c, a.biz, req.Id, uc.Uid)
	} else {
		// 取消点赞
		err = a.interSvc.CancelLike(c, a.biz, req.Id, uc.Uid)
	}
	if err != nil {
		return ginx.Result{
			Code: errs.ArticleInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		Code: http.StatusOK,
		Msg:  "喜欢/取消喜欢成功",
	}, nil
}

type ArticleCollectReq struct {
	Id      int64 `json:"id"`
	Cid     int64 `json:"cid"`
	Collect bool  `json:"collect"`
}

func (a *ArticleHandler) Collect(ctx *gin.Context, req ArticleCollectReq, uc jwt.UserClaims) (ginx.Result, error) {
	var err error
	if req.Collect {
		// 点赞
		err = a.interSvc.Collect(ctx, a.biz, req.Id, req.Cid, uc.Uid)
	} else {
		// 取消点赞
		err = a.interSvc.Collect(ctx, a.biz, req.Id, req.Cid, uc.Uid)
	}
	if err != nil {
		return ginx.Result{
			Code: errs.ArticleInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		Code: http.StatusOK,
		Msg:  "收藏/取消收藏成功",
	}, nil
}
