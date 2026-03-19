package web

import (
	"archi/internal/domain"
	"archi/internal/service"
	"archi/internal/service/ai"
	"archi/internal/web/errs"
	"archi/internal/web/middleware/jwt"
	"archi/pkg/ginx"
	"archi/pkg/logger"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
)

type ArticleHandler struct {
	ArtSvc   service.ArticleService //art
	interSvc service.InteractiveService
	rankSvc  service.RankingService
	aiSvc    ai.AiService
	//rewardSvc service.RewardService
	l   logger.Logger
	biz string
}

// rewardSvc service.RewardService,
func NewArticleHandler(artSvc service.ArticleService, interSvc service.InteractiveService, rankSvc service.RankingService, aiSvc ai.AiService, l logger.Logger) *ArticleHandler {
	return &ArticleHandler{
		ArtSvc:   artSvc,
		interSvc: interSvc,
		rankSvc:  rankSvc,
		aiSvc:    aiSvc,
		//rewardSvc: rewardSvc,
		l:   l,
		biz: "article",
	}
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
	//pub.POST("/reward", ginx.WrapBodyAndClaims(a.Reward))

	pub.GET("/:id/ai-summary", ginx.Wrap(a.GetAiSummary))
	pub.POST("/:id/ai-qa", a.AnswerQA)

	// 创作者专用 AI 助手 (Agent 模式)
	g.POST("/author-helper", a.AuthorHelper)

	g.GET("/hot", ginx.Wrap(a.GetHot))
}

type AuthorHelperReq struct {
	ArticleID   int64  `json:"article_id"`
	Content     string `json:"content"`
	Instruction string `json:"instruction"`
}

type ArticleQAReq struct {
	Question string `json:"question"`
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

type ArticleListPage struct {
	Limit  int
	Offset int
}

func (a *ArticleHandler) List(ctx *gin.Context, page ArticleListPage, uc jwt.UserClaims) (ginx.Result, error) {
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
	//er := a.interSvc.IncrReadCnt(newCtx, a.biz, art.ID)
	//if er != nil {
	//	a.l.Error("更新阅读数失败",
	//		logger.Int64("aid", art.ID),
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

func (a *ArticleHandler) GetAiSummary(ctx *gin.Context) (ginx.Result, error) {
	idstr := ctx.Param("id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		return ginx.Result{
			Code: errs.ArticleInvalidInput,
			Msg:  "id 参数错误",
		}, err
	}

	// 1. 获取文章内容 (不传 UID，因为总结是公开的)
	art, err := a.ArtSvc.GetPubById(ctx, id, 0)
	if err != nil {
		a.l.Error("AI 总结查询文章失败", logger.Int64("id", id), logger.Error(err))
		return ginx.Result{
			Code: errs.ArticleInternalServerError,
			Msg:  "系统错误",
		}, err
	}

	// 2. 调用 AI 服务获取总结
	summary, err := a.aiSvc.GetArticleSummary(ctx, art)
	if err != nil {
		a.l.Error("获取 AI 总结失败", logger.Int64("id", id), logger.Error(err))
		return ginx.Result{
			Code: errs.AiServiceError,
			Msg:  "AI 课代表正在开小差，请稍后再试",
		}, err
	}

	return ginx.Result{
		Code: http.StatusOK,
		Msg:  "AI 总结生成成功",
		Data: summary,
	}, nil
}

// AnswerQA 针对单篇文章的“沉浸式笔记问答” (SSE 实现)
func (a *ArticleHandler) AnswerQA(ctx *gin.Context) {
	// 1. 获取文章 ID
	idstr := ctx.Param("id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: errs.ArticleInvalidInput,
			Msg:  "id 参数错误",
		})
		return
	}

	// 2. 解析请求体
	var req ArticleQAReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: errs.ArticleInvalidInput,
			Msg:  "请求格式错误",
		})
		return
	}

	// 3. 获取文章内容 (用于注入 Context)
	// 这里可以加上权限校验，目前先假设公开文章
	art, err := a.ArtSvc.GetPubById(ctx, id, 0)
	if err != nil {
		a.l.Error("AI 问答查询文章失败", logger.Int64("id", id), logger.Error(err))
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: errs.ArticleInternalServerError,
			Msg:  "系统错误",
		})
		return
	}

	// 4. 调用 AI Service 获取流式响应
	reader, err := a.aiSvc.AnswerQuestionStream(ctx, id, art.Content, req.Question)
	if err != nil {
		a.l.Error("启动 AI 问答失败", logger.Int64("id", id), logger.Error(err))
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: errs.AiServiceError,
			Msg:  "AI 助手暂不可用",
		})
		return
	}
	defer reader.Close()

	// 5. 设置 SSE 响应头
	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("Transfer-Encoding", "chunked")

	// 6. 循环读取流并发送
	for {
		chunk, err := reader.Recv()
		if err != nil {
			// 流结束
			break
		}

		content, ok := chunk.(string)
		if !ok {
			continue
		}

		// 发送数据块给前端
		ctx.SSEvent("message", content)
		ctx.Writer.Flush()
	}

	// 结束标识
	ctx.SSEvent("end", "done")
	ctx.Writer.Flush()
}

