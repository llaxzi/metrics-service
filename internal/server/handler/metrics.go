package handler

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"metrics-service/internal/server/models"
	"metrics-service/internal/server/service"
	"net/http"
)

type MetricsHandler interface {
	Update(ctx *gin.Context)
	Get(ctx *gin.Context)
	UpdateJSON(ctx *gin.Context)
	GetJSON(ctx *gin.Context)
	Ping(ctx *gin.Context)
}

func NewMetricsHandler(service service.MetricsService) MetricsHandler {
	return &metricsHandler{service}
}

type metricsHandler struct {
	service service.MetricsService
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

	// Обновляем значение метрики
	err := h.service.Update(metricType, metricName, metricValStr)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	// Сохраняем на диск при синхронном режиме
	err = h.service.Save()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.String(http.StatusOK, "updated successfully")
}

func (h *metricsHandler) Get(ctx *gin.Context) {

	metricName := ctx.Param("metricName")
	metricType := ctx.Param("metricType")

	metricVal, err := h.service.Get(metricType, metricName)
	if err != nil {
		ctx.String(http.StatusNotFound, err.Error())
		return
	}
	ctx.String(http.StatusOK, metricVal)

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

	if requestData.MType != "counter" && requestData.MType != "gauge" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid metric type"})
		return
	}

	if requestData.MType == "counter" && requestData.Delta == nil || requestData.MType == "gauge" && requestData.Value == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	err = h.service.UpdateJSON(&requestData)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Сохраняем на диск при синхронном режиме
	err = h.service.Save()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
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

	if requestData.MType != "counter" && requestData.MType != "gauge" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid metric type"})
		return
	}

	err = h.service.GetJSON(&requestData)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, requestData)
}

// repository

func (h *metricsHandler) Ping(ctx *gin.Context) {
	err := h.service.Ping()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, "ok")
}
