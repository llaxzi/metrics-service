package handler

import (
	"github.com/gin-gonic/gin"
	"metrics-service/internal/server/storage"
	"net/http"
	"strconv"
	"strings"
)

type gaugeHandler struct {
	storage storage.MetricsStorage
}

func NewGaugeHandler(storage storage.MetricsStorage) MetricsHandler {
	return &gaugeHandler{storage}
}

func (h *gaugeHandler) Update(ctx *gin.Context) {
	// Проверяем http метод
	/*if req.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}*/

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

	//metricType := partsURL[1]
	metricName := partsURL[2]
	metricValStr := partsURL[3]

	metricVal, err := strconv.ParseFloat(metricValStr, 64)
	if err != nil {
		ctx.String(http.StatusBadRequest, "wrong URL")
		return
	}

	// TODO вынести в сервис
	h.storage.SetGauge(metricName, metricVal)

	/*metricValue, _ := h.storage.GetGauge(metricName)
	fmt.Printf("Metric: %s, Value: %.0f\n", metricName, metricValue)*/

	ctx.Set("Content-Type", "text/plain")
	ctx.String(http.StatusOK, "updated successfully")
}

func (h *gaugeHandler) Get(ctx *gin.Context) {
	// Проверяем Content-Type
	if ctx.GetHeader("Content-Type") != "text/plain" {

		ctx.String(http.StatusUnsupportedMediaType, "unsupported content type")
		return
	}

	metricName := ctx.Param("")

	metricVal, exists := h.storage.GetGauge(metricName)

	if !exists {
		ctx.String(http.StatusNotFound, "metric doesn't exist")
		return
	}

	ctx.String(http.StatusOK, strconv.FormatFloat(metricVal, 'f', -1, 64))

}