// AuthorHelper 创作者 AI 助手 (Agent 模式, SSE 流式返回)
func (a *ArticleHandler) AuthorHelper(ctx *gin.Context) {
	// 1. 获取用户 Claims
	val, ok := ctx.Get("user")
	if !ok {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	uc, ok := val.(jwt.UserClaims)
	if !ok {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// 2. 解析请求
	var req AuthorHelperReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: errs.ArticleInvalidInput,
			Msg:  "参数错误",
		})
		return
	}

	// 3. 调用 AI Service 获取 Agent 流
	reader, err := a.aiSvc.AuthorHelperStream(ctx, ai.AuthorHelperInput{
		ArticleID:   req.ArticleID,
		AuthorID:    uc.Uid,
		Content:     req.Content,
		Instruction: req.Instruction,
	})
	if err != nil {
		a.l.Error("启动 AI 创作助手失败", logger.Error(err))
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: errs.AiServiceError,
			Msg:  "AI 助手暂不可用",
		})
		return
	}
	defer reader.Close()

	// 4. 设置 SSE 响应头
	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("Transfer-Encoding", "chunked")

	// 5. 循环读取并发送
	for {
		chunk, err := reader.Recv()
		if err != nil {
			break
		}

		// ReAct Agent 的输出 chunk 可能是 *schema.Message (最终回答)
		// 或者是工具调用信息 (中间思考过程)
		// 这里简单处理，提取文本内容发送
		switch v := chunk.(type) {
		case string:
			ctx.SSEvent("message", v)
		case *schema.Message:
			ctx.SSEvent("message", v.Content)
		default:
			// 其他类型的 chunk (如中间思考) 也可以发送给前端显示
			// ctx.SSEvent("thought", fmt.Sprintf("%v", v))
		}
		ctx.Writer.Flush()
	}

	ctx.SSEvent("end", "done")
	ctx.Writer.Flush()
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
		err = a.interSvc.CancelCollect(ctx, a.biz, req.Id, req.Cid, uc.Uid)
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

type RewardReq struct {
	Id  int64 `json:"id"`
	Amt int64 `json:"amt"`
}

//	func (a *ArticleHandler) Reward(ctx *gin.Context, req RewardReq, uc jwt.UserClaims) (ginx.Result, error) {
//		art, err := a.ArtSvc.GetPubById(ctx.Request.Context(), req.ID, uc.Uid)
//		if err != nil {
//			return ginx.Result{
//				Code: errs.ArticleInternalServerError,
//				Msg:  "系统错误",
//			}, err
//		}
//
//		resp, err := a.rewardSvc.PreReward(ctx.Request.Context(), domain.Reward{
//			Uid: uc.Uid,
//			Target: domain.Target{
//				Biz:     a.biz,
//				BizID:   art.ID,
//				BizName: art.Title,
//				Uid:     art.Author.ID,
//			},
//			Amt: req.Amt,
//		})
//		if err != nil {
//			return ginx.Result{
//				Code: errs.ArticleInternalServerError,
//				Msg:  "系统错误",
//			}, err
//		}
//		return ginx.Result{
//			Code: http.StatusOK,
//			Msg:  "跳转打赏页面成功",
//			Data: map[string]any{
//				"codeURL": resp.URL,
//				"rid":     resp.Rid,
//			},
//		}, nil
//	}

func (a *ArticleHandler) GetHot(ctx *gin.Context) (ginx.Result, error) {
	// 直接调用 Ranking Service 获取热榜数据
	arts, err := a.rankSvc.GetTopN(ctx)
	if err != nil {
		a.l.Error("获取热榜文章失败", logger.Error(err))
		return ginx.Result{
			Code: errs.ArticleInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	// 2. 将领域对象（domain.Article）切片转换为视图对象（ArticleVo）切片
	vos := make([]ArticleVo, len(arts))
	for i, art := range arts {
		vos[i] = ArticleVo{
			ID:         art.ID,
			Title:      art.Title,
			Abstract:   art.Abstract(), // 调用方法生成摘要
			AuthorId:   art.Author.ID,
			AuthorName: art.Author.Name,
			Status:     art.Status.ToUint8(),
			Utime:      art.Utime.Format("2006-01-02 15:04:05"),
		}
	}
	// 将领域对象转换为视图对象（VO）返回给前端
	return ginx.Result{
		Code: http.StatusOK,
		Msg:  "热榜文章获取成功",
		Data: arts,
	}, nil
}
