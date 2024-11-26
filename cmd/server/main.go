package main

import (
	"fmt"
	"metrics-service/internal/server/handler"
	"metrics-service/internal/server/storage"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	// Создаем хранилища
	metricsStorage := storage.NewMetricsStorage()

	// Создаем handler's
	counterHandler := handler.NewCounterHandler(metricsStorage)
	gaugeHandler := handler.NewGaugeHandler(metricsStorage)

	mux.HandleFunc("/update/counter/", counterHandler.Update)
	mux.HandleFunc("/update/gauge/", gaugeHandler.Update)

	mux.HandleFunc("/update/", func(w http.ResponseWriter, req *http.Request) {
		http.Error(w, "unsupported metric type", http.StatusBadRequest)
	})

	err := http.ListenAndServe("localhost:8080", mux)
	if err != nil {
		fmt.Printf("Failed to start server: %v", err)
	}
}
