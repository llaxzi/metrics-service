package handler

import (
	"metrics-service/internal/server/storage"
	"net/http"
	"strconv"
	"strings"
)

type gaugeHandler struct {
	storage storage.MetricsStorage
}

func NewGaugeHandler(storage storage.MetricsStorage) Handler {
	return &gaugeHandler{storage}
}

func (h *gaugeHandler) Update(w http.ResponseWriter, req *http.Request) {
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
	partsURL := strings.Split(strings.TrimPrefix(req.URL.Path, "/"), "/") // убираем первый / и сплитим

	if len(partsURL) != 4 {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	//metricType := partsURL[1]
	metricName := partsURL[2]
	metricValStr := partsURL[3]

	metricVal, err := strconv.ParseFloat(metricValStr, 64)
	if err != nil {
		http.Error(w, "wrong url", http.StatusBadRequest)
		return
	}

	// TODO вынести в сервис
	h.storage.SetGauge(metricName, metricVal)

	/*metricValue, _ := h.storage.GetGauge(metricName)
	fmt.Printf("Metric: %s, Value: %.0f\n", metricName, metricValue)*/

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}
