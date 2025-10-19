package web

import (
	"archi/internal/service"
	"archi/internal/web/errs"
	jwtware "archi/internal/web/middleware/jwt"
	"archi/pkg/ginx"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RewardHandler struct {
	rewardSvc  service.RewardService
	articleSvc service.ArticleService
}

func (h *RewardHandler) RegisterRoutes(server *gin.Engine) {
	rg := server.Group("/reward")
	rg.POST("/detail", ginx.WrapBodyAndClaims[GetRewardReq](h.GetReward))
}

type GetRewardReq struct {
	Rid int64
}

func (h *RewardHandler) GetReward(ctx *gin.Context, req GetRewardReq, uc jwtware.UserClaims) (ginx.Result, error) {
	resp, err := h.rewardSvc.GetReward(ctx, req.Rid, uc.Uid)
	if err != nil {
		return ginx.Result{
			Code: errs.RewardInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		Code: http.StatusOK,
		Msg:  "获取打赏信息成功",
		Data: resp.Status, // 暂时也就是只需要状态
	}, nil
}
