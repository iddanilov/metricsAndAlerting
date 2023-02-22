package http

import (
	"github.com/gin-gonic/gin"
	serverApp "github.com/iddanilov/metricsAndAlerting/internal/app/server"
	"github.com/iddanilov/metricsAndAlerting/internal/pkg/middleware"
	"log"
	"net/http"
)

type handlers struct {
	uc serverApp.Usecase
}

func NewHandlers(serverUseCase serverApp.Usecase) handlers {
	return handlers{
		uc: serverUseCase,
	}
}

func (h *handlers) MetricList(c *gin.Context) ([]byte, error) {
	r := c.Request
	w := c.Writer
	log.Println("UpdateMetricByPath Metrics", r.URL)
	mType := c.Params.ByName("type")
	name := c.Params.ByName("name")
	mValue := c.Params.ByName("value")
	if mType == "" || name == "" || mValue == "" {
		w.WriteHeader(http.StatusNotFound)
		return nil, middleware.ErrNotFound
	}

	return h.uc.MetricList(c)

}
