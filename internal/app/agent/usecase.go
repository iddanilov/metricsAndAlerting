package agent

import (
	"context"
	"github.com/iddanilov/metricsAndAlerting/internal/models"
	"github.com/shirou/gopsutil/v3/mem"
	"runtime"
)

type AgentUseCase interface {
	SendMetricByPath(params models.Metrics) error
	SendMetric(metrics models.Metrics) error
	SendMetrics(jobs <-chan []models.Metrics, cfg *models.Agent)
	GetRuntimeStat(metrics *runtime.MemStats)
	GetVirtualMemoryStat(ctx context.Context) *mem.VirtualMemoryStat
}
