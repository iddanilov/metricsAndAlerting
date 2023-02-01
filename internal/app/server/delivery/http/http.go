package http

import (
	"github.com/gin-gonic/gin"
	serverapp "github.com/iddanilov/metricsAndAlerting/internal/app/server"
	"github.com/iddanilov/metricsAndAlerting/internal/pkg/middleware"
	"github.com/iddanilov/metricsAndAlerting/internal/pkg/repository/io"
)

type routerGroup struct {
	rg  *gin.RouterGroup
	s   *io.Storage
	key string
	uc  serverapp.Usecase
}

// NewRouterGroup - create new gin route group
func NewRouterGroup(rg *gin.RouterGroup, serverUseCase serverapp.Usecase) *routerGroup {
	return &routerGroup{
		rg: rg,
		//s:             s,
		//key:           key,
		uc: serverUseCase,
	}
}

func (h *routerGroup) Routes() {
	group := h.rg.Group("/")
	group.Use()
	{
		group.GET("/", middleware.Middleware(h.uc.MetricList))
		group.POST("/update/:type/:name/:value", middleware.Middleware(h.uc.UpdateMetricByPath))
		group.POST("/update/", middleware.Middleware(h.uc.UpdateMetric))
		group.POST("/updates/", middleware.Middleware(h.uc.UpdateMetrics))
		group.POST("/value/", middleware.Middleware(h.uc.GetMetric))
		group.GET("/value/:type/:name", middleware.Middleware(h.uc.GetMetricByPath))
		group.GET("/ping", middleware.Middleware(h.uc.Ping))
	}
}
