package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"metrics-service/internal/server/handler"
	"metrics-service/internal/server/storage"
)

func main() {

	server := gin.Default()
	server.LoadHTMLGlob("internal/server/templates/*")

	// Создаем хранилища
	metricsStorage := storage.NewMetricsStorage()

	// Создаем handler's

	updateHandler := handler.NewUpdateHandler(metricsStorage)

	htmlHandler := handler.NewHTMLHandler(metricsStorage)

	// Рутинг
	server.POST("/update/:metricType/:metricName/:metricVal", updateHandler.Update)

	//server.GET("/value/counter/:metricName", counterHandler.Get)

	//server.GET("/value/gauge/:metricName", gaugeHandler.Get)

	server.GET("/", htmlHandler.Get)

	err := server.Run("localhost:8080")
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

}
