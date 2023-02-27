package task

import (
	"github.com/gin-gonic/gin"
)

//go:generate mockgen -package mock -destination usecase/mock/server_mock.go -source=usecase.go

// Usecase represent the metric and server's usecases
type Usecase interface {
	Ping(c *gin.Context) ([]byte, error)
	GetMetric(c *gin.Context) ([]byte, error)
	GetMetricByPath(c *gin.Context) ([]byte, error)
	MetricList(c *gin.Context) ([]byte, error)
	UpdateMetricByPath(c *gin.Context) ([]byte, error)
	UpdateMetric(c *gin.Context) ([]byte, error)
	UpdateMetrics(c *gin.Context) ([]byte, error)
}
