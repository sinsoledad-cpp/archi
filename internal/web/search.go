package web

import (
	"archi/internal/service"
	"archi/internal/web/middleware/jwt"
	"archi/pkg/ginx"
	"github.com/gin-gonic/gin"
)

// SearchHandler 负责处理搜索相关的HTTP请求
type SearchHandler struct {
	svc service.SearchService
}

// NewSearchHandler 创建一个新的 SearchHandler 实例
func NewSearchHandler(svc service.SearchService) *SearchHandler {
	return &SearchHandler{
		svc: svc,
	}
}

// RegisterRoutes 注册搜索相关的路由
func (h *SearchHandler) RegisterRoutes(e *gin.Engine) {
	sg := e.Group("/api/search")
	// 使用 POST /api/search/ 接口进行搜索，需要JWT认证
	sg.POST("/", ginx.WrapBodyAndClaims(h.Search))
}

// SearchReq 定义了搜索请求的结构体
type SearchReq struct {
	Expression string `json:"expression"`
}

// Search 执行搜索
// 前端需要以POST方式请求，并携带JSON body: {"expression": "搜索关键词"}
func (h *SearchHandler) Search(ctx *gin.Context, req SearchReq, uc jwt.UserClaims) (ginx.Result, error) {
	// 调用业务层的搜索方法
	res, err := h.svc.Search(ctx, uc.Uid, req.Expression)
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		Data: res,
	}, nil
}
