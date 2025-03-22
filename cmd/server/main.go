package main

import (
	"context"
	"errors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"log"
	apperrors "metrics-service/internal/server/errors"
	"metrics-service/internal/server/handler"
	"metrics-service/internal/server/middleware"
	"metrics-service/internal/server/retry"
	"metrics-service/internal/server/storage"
	"syscall"
	"time"
)

func main() {

	// Обрабатываем аргументы командной строки
	parseFlags()

	// Создаем middleware (логгер, gzip)
	mid := middleware.NewMiddleware([]byte(flagHashKey))
	err := mid.InitializeZap(flagLogLevel)
	if err != nil {
		log.Fatalf("Failed to initialize middleware: %v", err)
	}

	// Создаем storage
	storage, err := storage.NewStorage(flagDatabaseDSN, flagFileStoragePath, flagRestore, flagStoreInterval)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	storageRetryer := retry.NewRetryer()
	storageRetryer.SetConditionFunc(func(err error) bool {
		return errors.Is(err, apperrors.ErrPgConnExc) || errors.Is(err, syscall.EBUSY)
	})

	// Подготавливаем бд
	err = storageRetryer.Retry(func() error {
		return storage.Bootstrap(context.Background())
	})
	if err != nil {
		log.Printf("Failed to bootstrap storage: %v", err)
	}
	defer storage.Close()

	isSync := flagStoreInterval <= 0
	// Сохранение данных на диск
	if flagDatabaseDSN == "" && !isSync {
		ticker := time.NewTicker(time.Duration(flagStoreInterval) * time.Second)
		defer ticker.Stop()
		go func() {
			for range ticker.C {
				err = storageRetryer.Retry(func() error {
					return storage.Save()
				})
				if err != nil {
					log.Printf("Failed to save metrics on disk: %v\n", err)
				}
			}
		}()
	}

	// Создаем handler's
	metricsHandler := handler.NewMetricsHandler(storage, storageRetryer, isSync)
	htmlHandler := handler.NewHTMLHandler(storage, storageRetryer)

	server := gin.Default()
	// Роутинг
	// Для всех эндпоинтов используем логирование
	server.Use(mid.WithLogging())

	server.POST("/update/:metricType/:metricName/:metricVal", metricsHandler.Update)
	server.GET("/value/:metricType/:metricName", metricsHandler.Get)
	server.GET("/ping", metricsHandler.Ping)

	// Группа для методов с gzip
	gzipGroup := server.Group("")
	gzipGroup.Use(mid.WithGzip())

	gzipGroup.GET("/", htmlHandler.Get)

	gzipGroup.POST("/update/", metricsHandler.UpdateJSON)
	gzipGroup.POST("/value/", metricsHandler.GetJSON)
	gzipGroup.POST("/updates/", metricsHandler.UpdateBatch)

	pprof.Register(server, "dev/pprof")

	err = server.Run(flagRunAddr)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
