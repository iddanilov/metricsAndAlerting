package http

import (
	"github.com/gin-gonic/gin"
	serverApp "github.com/iddanilov/metricsAndAlerting/internal/app/server"
	"github.com/iddanilov/metricsAndAlerting/internal/pkg/middleware"
)

type RouterGroup struct {
	rg *gin.RouterGroup
	uc serverApp.Usecase
}

// NewRouterGroup - create new gin route group
func NewRouterGroup(rg *gin.RouterGroup, serverUseCase serverApp.Usecase) *RouterGroup {
	return &RouterGroup{
		rg: rg,
		uc: serverUseCase,
	}
}

func (rg *RouterGroup) Routes() {
	h := NewHandlers(rg.uc)
	group := rg.rg.Group("/")
	group.Use()
	{
		group.GET("/", middleware.Middleware(h.MetricList))
		group.POST("/update/:type/:name/:value", middleware.Middleware(h.uc.UpdateMetricByPath))
		group.POST("/update/", middleware.Middleware(h.uc.UpdateMetric))
		group.POST("/updates/", middleware.Middleware(h.uc.UpdateMetrics))
		group.POST("/value/", middleware.Middleware(h.uc.GetMetric))
		group.GET("/value/:type/:name", middleware.Middleware(h.uc.GetMetricByPath))
		group.GET("/ping", middleware.Middleware(h.uc.Ping))
	}
}
