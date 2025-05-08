package router

import (
	"report/service"

	"github.com/gin-gonic/gin"
)

func IntiateRoutes(router *gin.Engine) {

	ApiGroup := router.Group("/api")
	ApiGroup.POST("/summeriseReport", service.GenerateReportSummary)
	ApiGroup.POST("/predict",service)
}
