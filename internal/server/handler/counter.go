package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"metrics-service/internal/server/storage"
	"net/http"
	"strconv"
	"strings"
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

	// Парсим url
	/*
		формат: http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
		req.URL.Path возвращает /update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
	*/
	partsURL := strings.Split(strings.TrimPrefix(ctx.Request.URL.Path, "/"), "/") // убираем первый / и сплитим

	if len(partsURL) != 4 {
		ctx.String(http.StatusNotFound, "Not found")
		return
	}

	metricName := partsURL[2]
	metricValStr := partsURL[3]

	metricVal, err := strconv.ParseInt(metricValStr, 10, 64)
	if err != nil {
		ctx.String(http.StatusBadRequest, "wrong URL")
		return
	}

	// TODO вынести в сервис
	h.storage.SetCounter(metricName, metricVal)

	fmt.Print("PollCount= ")
	fmt.Println(h.storage.GetCounter("PollCount"))

	ctx.String(http.StatusOK, "updated successfully")
}

func (h *counterHandler) Get(ctx *gin.Context) {
	// Проверяем Content-Type
	if ctx.GetHeader("Content-Type") != "text/plain" {

		ctx.String(http.StatusUnsupportedMediaType, "unsupported content type")
		return
	}

	partsURL := strings.Split(strings.TrimPrefix(ctx.Request.URL.Path, "/"), "/") // убираем первый / и сплитим

	if len(partsURL) != 3 {
		ctx.String(http.StatusNotFound, "Not found")
		return
	}

	metricName := partsURL[2]

	metricVal, exists := h.storage.GetCounter(metricName)

	if !exists {
		ctx.String(http.StatusNotFound, "metric doesn't exist")
		return
	}

	ctx.String(http.StatusOK, strconv.FormatInt(metricVal, 10))

}
