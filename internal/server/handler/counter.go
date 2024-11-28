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

type counterHandler struct {
	storage storage.MetricsStorage
}

func NewCounterHandler(storage storage.MetricsStorage) MetricsHandler {
	return &counterHandler{storage}
}

func (h *counterHandler) Update(ctx *gin.Context) {

	// Проверяем Content-Type
	if ctx.GetHeader("Content-Type") != "text/plain" {

		ctx.String(http.StatusUnsupportedMediaType, "unsupported content type")
		return
	}

	metricName := ctx.Param("metricName")
	metricValStr := ctx.Param("metricVal")

	metricVal, err := strconv.ParseInt(metricValStr, 10, 64)
	if err != nil {
		ctx.String(http.StatusBadRequest, "wrong URL")
		return
	}

	// TODO вынести в сервис
	h.storage.SetCounter(metricName, metricVal)

	ctx.String(http.StatusOK, "updated successfully")
}

func (h *counterHandler) Get(ctx *gin.Context) {
	// Проверяем Content-Type
	if ctx.GetHeader("Content-Type") != "text/plain" {

		ctx.String(http.StatusUnsupportedMediaType, "unsupported content type")
		return
	}

	metricName := ctx.Param("metricName")

	metricVal, exists := h.storage.GetCounter(metricName)

	if !exists {
		ctx.String(http.StatusNotFound, "metric doesn't exist")
		return
	}

	ctx.String(http.StatusOK, strconv.FormatInt(metricVal, 10))

}
