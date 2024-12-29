package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"metrics-service/internal/server/handler"
	"metrics-service/internal/server/middleware"
	"metrics-service/internal/server/service"
	"metrics-service/internal/server/storage"
	"time"
)

func main() {

	// Обрабатываем аргументы командной строки
	parseFlags()

	// Создаем логгер
	mid := middleware.NewMiddleware()
	err := mid.InitializeZap(flagLogLevel)
	if err != nil {
		return
	}

	server := gin.Default()

	// Создаем хранилище
	metricsStorage := storage.NewMetricsStorage()

	// Загружаем storage из файла, если необходимо
	if flagRestore {
		reader, err := storage.NewDiskReader(metricsStorage, flagFileStoragePath)
		if err != nil {
			log.Fatalf("Failed to restore metrics from %v: %v", flagFileStoragePath, err)
		}
		err = reader.Load()
		if err != nil {
			log.Fatalf("Failed to restore metrics from %v: %v", flagFileStoragePath, err)
		}
		err = reader.Close()
		if err != nil {
			log.Fatalf("Failed to close diskReader: %v", err)
		}
	}

	// Создаем diskWriter
	diskW, err := storage.NewDiskWriter(metricsStorage, flagFileStoragePath)
	defer diskW.Save() // штатное сохранение
	if err != nil {
		log.Fatalf("Failed to create disk writer: %v", err)
	}
	isStoreInterval := flagStoreInterval > 0

	// Сохранение данных
	if isStoreInterval {
		ticker := time.NewTicker(time.Duration(flagStoreInterval) * time.Second)
		defer ticker.Stop()
		go func() {
			for range ticker.C {
				err := diskW.Save()
				if err != nil {
					log.Printf("Failed to save metrics: %v", err)
				}
			}
		}()
	}

	// Создаем service'ы

	metricsService := service.NewMetricsService(metricsStorage, diskW)
	htmlService := service.NewHtmlService(metricsStorage)

	// Создаем handler's

	metricsHandler := handler.NewMetricsHandler(metricsService, isStoreInterval)

	htmlHandler := handler.NewHTMLHandler(htmlService)

	// Роутинг

	// Для всех эндпоинтов используем логирование
	server.Use(mid.WithLogging())

	server.POST("/update/:metricType/:metricName/:metricVal", metricsHandler.Update)
	server.GET("/value/:metricType/:metricName", metricsHandler.Get)

	// Группа для методов с gzip
	gzipGroup := server.Group("")
	gzipGroup.Use(mid.WithGzip())

	gzipGroup.GET("/", htmlHandler.Get)

	gzipGroup.POST("/update/", metricsHandler.UpdateJSON)
	gzipGroup.POST("/value/", metricsHandler.GetJSON)

	err = server.Run(flagRunAddr)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

}
