package handler

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"metrics-service/internal/server/models"
	"metrics-service/internal/server/retry"
	"metrics-service/internal/server/storage"
	"net/http"
)

type MetricsHandler interface {
	Update(ctx *gin.Context)
	Get(ctx *gin.Context)
	UpdateJSON(ctx *gin.Context)
	GetJSON(ctx *gin.Context)
	Ping(ctx *gin.Context)
	UpdateBatch(ctx *gin.Context)
}

func NewMetricsHandler(storage storage.Storage, retryer retry.Retryer, isSync bool) MetricsHandler {
	return &metricsHandler{storage, retryer, isSync}
}

type metricsHandler struct {
	storage storage.Storage
	retryer retry.Retryer
	isSync  bool
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
	err := h.retryer.Retry(func() error {
		return h.storage.Update(ctx, metricType, metricName, metricValStr)
	})

	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	// Сохраняем на диск при синхронном режиме
	if h.isSync {
		err = h.storage.Save()
		if err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
		}
	}

	ctx.String(http.StatusOK, "updated successfully")
}

func (h *metricsHandler) Get(ctx *gin.Context) {

	metricName := ctx.Param("metricName")
	metricType := ctx.Param("metricType")

	var metricVal string
	err := h.retryer.Retry(func() error {
		var err error
		metricVal, err = h.storage.Get(ctx, metricType, metricName)
		return err
	})
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

	err = h.retryer.Retry(func() error {
		return h.storage.UpdateJSON(ctx, &requestData)
	})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Сохраняем на диск при синхронном режиме
	if h.isSync {
		err = h.storage.Save()
		if err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
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

	if requestData.MType != "counter" && requestData.MType != "gauge" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid metric type"})
		return
	}

	err = h.retryer.Retry(func() error {
		return h.storage.GetJSON(ctx, &requestData)
	})

	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, requestData)
}

func (h *metricsHandler) Ping(ctx *gin.Context) {
	err := h.retryer.Retry(func() error {
		return h.storage.Ping(ctx)
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, "ok")
}

func (h *metricsHandler) UpdateBatch(ctx *gin.Context) {
	var metrics []models.Metrics
	dec := json.NewDecoder(ctx.Request.Body)
	err := dec.Decode(&metrics)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}
	err = h.retryer.Retry(func() error {
		return h.storage.UpdateBatch(ctx, metrics)
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Сохраняем на диск при синхронном режиме
	if h.isSync {
		err = h.storage.Save()
		if err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "updated successfully"})
}
