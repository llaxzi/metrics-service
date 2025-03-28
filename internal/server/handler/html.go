package handler

import (
	"github.com/llaxzi/retryables/v2"
	"net/http"

	"github.com/gin-gonic/gin"

	"metrics-service/internal/server/storage"
)

// IHTMLHandler определяет интерфейс для обработки HTTP-запросов, связанных с HTML отдачей метрик.
type IHTMLHandler interface {
	Get(ctx *gin.Context)
}

// NewHTMLHandler создает новый экземпляр IHTMLHandler
func NewHTMLHandler(storage storage.Storage, retryer *retryables.Retryer) IHTMLHandler {
	return &HTMLHandler{storage, retryer}
}

// HTMLHandler реализует интерфейс IHTMLHandler и отвечает за обработку HTTP-запросов, связанных с HTML отдачей метрик.
type HTMLHandler struct {
	storage storage.Storage
	retryer *retryables.Retryer
}

// Get возвращает все метрики в HTML формате.
func (h *HTMLHandler) Get(ctx *gin.Context) {
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
