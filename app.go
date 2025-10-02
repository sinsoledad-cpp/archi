package main

import (
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
)

type App struct {
	engine *gin.Engine
	//consumers []event.Consumer
	cron *cron.Cron
}
