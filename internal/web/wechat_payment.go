package web

import (
	"archi/internal/service/payment/wechat"
	"archi/pkg/ginx"
	"archi/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
)

type NativePayWechatHandler struct {
	handler   *notify.Handler
	l         logger.Logger
	nativeSvc *wechat.NativePaymentService
}

func NewNativePayWechatHandler(handler *notify.Handler, nativeSvc *wechat.NativePaymentService, l logger.Logger) *NativePayWechatHandler {
	return &NativePayWechatHandler{
		handler:   handler,
		nativeSvc: nativeSvc,
		l:         l}
}

func (h *NativePayWechatHandler) RegisterRoutes(server *gin.Engine) {
	//server.GET("/hello", func(context *gin.Context) {
	//	context.String(http.StatusOK, "我进来了")
	//})
	server.Any("/pay/callback", ginx.Wrap(h.HandleNative))
}

func (h *NativePayWechatHandler) HandleNative(ctx *gin.Context) (ginx.Result, error) {
	transaction := &payments.Transaction{}
	// 第一个返回值里面的内容我们暂时用不上
	_, err := h.handler.ParseNotifyRequest(ctx, ctx.Request, transaction)
	if err != nil {
		return ginx.Result{}, err
	}
	err = h.nativeSvc.HandleCallback(ctx, transaction)
	return ginx.Result{}, err
}
