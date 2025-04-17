package main

import (
	"context"
	"errors"
	"log"
	"syscall"
	"time"

	"github.com/llaxzi/retryables/v2"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"

	apperrors "metrics-service/internal/server/errors"
	"metrics-service/internal/server/handler"
	"metrics-service/internal/server/middleware"
	storageP "metrics-service/internal/server/storage"
)

func main() {

	printBuildInfo()

	parseFlags()
	parseJSON()
	overrideEnv()

	// Создаем middleware (логгер, gzip)
	mid, err := middleware.NewMiddleware([]byte(flagHashKey), cryptoKeyPath)
	if err != nil {
		log.Fatalf("Failed to initialize middleware: %v", err)
	}

	err = mid.InitializeZap(flagLogLevel)
	if err != nil {
		log.Fatalf("Failed to initialize middleware: %v", err)
	}

	// Создаем storage
	storage, err := storageP.NewStorage(flagDatabaseDSN, flagFileStoragePath, flagRestore, flagStoreInterval)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	storageRetryer := retryables.NewRetryer(nil)
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
	gzipGroup.Use(mid.WithDecryptRSA(), mid.WithGzip())

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
