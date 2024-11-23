package handler

import (
	"fmt"
	"metrics-service/internal/storage"
	"net/http"
	"strconv"
	"strings"
)

type MetricsHandler interface {
	Update(w http.ResponseWriter, req *http.Request)
}

type metricsHandler struct {
	metricsStorage storage.MetricsStorage
}

func NewMetricsHandler(metricsStorage storage.MetricsStorage) MetricsHandler {
	return &metricsHandler{metricsStorage}
}

func (h *metricsHandler) Update(w http.ResponseWriter, req *http.Request) {
	// Проверяем http метод
	if req.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Проверяем Content-Type
	if req.Header.Get("Content-Type") != "text/plain" {
		http.Error(w, "unsupported content type", http.StatusUnsupportedMediaType)
		return
	}

	// Парсим url
	/*
		формат: http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
		req.URL.Path возвращает /update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
	*/
	partsUrl := strings.Split(strings.TrimPrefix(req.URL.Path, "/"), "/") // убираем первый / и сплитим

	if len(partsUrl) != 4 {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	metricType := partsUrl[1]
	metricName := partsUrl[2]
	metricValStr := partsUrl[3]

	// Обновляем метрику в зивисимости от типа
	// to-do: вынести в сервис

	switch metricType {
	case "gauge":
		metricVal, err := strconv.ParseFloat(metricValStr, 64)
		if err != nil {
			http.Error(w, "wrong url", http.StatusBadRequest)
			return
		}

		h.metricsStorage.SetGauge(metricName, metricVal)
	case "counter":
		metricVal, err := strconv.ParseInt(metricValStr, 10, 64)
		if err != nil {
			http.Error(w, "wrong url", http.StatusBadRequest)
			return
		}

		h.metricsStorage.SetCounter(metricName, metricVal)
		fmt.Println(h.metricsStorage.GetCounter(metricName))
	default:
		http.Error(w, "wrong metric type", http.StatusBadRequest)

	}

	w.Header().Set("Content-Type", "text/plain")
	//w.Header().Set("Content-Length", )
	w.WriteHeader(http.StatusOK)

}
