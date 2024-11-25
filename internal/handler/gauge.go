package handler

import (
	"metrics-service/internal/storage"
	"net/http"
	"strconv"
	"strings"
)

type gaugeHandler struct {
	gaugeStorage storage.GaugeStorage
}

func NewGaugeHandler(storage storage.GaugeStorage) Handler {
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
	h.gaugeStorage.Set(metricName, metricVal)
	//fmt.Println(h.gaugeStorage.Get(metricName))
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}
