package main

import (
	"fmt"
	"metrics-service/internal/handler"
	storage "metrics-service/internal/storage"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	// Создаем экземпляр хранилища
	metricsStorage := storage.NewMetricsStorage()

	// Создаем handler
	metricsHandler := handler.NewMetricsHandler(metricsStorage)

	mux.HandleFunc("/update/", metricsHandler.Update)

	err := http.ListenAndServe("localhost:8080", mux)
	if err != nil {
		fmt.Printf("Failed to start server: %v", err)
	}
}
