package handler

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"metrics-service/internal/server/models"
	"metrics-service/internal/server/storage"
	"net/http"
	"strconv"
)

type MetricsHandler interface {
	Update(ctx *gin.Context)
	Get(ctx *gin.Context)
	UpdateJSON(ctx *gin.Context)
	GetJSON(ctx *gin.Context)
}

func NewMetricsHandler(storage storage.MetricsStorage, diskW storage.DiskWriter, isStoreInterval bool) MetricsHandler {
	return &metricsHandler{storage, diskW, isStoreInterval}
}

type metricsHandler struct {
	storage         storage.MetricsStorage
	diskW           storage.DiskWriter
	isStoreInterval bool
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

	if !h.isStoreInterval {
		err := h.diskW.Save()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
			return
		}
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

// JSON

func (h *metricsHandler) UpdateJSON(ctx *gin.Context) {

	contentType := ctx.GetHeader("Content-type")
	if contentType != "application/json" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid content type"})
		return
	}

	var requestData models.Metrics
	dec := json.NewDecoder(ctx.Request.Body)
	err := dec.Decode(&requestData)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	// Парсим значение метрики в зависимости от типа
	// TODO: Вынести в сервис
	switch requestData.MType {
	case "counter":
		h.storage.SetCounter(requestData.ID, *requestData.Delta)

		actualVal, exists := h.storage.GetCounter(requestData.ID)
		if !exists {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
			return
		}
		*requestData.Delta = actualVal
	case "gauge":
		h.storage.SetGauge(requestData.ID, *requestData.Value)
		actualVal, exists := h.storage.GetGauge(requestData.ID)
		if !exists {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
			return
		}
		*requestData.Value = actualVal

	default:
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid metric type"})
		return
	}

	if !h.isStoreInterval {
		err := h.diskW.Save()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
			return
		}
	}

	ctx.JSON(http.StatusOK, requestData)
}

func (h *metricsHandler) GetJSON(ctx *gin.Context) {

	contentType := ctx.GetHeader("Content-type")
	if contentType != "application/json" {
		ctx.JSON(http.StatusBadRequest, "invalid content type")
		return
	}

	var requestData models.Metrics
	dec := json.NewDecoder(ctx.Request.Body)
	err := dec.Decode(&requestData)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	// TODO: Вынести в сервис
	switch requestData.MType {
	case "counter":
		metricVal, exists := h.storage.GetCounter(requestData.ID)
		if !exists {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "metric doesn't exist"})
			return
		}
		requestData.Delta = &metricVal
	case "gauge":
		metricVal, exists := h.storage.GetGauge(requestData.ID)

		if !exists {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "metric doesn't exist"})
			return
		}
		requestData.Value = &metricVal
	}

	ctx.JSON(http.StatusOK, requestData)
}
