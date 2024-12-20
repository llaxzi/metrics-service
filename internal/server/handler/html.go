package handler

import (
	"github.com/gin-gonic/gin"
	"metrics-service/internal/server/storage"
)

type HTMLHandler interface {
	Get(ctx *gin.Context)
}

func NewHTMLHandler(storage storage.MetricsStorage) HTMLHandler {
	return &htmlHandler{storage}
}

type htmlHandler struct {
	storage storage.MetricsStorage
}

func (h *htmlHandler) Get(ctx *gin.Context) {
	metrics := h.storage.GetMetrics()

	// Формируем html

	metricsHTML := "<h1>Metrics List</h1><div>"
	if len(metrics) == 0 {
		metricsHTML += "<p>No metrics available</p>"
	} else {
		for _, metric := range metrics {
			metricsHTML += "<p>" + metric[0] + ": " + metric[1] + "</p>" // где metric []string, [0] - metricName, [1] - metricVal
		}
	}
	metricsHTML += "</div>"

	ctx.Data(200, "text/html", []byte(metricsHTML))
}
