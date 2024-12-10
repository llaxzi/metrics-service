package handler

import (
	"github.com/gin-gonic/gin"
	"metrics-service/internal/server/storage"
	"net/http"
	"strconv"
)

type MetricsHandler interface {
	Update(ctx *gin.Context)
	Get(ctx *gin.Context)
}

func NewMetricsHandler(storage storage.MetricsStorage) MetricsHandler {
	return &metricsHandler{storage}
}

type metricsHandler struct {
	storage storage.MetricsStorage
}

func (h *metricsHandler) Update(ctx *gin.Context) {

	metricType := ctx.Param("metricType")
	metricName := ctx.Param("metricName")
	metricValStr := ctx.Param("metricVal")

	// Проверяем имя метрики
	if metricName == "" {
		ctx.String(http.StatusNotFound, "metric name is missing")
		return
	}

	// Парсим значение метрики в зависимости от типа
	// TODO: Вынести в сервис
	switch metricType {
	case "counter":
		metricVal, err := strconv.ParseInt(metricValStr, 10, 64)
		if err != nil {
			ctx.String(http.StatusBadRequest, "wrong metric value")
			return
		}
		h.storage.SetCounter(metricName, metricVal)

	case "gauge":
		metricVal, err := strconv.ParseFloat(metricValStr, 64)
		if err != nil {
			ctx.String(http.StatusBadRequest, "wrong metric value")
			return
		}
		h.storage.SetGauge(metricName, metricVal)

	default:
		ctx.String(http.StatusBadRequest, "invalid metric type")
		return
	}

	ctx.String(http.StatusOK, "updated successfully")
}

func (h *metricsHandler) Get(ctx *gin.Context) {

	metricName := ctx.Param("metricName")
	metricType := ctx.Param("metricType")

	// TODO: Вынести в сервис
	switch metricType {
	case "counter":
		metricVal, exists := h.storage.GetCounter(metricName)
		if !exists {
			ctx.String(http.StatusNotFound, "metric doesn't exist")
			return
		}
		ctx.String(http.StatusOK, strconv.FormatInt(metricVal, 10))
	case "gauge":
		metricVal, exists := h.storage.GetGauge(metricName)

		if !exists {
			ctx.String(http.StatusNotFound, "metric doesn't exist")
			return
		}
		ctx.String(http.StatusOK, strconv.FormatFloat(metricVal, 'f', -1, 64))
	}

}
