package main

import (
	"fmt"
	handler2 "metrics-service/internal/server/handler"
	storage2 "metrics-service/internal/server/storage"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	// Создаем хранилища
	counterStorage := storage2.NewCounterStorage()
	gaugeStorage := storage2.NewGaugeStorage()

	// Создаем handler's
	counterHandler := handler2.NewCounterHandler(counterStorage)
	gaugeHandler := handler2.NewGaugeHandler(gaugeStorage)

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
