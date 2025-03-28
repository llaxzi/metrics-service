package handler

import (
	"encoding/json"
	"github.com/llaxzi/retryables/v2"
	"net/http"

	"github.com/gin-gonic/gin"

	"metrics-service/internal/server/models"
	"metrics-service/internal/server/storage"
)

// IMetricsHandler определяет интерфейс для обработки HTTP-запросов к метрикам.
type IMetricsHandler interface {
	Update(ctx *gin.Context)
	Get(ctx *gin.Context)
	UpdateJSON(ctx *gin.Context)
	GetJSON(ctx *gin.Context)
	Ping(ctx *gin.Context)
	UpdateBatch(ctx *gin.Context)
}

// NewMetricsHandler создает новый экземпляр IMetricsHandler
//
// Storage - хранилище метрик.
// Retryer - экземпляр retry.
// IsSync - флаг синхронного сохранения на диск.
func NewMetricsHandler(storage storage.Storage, retryer *retryables.Retryer, isSync bool) IMetricsHandler {
	return &MetricsHandler{storage, retryer, isSync}
}

// MetricsHandler реализует интерфейс IMetricsHandler и отвечает за обработку HTTP-запросов к метрикам.
type MetricsHandler struct {
	storage storage.Storage
	retryer *retryables.Retryer
	isSync  bool
}

// Update обновляет значение метрики по URL параметрам запроса.
func (h *MetricsHandler) Update(ctx *gin.Context) {

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

// Get возвращает значение метрики по имени и типу.
func (h *MetricsHandler) Get(ctx *gin.Context) {

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

// UpdateJSON обновляет метрику, принимая JSON в теле запроса.
func (h *MetricsHandler) UpdateJSON(ctx *gin.Context) {

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

// GetJSON возвращает значение метрики в формате JSON.
func (h *MetricsHandler) GetJSON(ctx *gin.Context) {

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

// Ping проверяет доступность хранилища.
func (h *MetricsHandler) Ping(ctx *gin.Context) {
	err := h.retryer.Retry(func() error {
		return h.storage.Ping(ctx)
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, "ok")
}

// UpdateBatch обновляет несколько метрик из массива JSON-объектов.
func (h *MetricsHandler) UpdateBatch(ctx *gin.Context) {

	// Избегаем reflect.growslice поэтапным декодированием.
	// Не сказал бы, что проблема была критичная (или вообще была), скорее просто пощупать профилирование
	var raw []json.RawMessage
	if err := json.NewDecoder(ctx.Request.Body).Decode(&raw); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	metrics := make([]models.Metrics, 0, len(raw))
	for _, r := range raw {
		var m models.Metrics
		if err := json.Unmarshal(r, &m); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid metric in array"})
			return
		}
		metrics = append(metrics, m)
	}

	err := h.retryer.Retry(func() error {
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
