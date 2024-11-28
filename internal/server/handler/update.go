package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"metrics-service/internal/server/storage"
	"net/http"
	"strconv"
)

func NewUpdateHandler(storage storage.MetricsStorage) MetricsHandler {
	return &updateHandler{storage}
}

type updateHandler struct {
	storage storage.MetricsStorage
}

func (h *updateHandler) Update(ctx *gin.Context) {

	fmt.Println("update handler")

	// Проверка Content-Type
	if ctx.GetHeader("Content-Type") != "text/plain" {
		ctx.String(http.StatusUnsupportedMediaType, "unsupported content type")
		return
	}

	metricType := ctx.Param("metricType")
	metricName := ctx.Param("metricName")
	metricValStr := ctx.Param("metricVal")

	// Проверяем имя метрики
	if metricName == "" {
		ctx.String(http.StatusNotFound, "metric name is missing")
		return
	}

	// Парсим значение метрики в зависимости от типа
	switch metricType {
	case "counter":
		metricVal, err := strconv.ParseInt(metricValStr, 10, 64)
		if err != nil {
			ctx.String(http.StatusBadRequest, "wrong metric value")
			return
		}
		// TODO: Передать в сервис для обработки
		h.storage.SetCounter(metricName, metricVal)

	case "gauge":
		metricVal, err := strconv.ParseFloat(metricValStr, 64)
		if err != nil {
			ctx.String(http.StatusBadRequest, "wrong metric value")
			return
		}
		// TODO: Передать в сервис для обработки
		h.storage.SetGauge(metricName, metricVal)

	default:
		ctx.String(http.StatusBadRequest, "invalid metric type")
		return
	}

	ctx.String(http.StatusOK, "updated successfully")
}
