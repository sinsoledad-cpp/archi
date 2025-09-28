package web

import (
	"github.com/gin-gonic/gin"
)

type Handler interface {
	RegisterRoutes(e *gin.Engine)
}
