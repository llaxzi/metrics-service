package handler

import (
	"github.com/gin-gonic/gin"
	"metrics-service/internal/server/retry"
	"metrics-service/internal/server/storage"
	"net/http"
)

type HTMLHandler interface {
	Get(ctx *gin.Context)
}

func NewHTMLHandler(storage storage.Storage, retryer retry.Retryer) HTMLHandler {
	return &htmlHandler{storage, retryer}
}

type htmlHandler struct {
	storage storage.Storage
	retryer retry.Retryer
}

func (h *htmlHandler) Get(ctx *gin.Context) {
	var metrics [][]string
	err := h.retryer.Retry(func() error {
		var err error
		metrics, err = h.storage.GetMetrics(ctx)
		return err
	})

	// Формируем HTML
	metricsHTML := "<h1>Metrics List</h1><div>"
	if len(metrics) == 0 {
		metricsHTML += "<p>No metrics available</p>"
	} else {
		for _, metric := range metrics {
			metricsHTML += "<p>" + metric[0] + ": " + metric[1] + "</p>" // где metric []string, [0] - metricName, [1] - metricVal
		}
	}
	metricsHTML += "</div>"

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	ctx.Data(200, "text/html", []byte(metricsHTML))
}
